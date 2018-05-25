package server

import (
	"context"
	"net/http"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/getter/get"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/services/session"
)

func (h *Handler) playInitialise(ctx context.Context, info messages.Initialise, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	s := session.New(nil, assets.Assets, config.ValidExtensions)

	gitreq := h.Cache.NewRequest(true)
	if err := gitreq.InitialiseFromHints(ctx, info.Path); err != nil {
		return err
	}
	g := get.New(s, downloadWriter{send: send}, gitreq)

	source, err := getSource(ctx, g, s, info.Path, send)
	if err != nil {
		return err
	}

	if err := s.SetSource(source); err != nil {
		return err
	}

	// Start the download process - just like the "go get" command.
	if err := g.Get(ctx, info.Path, false, false, false); err != nil {
		return err
	}

	if err := gitreq.Close(ctx); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	if err := compile.New(s, h.Fileserver, send).Update(ctx, source, map[string]string{}, info.Minify); err != nil {
		return err
	}

	return nil

}
