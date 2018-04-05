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

func playUpdate(ctx context.Context, info messages.Update, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	s, err := session.New(info.Source, nil, assets.Assets)
	if err != nil {
		return err
	}

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	// Start the download process - just like the "go get" command.
	g := getter.New(s, downloadWriter{send: send})
	for path := range info.Source {
		if err := g.Get(ctx, path, false, false, false); err != nil {
			return err
		}
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	if err := compile.New(s, send).Update(ctx, info); err != nil {
		return err
	}

	return nil

}
