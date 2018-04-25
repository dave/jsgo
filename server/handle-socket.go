package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/messages"
	"github.com/gorilla/websocket"
)

func (h *Handler) handleSocketCommand(ctx context.Context, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {
	select {
	case m := <-receive:
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

func (h *Handler) SocketHandler(w http.ResponseWriter, req *http.Request) {

	h.Waitgroup.Add(1)
	defer h.Waitgroup.Done()

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
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					// Don't bother storing an error if the client disconnects gracefully
					break
				}
				storeError(ctx, err, req)
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

	// Send a message to the client that queue step has finished.
	send(messages.Queueing{Done: true})

	if err := h.handleSocketCommand(ctx, req, send, receive); err != nil {
		sendAndStoreError(ctx, send, "", err, req)
		return
	}

	return
}
