package play

import (
	"context"
	"net/http"

	"fmt"

	"time"

	"strings"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/play/messages"
	"github.com/dave/jsgo/server/store"
	"github.com/dave/services"
	"github.com/dave/services/getter/get"
	"github.com/dave/services/session"
)

func (h *Handler) Deploy(ctx context.Context, info messages.Deploy, req *http.Request, send func(message services.Message), receive chan services.Message) error {

	if info.Source[info.Main] == nil {
		return fmt.Errorf("can't find main package %s in source", info.Main)
	}

	s := session.New(info.Tags, assets.Assets, assets.Archives, h.Fileserver, config.ValidExtensions)

	if err := s.SetSource(info.Source); err != nil {
		return err
	}

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	gitreq := h.Cache.NewRequest(false)
	if info.Main == "main" {
		// Using package path "main" as a hint isn't useful... Instead use the imports.
		// TODO: ignore standard library packages in this list.
		if err := gitreq.InitialiseFromHints(ctx, info.Imports...); err != nil {
			return err
		}
	} else {
		if err := gitreq.InitialiseFromHints(ctx, info.Main); err != nil {
			return err
		}
	}

	// set insecure = true in local mode or it will fail if git repo has git protocol
	insecure := config.LOCAL

	// Start the download process - just like the "go get" command.
	if err := get.New(s, downloadWriter{send: send}, gitreq).Get(ctx, info.Main, false, insecure, false); err != nil {
		return err
	}

	if err := gitreq.Close(ctx); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	// Start the compile process - this compiles to JS and sends the files to a GCS bucket.
	output, err := compile.New(s, send).Compile(ctx, info.Main, true)
	if err != nil {
		return err
	}

	if err := h.storeDeploy(ctx, send, true, req, output[true]); err != nil {
		return err
	}

	// Send a message to the client that the process has successfully finished
	// TODO: make minify configurable
	send(messages.DeployComplete{
		Main:  fmt.Sprintf("%x", output[true].MainHash),
		Index: fmt.Sprintf("%x", output[true].IndexHash),
	})

	return nil
}

func (h *Handler) storeDeploy(ctx context.Context, send func(services.Message), min bool, req *http.Request, output *compile.CompileOutput) error {
	data := store.DeployData{
		Time:     time.Now(),
		Contents: getDeployContents(output, min),
		Minify:   min, // TODO: make this configurable
		Ip:       req.Header.Get("X-Forwarded-For"),
	}
	if err := store.StoreDeploy(ctx, h.Database, data); err != nil {
		return err
	}
	return nil
}

func getDeployContents(c *compile.CompileOutput, min bool) store.DeployContents {
	val := store.DeployContents{}
	val.Main = fmt.Sprintf("%x", c.MainHash)
	val.Index = fmt.Sprintf("%x", c.IndexHash)
	preludeHash := std.Prelude[min]
	val.Packages = []store.CompilePackage{
		{
			Path:     "prelude",
			Hash:     preludeHash,
			Standard: true,
		},
	}
	for _, p := range c.Packages {
		val.Packages = append(val.Packages, store.CompilePackage{
			Path:     p.Path,
			Hash:     fmt.Sprintf("%x", p.Hash),
			Standard: p.Standard,
		})
	}
	return val
}

type downloadWriter struct {
	send func(services.Message)
}

func (w downloadWriter) Write(b []byte) (n int, err error) {
	w.send(messages.Downloading{Message: strings.TrimSuffix(string(b), "\n")})
	return len(b), nil
}