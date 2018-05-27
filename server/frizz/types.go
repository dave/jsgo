package frizz

import (
	"context"
	"net/http"

	"github.com/dave/jsgo/server/frizz/messages"
	"github.com/dave/services"
)

func (h *Handler) Types(ctx context.Context, info messages.Types, req *http.Request, send func(services.Message), receive chan services.Message) error {
	panic("TODO")
}
