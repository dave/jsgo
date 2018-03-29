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

func jsgoCompile(ctx context.Context, info messages.Compile, req *http.Request, send func(messages.Message), receive chan messages.Message) error {

	path := info.Path

	// Create a memory filesystem for the getter to store downloaded files (e.g. GOPATH).
	fs := memfs.New()

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	// Start the download process - just like the "go get" command.
	if err := getter.New(fs, downloadWriter{send: send}, []string{"jsgo"}).Get(ctx, path, false, false, false); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	// Start the compile process - this compiles to JS and sends the files to a GCS bucket.
	min, max, err := compile.New(assets.Assets, fs, send).Compile(ctx, path, compileWriter{send: send})
	if err != nil {
		return err
	}

	// Logs the success in the datastore
	storeSuccess(ctx, send, path, req, min, max)

	// Send a message to the client that the process has successfully finished
	send(messages.Complete{
		Path:    path,
		Short:   strings.TrimPrefix(path, "github.com/"),
		HashMin: fmt.Sprintf("%x", min.Hash),
		HashMax: fmt.Sprintf("%x", max.Hash),
	})
	return nil
}

func storeSuccess(ctx context.Context, send func(messages.Message), path string, req *http.Request, min, max *compile.CompileOutput) {
	getCompileContents := func(c *compile.CompileOutput, min bool) store.CompileContents {
		val := store.CompileContents{}
		val.Main = fmt.Sprintf("%x", c.Hash)
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

type compileWriter struct {
	send func(messages.Message)
}

func (w compileWriter) Write(b []byte) (n int, err error) {
	w.send(messages.Compiling{Message: strings.TrimSuffix(string(b), "\n")})
	return len(b), nil
}
