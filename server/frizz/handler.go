package frizz

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"

	"time"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/frizz/messages"
	"github.com/dave/jsgo/server/store"
	"github.com/dave/services"
	"github.com/dave/services/getter/cache"
	"github.com/dave/services/queue"
	"github.com/dave/services/tracker"
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
		case messages.Types:
			return h.Types(ctx, m, req, send, receive)
		default:
			return fmt.Errorf("invalid init message %T", m)
		}
	case <-time.After(config.WebsocketInstructionTimeout):
		tj.Log("timeout")
		return errors.New("timed out waiting for instruction from client")
	}
}

func (h *Handler) RequestTimeout() time.Duration {
	panic("implement me")
}

func (h *Handler) WebsocketPingPeriod() time.Duration {
	panic("implement me")
}

func (h *Handler) WebsocketTimeout() time.Duration {
	panic("implement me")
}

func (h *Handler) WebsocketPongTimeout() time.Duration {
	panic("implement me")
}

func (h *Handler) MarshalMessage(services.Message) (payload []byte, messageType int, err error) {
	panic("implement me")
}

func (h *Handler) UnarshalMessage([]byte) (services.Message, error) {
	panic("implement me")
}

func (h *Handler) SendQueueing(send func(message services.Message), position int, done bool) {
	panic("implement me")
}

func (h *Handler) SendError(send func(message services.Message), err error) {
	panic("implement me")
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
