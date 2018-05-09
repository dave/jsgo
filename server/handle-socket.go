package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"math/rand"

	"html/template"

	"sort"

	"encoding/json"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/messages"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/websocket"
)

func (h *Handler) handleSocketCommand(ctx context.Context, req *http.Request, send func(message messages.Message), receive chan messages.Message, tj *trackerJob) error {
	select {
	case m := <-receive:
		tj.logMessage(m)
		switch m := m.(type) {
		case messages.Compile:
			return h.jsgoCompile(ctx, m, req, send, receive)
		case messages.Update:
			return h.playUpdate(ctx, m, req, send, receive)
		case messages.Share:
			return h.playShare(ctx, m, req, send, receive)
		case messages.Get:
			return h.playGet(ctx, m, req, send, receive)
		case messages.Deploy:
			return h.playDeploy(ctx, m, req, send, receive)
		case messages.Initialise:
			return h.playInitialise(ctx, m, req, send, receive)
		default:
			return fmt.Errorf("invalid init message %T", m)
		}
	case <-time.After(config.WebsocketInstructionTimeout):
		return errors.New("timed out waiting for instruction from client")
	}
}

var infoTmpl = template.Must(template.New("main").Parse(`<html>
	<head>
		<meta charset="utf-8">
		<link href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
        <script src="https://code.jquery.com/jquery-3.2.1.slim.min.js" integrity="sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN" crossorigin="anonymous"></script>
        <script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js" integrity="sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl" crossorigin="anonymous"></script>
	</head>
	<body id="wrapper">
		<h1>
			{{ .Id }}
		</h1>
		<table border="1">
			<tr>
				<th>Start</th>
				<th>Position</th>
				<th>Queue</th>
				<th>Message</th>
				<th>End</th>
				<th>Message</th>
			</tr>
			{{ range .Jobs }}
				<tr>
					<td>
						{{ .SinceStart }}
					</td>
					<td>
						{{ .QueuePos }}
					</td>
					<td>
						{{ .SinceQueue }}
					</td>
					<td>
						{{ .SinceMessage }}
					</td>
					<td>
						{{ .SinceEnd }}
					</td>
					<td>
						<pre>{{ .Message }}</pre>
					</td>
				</tr>
			{{ end }}
		</table>
	</body>
</html>
`))

func (h *Handler) InfoHandler(w http.ResponseWriter, req *http.Request) {
	info := track.info()
	if err := infoTmpl.Execute(w, info); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
	track = &tracker{jobs: make(map[*trackerJob]bool), id: rand.Intn(1000)}
}

var track *tracker

type tracker struct {
	sync.Mutex
	id   int // random number so we can tell the servers apart
	jobs map[*trackerJob]bool
}

type trackerJob struct {
	*tracker
	startTime   time.Time
	queueTime   time.Time
	messageTime time.Time
	endTime     time.Time
	queuePos    int
	message     messages.Message
}

type jobInfo struct {
	SinceStart, SinceQueue, SinceMessage, SinceEnd string
	QueuePos                                       string
	Message                                        interface{}
}

type pageInfo struct {
	Id   string
	Jobs []jobInfo
}

func (t *tracker) start() *trackerJob {
	t.Lock()
	defer t.Unlock()
	tj := &trackerJob{tracker: t, startTime: time.Now()}
	t.jobs[tj] = true
	return tj
}

func (t *tracker) info() pageInfo {
	t.Lock()
	defer t.Unlock()
	var jobs []*trackerJob
	for j := range t.jobs {
		jobs = append(jobs, j)
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].queueTime.UnixNano() < jobs[j].queueTime.UnixNano()
	})
	var info []jobInfo
	for _, j := range jobs {
		b, _ := json.MarshalIndent(j.message, "", "\t")
		ji := jobInfo{Message: fmt.Sprintf("%T: %s", j.message, string(b))}
		if !j.startTime.IsZero() {
			ji.SinceStart = humanize.Time(j.startTime)
		}
		if !j.queueTime.IsZero() {
			ji.SinceQueue = humanize.Time(j.queueTime)
		}
		if !j.messageTime.IsZero() {
			ji.SinceMessage = humanize.Time(j.messageTime)
		}
		if !j.endTime.IsZero() {
			ji.SinceEnd = humanize.Time(j.endTime)
		}
		ji.QueuePos = fmt.Sprint(j.queuePos)
		info = append(info, ji)
	}
	return pageInfo{
		Id:   fmt.Sprint(t.id),
		Jobs: info,
	}
}

func (j *trackerJob) queue(pos int) {
	j.Lock()
	defer j.Unlock()
	j.queuePos = pos
}

func (j *trackerJob) end() {
	j.Lock()
	defer j.Unlock()
	j.endTime = time.Now()
	go func() {
		<-time.After(time.Second * 20)
		j.Lock()
		defer j.Unlock()
		delete(j.jobs, j)
	}()
}

func (j *trackerJob) queueDone() {
	j.Lock()
	defer j.Unlock()
	j.queueTime = time.Now()
}

func (j *trackerJob) logMessage(m messages.Message) {
	j.Lock()
	defer j.Unlock()
	j.messageTime = time.Now()
	j.message = m
}

func (h *Handler) SocketHandler(w http.ResponseWriter, req *http.Request) {

	h.Waitgroup.Add(1)
	defer h.Waitgroup.Done()

	tj := track.start()
	defer tj.end()

	ctx, cancel := context.WithTimeout(req.Context(), config.CompileTimeout)
	defer cancel()

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		storeError(ctx, fmt.Errorf("upgrading request to websocket: %v", err), req)
		return
	}

	var sendWg sync.WaitGroup
	sendChan := make(chan messages.Message, 256)
	receive := make(chan messages.Message, 256)

	send := func(message messages.Message) {
		sendWg.Add(1)
		sendChan <- message
	}

	defer func() {
		// wait for sends to finish before closing websocket
		sendWg.Wait()
		conn.Close()
	}()

	// Recover from any panic and log the error.
	defer func() {
		if r := recover(); r != nil {
			sendAndStoreError(ctx, send, "", errors.New(fmt.Sprintf("panic recovered: %s", r)), req)
		}
	}()

	// Set up a ticker to ping the client regularly
	go func() {
		ticker := time.NewTicker(config.WebsocketPingPeriod)
		defer func() {
			ticker.Stop()
			cancel()
		}()
		for {
			select {
			case message, ok := <-sendChan:
				func() {
					defer sendWg.Done()
					conn.SetWriteDeadline(time.Now().Add(config.WebsocketWriteTimeout))
					if !ok {
						// The send channel was closed.
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
					b, err := messages.Marshal(message)
					if err != nil {
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
					if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
						// Error writing message, close and exit
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
				}()
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(config.WebsocketWriteTimeout))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// React to pongs from the client
	go func() {
		defer func() {
			cancel()
		}()
		conn.SetReadDeadline(time.Now().Add(config.WebsocketPongTimeout))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(config.WebsocketPongTimeout))
			return nil
		})
		for {
			messageType, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					// Don't bother storing an error if the client disconnects gracefully
					break
				}
				if err, ok := err.(*net.OpError); ok && err.Err.Error() == "use of closed network connection" {
					// Don't bother storing an error if the client disconnects gracefully
					break
				}
				storeError(ctx, err, req)
				break
			}
			if messageType == websocket.CloseMessage {
				break
			}
			message, err := messages.Unmarshal(messageBytes)
			if err != nil {
				storeError(ctx, err, req)
				break
			}
			select {
			case receive <- message:
			default:
			}
		}
	}()

	// React to the server shutdown signal
	go func() {
		select {
		case <-h.shutdown:
			sendAndStoreError(ctx, send, "", errors.New("server shut down"), req)
			cancel()
		case <-ctx.Done():
		}
	}()

	// Request a slot in the queue...
	start, end, err := h.Queue.Slot(func(position int) {
		tj.queue(position)
		send(messages.Queueing{Position: position})
	})
	if err != nil {
		sendAndStoreError(ctx, send, "", err, req)
		return
	}

	// Signal to the queue that processing has finished.
	defer close(end)

	// Wait for the slot to become available.
	select {
	case <-start:
		// continue
	case <-ctx.Done():
		return
	}

	tj.queueDone()

	// Send a message to the client that queue step has finished.
	send(messages.Queueing{Done: true})

	if err := h.handleSocketCommand(ctx, req, send, receive, tj); err != nil {
		sendAndStoreError(ctx, send, "", err, req)
		return
	}

	return
}
