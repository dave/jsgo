package play

import (
	"context"
	"net/http"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/play/messages"
	"github.com/dave/services"
	"github.com/dave/services/deployer"
	"github.com/dave/services/getter/get"
	"github.com/dave/services/getter/gettermsg"
	"github.com/dave/services/session"
)

func (h *Handler) Initialise(ctx context.Context, info messages.Initialise, req *http.Request, send func(message services.Message), receive chan services.Message) error {

	s := session.New(nil, assets.Assets, assets.Archives, h.Fileserver, config.ValidExtensions)

	gitreq := h.Cache.NewRequest(true)
	if err := gitreq.InitialiseFromHints(ctx, info.Path); err != nil {
		return err
	}
	g := get.New(s, send, gitreq)

	source, err := getSource(ctx, g, s, info.Path, send)
	if err != nil {
		return err
	}

	if err := s.SetSource(source); err != nil {
		return err
	}

	// set insecure = true in local mode or it will fail if git repo has git protocol
	insecure := config.LOCAL

	// Start the download process - just like the "go get" command.
	if err := g.Get(ctx, info.Path, false, insecure, false); err != nil {
		return err
	}

	if err := gitreq.Close(ctx); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(gettermsg.Downloading{Done: true})

	if err := deployer.New(s, send, std.Index, std.Prelude, deployerConfig).Update(ctx, source, map[string]string{}, info.Minify); err != nil {
		return err
	}

	return nil

}
