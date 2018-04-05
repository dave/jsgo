package server

import (
	"context"
	"net/http"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/session"
)

func playInitialise(ctx context.Context, info messages.Initialise, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	s := session.New(nil, assets.Assets)

	source, err := getSource(ctx, s, info.Path, send)
	if err != nil {
		return err
	}

	if err := s.SetSource(source); err != nil {
		return err
	}

	// Start the download process - just like the "go get" command.
	g := getter.New(s, downloadWriter{send: send})
	for path := range source {
		if err := g.Get(ctx, path, false, false, false); err != nil {
			return err
		}
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	if err := compile.New(s, send).Update(ctx, source, map[string]string{}); err != nil {
		return err
	}

	return nil

}
