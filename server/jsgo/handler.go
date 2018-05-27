package jsgo

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"

	"time"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/jsgo/messages"
	"github.com/dave/jsgo/server/store"
	"github.com/dave/jsgo/server/tracker"
	"github.com/dave/services"
	"github.com/dave/services/getter/cache"
	"github.com/dave/services/queue"
)

type Handler struct {
	Cache      *cache.Cache
	Fileserver services.Fileserver
	Database   services.Database
}

func (h *Handler) Handle(ctx context.Context, req *http.Request, send func(message services.Message), receive chan services.Message, tj *tracker.Job) error {
	select {
	case m := <-receive:
		tj.LogMessage(m)
		switch m := m.(type) {
		case messages.Compile:
			return h.Compile(ctx, m, req, send, receive)
		default:
			return fmt.Errorf("invalid init message %T", m)
		}
	case <-time.After(config.WebsocketInstructionTimeout):
		tj.Log("timeout")
		return errors.New("timed out waiting for instruction from client")
	}
}

func (h *Handler) RequestTimeout() time.Duration {
	return config.RequestTimeout
}

func (h *Handler) WebsocketPingPeriod() time.Duration {
	return config.WebsocketPingPeriod
}

func (h *Handler) WebsocketTimeout() time.Duration {
	return config.WebsocketWriteTimeout
}

func (h *Handler) WebsocketPongTimeout() time.Duration {
	return config.WebsocketPongTimeout
}

func (h *Handler) MarshalMessage(m services.Message) (payload []byte, messageType int, err error) {
	return messages.Marshal(m)
}

func (h *Handler) UnarshalMessage(b []byte) (services.Message, error) {
	return messages.Unmarshal(b)
}

func (h *Handler) SendQueueing(send func(message services.Message), position int, done bool) {
	send(messages.Queueing{Position: position, Done: done})
}

func (h *Handler) SendError(send func(message services.Message), err error) {
	send(messages.Error{Message: err.Error()})
}

func (h *Handler) StoreError(ctx context.Context, err error, req *http.Request) {

	fmt.Println(err)

	if err == queue.TooManyItemsQueued {
		// If the server is getting flooded by a DOS, this will prevent database flooding
		return
	}

	h.Database.Put(ctx, datastore.IncompleteKey(config.ErrorKind, nil), &store.Error{
		Time:  time.Now(),
		Error: err.Error(),
		Ip:    req.Header.Get("X-Forwarded-For"),
	})

}