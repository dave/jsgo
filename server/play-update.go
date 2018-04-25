package server

import (
	"context"
	"net/http"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/builder/session"
	"github.com/dave/jsgo/getter/get"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
)

func (h *Handler) playUpdate(ctx context.Context, info messages.Update, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	s := session.New(nil, assets.Assets)

	if err := s.SetSource(info.Source); err != nil {
		return err
	}

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	gitreq := h.Cache.NewRequest(false)
	var paths []string
	for path := range info.Source {
		paths = append(paths, path)
	}
	if err := gitreq.InitialiseFromHints(ctx, paths...); err != nil {
		return err
	}

	// Start the download process - just like the "go get" command.
	g := get.New(s, downloadWriter{send: send}, gitreq)
	for path := range info.Source {
		if err := g.Get(ctx, path, false, false, false); err != nil {
			return err
		}
	}

	if err := gitreq.Close(ctx); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	if err := compile.New(s, h.Fileserver, send).Update(ctx, info.Source, info.Cache); err != nil {
		return err
	}

	return nil

}
