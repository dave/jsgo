package tracker

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dave/services"
	humanize "github.com/dustin/go-humanize"
)

func Handler(w http.ResponseWriter, req *http.Request) {
	info := Default.info()
	if err := infoTmpl.Execute(w, info); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
	Default = &Tracker{jobs: make(map[*Job]bool), id: rand.Intn(1000)}
}

var Default *Tracker

type Tracker struct {
	sync.Mutex
	id   int // random number so we can tell the servers apart
	jobs map[*Job]bool
}

type Job struct {
	*Tracker
	startTime   time.Time
	queueTime   time.Time
	messageTime time.Time
	endTime     time.Time
	queuePos    int
	message     services.Message
	logs        []string
}

type jobInfo struct {
	SinceStart, SinceQueue, SinceMessage, SinceEnd string
	QueuePos                                       string
	Message                                        interface{}
	Logs                                           string
}

type pageInfo struct {
	Id   string
	Jobs []jobInfo
}

func (t *Tracker) Start() *Job {
	t.Lock()
	defer t.Unlock()
	tj := &Job{Tracker: t, startTime: time.Now()}
	t.jobs[tj] = true
	return tj
}

func (t *Tracker) info() pageInfo {
	t.Lock()
	defer t.Unlock()
	var jobs []*Job
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
		ji.Logs = strings.Join(j.logs, "\n")
		ji.QueuePos = fmt.Sprint(j.queuePos)
		info = append(info, ji)
	}
	return pageInfo{
		Id:   fmt.Sprint(t.id),
		Jobs: info,
	}
}

func (j *Job) Queue(pos int) {
	j.Lock()
	defer j.Unlock()
	j.queuePos = pos
}

func (j *Job) End() {
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

func (j *Job) Log(message string) {
	j.Lock()
	defer j.Unlock()
	j.logs = append(j.logs, message)
}

func (j *Job) QueueDone() {
	j.Lock()
	defer j.Unlock()
	j.queueTime = time.Now()
}

func (j *Job) LogMessage(m services.Message) {
	j.Lock()
	defer j.Unlock()
	j.messageTime = time.Now()
	j.message = m
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
				<th>Logs</th>
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
						<pre>{{ .Logs }}</pre>
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
