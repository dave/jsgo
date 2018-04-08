package server

import (
	"context"
	"net/http"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/gitcache"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/session"
)

func playInitialise(ctx context.Context, info messages.Initialise, req *http.Request, send func(message messages.Message), receive chan messages.Message, cache *gitcache.Cache) error {

	s := session.New(nil, assets.Assets)

	gitreq := cache.NewRequest()
	if err := gitreq.Hint(ctx, info.Path); err != nil {
		return err
	}

	source, err := getSource(ctx, s, info.Path, send, gitreq)
	if err != nil {
		return err
	}

	if err := s.SetSource(source); err != nil {
		return err
	}

	// Start the download process - just like the "go get" command.
	if err := getter.New(s, downloadWriter{send: send}, gitreq).Get(ctx, info.Path, false, false, false); err != nil {
		return err
	}

	if err := gitreq.Close(ctx); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	if err := compile.New(s, send).Update(ctx, source, map[string]string{}); err != nil {
		return err
	}

	return nil

}
