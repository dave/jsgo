package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/builder/std"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/server/store"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

func jsgoCompile(ctx context.Context, path string, req *http.Request, send chan messages.Message) {
	// Create a memory filesystem for the getter to store downloaded files (e.g. GOPATH).
	fs := memfs.New()

	// Send a message to the client that downloading step has started.
	send <- messages.Download{Starting: true}

	// Start the download process - just like the "go get" command.
	if err := getter.New(fs, messages.DownloadWriter(send), []string{"jsgo"}).Get(ctx, path, false, false); err != nil {
		sendAndStoreError(ctx, send, path, err, req)
		return
	}

	// Send a message to the client that downloading step has finished.
	send <- messages.Download{Done: true}

	// Start the compile process - this compiles to JS and sends the files to a GCS bucket.
	min, max, err := compile.New(assets.Assets, fs, send).Compile(ctx, path)
	if err != nil {
		sendAndStoreError(ctx, send, path, err, req)
		return
	}

	// Logs the success in the datastore
	storeSuccess(ctx, send, path, req, min, max)

	// Send a message to the client that the process has successfully finished
	send <- messages.Complete{
		Path:    path,
		Short:   strings.TrimPrefix(path, "github.com/"),
		HashMin: fmt.Sprintf("%x", min.Hash),
		HashMax: fmt.Sprintf("%x", max.Hash),
	}
}

func storeSuccess(ctx context.Context, send chan messages.Message, path string, req *http.Request, min, max *compile.CompileOutput) {
	getCompileContents := func(c *compile.CompileOutput, min bool) store.CompileContents {
		val := store.CompileContents{}
		val.Main = fmt.Sprintf("%x", c.Hash)
		var preludeHash string
		if min {
			preludeHash = std.PreludeMin
		} else {
			preludeHash = std.PreludeMax
		}
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

	data := store.CompileData{
		Path:    path,
		Time:    time.Now(),
		Success: true,
		Min:     getCompileContents(min, true),
		Max:     getCompileContents(max, false),
		Ip:      req.Header.Get("X-Forwarded-For"),
	}

	if err := store.Save(ctx, path, data); err != nil {
		// don't save this one to the datastore because it's an error from the datastore.
		sendAndStoreError(ctx, send, path, err, req)
		return
	}

}
