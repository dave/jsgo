package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/builder/session"
	"github.com/dave/jsgo/builder/std"
	"github.com/dave/jsgo/getter/get"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/server/store"
)

func (h *Handler) jsgoCompile(ctx context.Context, info messages.Compile, req *http.Request, send func(messages.Message), receive chan messages.Message) error {

	path := info.Path

	s := session.New(nil, assets.Assets)

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	gitreq := h.Cache.NewRequest(true)
	if err := gitreq.InitialiseFromHints(ctx, path); err != nil {
		return err
	}

	// Start the download process - just like the "go get" command.
	if err := get.New(s, downloadWriter{send: send}, gitreq).Get(ctx, path, false, false, false); err != nil {
		return err
	}

	if err := gitreq.Close(ctx); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	// Start the compile process - this compiles to JS and sends the files to a GCS bucket.
	output, err := compile.New(s, h.Fileserver, send).Compile(ctx, path, false)
	if err != nil {
		return err
	}

	// Logs the success in the datastore
	storeCompile(ctx, send, path, req, output)

	// Send a message to the client that the process has successfully finished
	send(messages.Complete{
		Path:    path,
		Short:   strings.TrimPrefix(path, "github.com/"),
		HashMin: fmt.Sprintf("%x", output[true].MainHash),
		HashMax: fmt.Sprintf("%x", output[false].MainHash),
	})
	return nil
}

func storeCompile(ctx context.Context, send func(messages.Message), path string, req *http.Request, output map[bool]*compile.CompileOutput) {
	data := store.CompileData{
		Path:    path,
		Time:    time.Now(),
		Min:     getCompileContents(output[true], true),
		Max:     getCompileContents(output[false], false),
		Ip:      req.Header.Get("X-Forwarded-For"),
		Success: true,
	}
	if err := store.StoreCompile(ctx, path, data); err != nil {
		// don't save this one to the datastore because it's an error from the datastore.
		sendError(send, err)
		return
	}
}

func getCompileContents(c *compile.CompileOutput, min bool) store.CompileContents {
	val := store.CompileContents{}
	val.Main = fmt.Sprintf("%x", c.MainHash)
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
