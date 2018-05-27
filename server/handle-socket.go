package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dave/jsgo/server/servermsg"
	"github.com/dave/services"
	"github.com/dave/services/tracker"
	"github.com/gorilla/websocket"
)

type SocketHandlerInterface interface {
	Handle(ctx context.Context, req *http.Request, send func(message services.Message), receive chan services.Message, tj *tracker.Job) error
	RequestTimeout() time.Duration
	WebsocketPingPeriod() time.Duration
	WebsocketTimeout() time.Duration
	WebsocketPongTimeout() time.Duration
	MarshalMessage(services.Message) (payload []byte, messageType int, err error)
	UnarshalMessage([]byte) (services.Message, error)
	StoreError(ctx context.Context, err error, req *http.Request)
}

func (h *Handler) SocketHandler(s SocketHandlerInterface) func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {

		h.Waitgroup.Add(1)
		defer func() {
			h.Waitgroup.Done()
		}()

		tj := tracker.Default.Start()
		defer func() {
			tj.End()
		}()

		ctx, cancel := context.WithTimeout(req.Context(), s.RequestTimeout())
		defer func() {
			cancel()
		}()

		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			h.storeError(ctx, fmt.Errorf("upgrading request to websocket: %v", err), req)
			return
		}

		var sendWg sync.WaitGroup
		sendCh := make(chan services.Message, 256)
		receive := make(chan services.Message, 256)
		var finished bool

		send := func(message services.Message) {
			if finished {
				return // prevent more messages from being sent after we want to finish
			}
			sendWg.Add(1)
			sendCh <- message
		}

		defer func() {
			finished = true // we won't be adding any more messages to the send channel
			sendWg.Wait()   // wait for in-flight sends to finish
			close(sendCh)   // close the sendChan, so the send loop will exit
			conn.Close()    // finally close the websocket
		}()

		// Recover from any panic and log the error.
		defer func() {
			if r := recover(); r != nil {
				s.StoreError(ctx, fmt.Errorf("panic recovered: %s", r), req)
				send(servermsg.Error{Message: fmt.Sprintf("panic recovered: %s", r)})
			}
		}()

		// Set up a ticker to ping the client regularly
		go func() {
			ticker := time.NewTicker(s.WebsocketPingPeriod())
			defer func() {
				ticker.Stop()
				cancel()
			}()
			for {
				select {
				case message, ok := <-sendCh:
					if !ok {
						// the send channel was closed - exit immediately
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
					func() {
						defer sendWg.Done()
						b, messageType, err := s.MarshalMessage(message)
						if err != nil {
							return
						}
						conn.SetWriteDeadline(time.Now().Add(s.WebsocketTimeout()))
						conn.WriteMessage(messageType, b)
					}()
				case <-ticker.C:
					conn.SetWriteDeadline(time.Now().Add(s.WebsocketTimeout()))
					conn.WriteMessage(websocket.PingMessage, nil)
				}
			}
		}()

		// React to pongs from the client
		go func() {
			defer func() {
				cancel()
			}()
			conn.SetReadDeadline(time.Now().Add(s.WebsocketPongTimeout()))
			conn.SetPongHandler(func(string) error {
				conn.SetReadDeadline(time.Now().Add(s.WebsocketPongTimeout()))
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
					h.storeError(ctx, err, req)
					break
				}
				if messageType == websocket.CloseMessage {
					break
				}
				message, err := s.UnarshalMessage(messageBytes)
				if err != nil {
					h.storeError(ctx, err, req)
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
				s.StoreError(ctx, errors.New("server shut down"), req)
				send(servermsg.Error{Message: "server shut down"})
				cancel()
			case <-ctx.Done():
			}
		}()

		// Request a slot in the queue...
		start, end, err := h.Queue.Slot(func(position int) {
			tj.Queue(position)
			send(servermsg.Queueing{Position: position})
		})
		if err != nil {
			s.StoreError(ctx, err, req)
			send(servermsg.Error{Message: err.Error()})
			return
		}

		// Signal to the queue that processing has finished.
		defer func() {
			close(end)
		}()

		// Wait for the slot to become available.
		select {
		case <-start:
			// continue
		case <-ctx.Done():
			return
		}

		tj.QueueDone()

		// Send a message to the client that queue step has finished.
		send(servermsg.Queueing{Done: true})

		if err := s.Handle(ctx, req, send, receive, tj); err != nil {
			s.StoreError(ctx, err, req)
			send(servermsg.Error{Message: err.Error()})
			return
		}

		return
	}
}
