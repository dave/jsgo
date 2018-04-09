package server

import (
	"context"
	"net/http"

	"fmt"

	"time"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/builder/std"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/gitcache"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/server/store"
	"github.com/dave/jsgo/session"
)

func playDeploy(ctx context.Context, info messages.Deploy, req *http.Request, send func(message messages.Message), receive chan messages.Message, cache *gitcache.Cache) error {

	if info.Source[info.Main] == nil {
		return fmt.Errorf("can't find main package %s in source", info.Main)
	}

	s := session.New(nil, assets.Assets)

	if err := s.SetSource(info.Source); err != nil {
		return err
	}

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	gitreq := cache.NewRequest(false)
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

	// Start the download process - just like the "go get" command.
	if err := getter.New(s, downloadWriter{send: send}, gitreq).Get(ctx, info.Main, false, false, false); err != nil {
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

	if err := storeDeploy(ctx, send, true, req, output[true]); err != nil {
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

func storeDeploy(ctx context.Context, send func(messages.Message), min bool, req *http.Request, output *compile.CompileOutput) error {
	data := store.DeployData{
		Time:     time.Now(),
		Contents: getDeployContents(output, min),
		Minify:   min, // TODO: make this configurable
		Ip:       req.Header.Get("X-Forwarded-For"),
	}
	if err := store.StoreDeploy(ctx, data); err != nil {
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
