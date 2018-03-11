package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"fmt"

	"go/parser"
	"go/token"

	"strconv"

	"path/filepath"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

func playgroundCompile(ctx context.Context, path string, req *http.Request, send, receive chan messages.Message) {
	if err := playgroundCompiler(ctx, path, req, send, receive); err != nil {
		sendAndStoreError(ctx, send, path, err, req)
		return
	}
}

func playgroundCompiler(ctx context.Context, path string, req *http.Request, send, receive chan messages.Message) error {
	var info messages.PlaygroundCompile
	select {
	case m := <-receive:
		var ok bool
		if info, ok = m.(messages.PlaygroundCompile); !ok {
			return fmt.Errorf("invalid init message %T", m)
		}
	case <-time.After(config.WebsocketInstructionTimeout):
		return errors.New("timed out waiting for instruction from client")
	}

	mainPackageSource, ok := info.Source["main"]
	if !ok {
		return errors.New("can't find main package in source")
	}

	mainSource, ok := mainPackageSource["main.go"]
	if !ok {
		return errors.New("can't find main.go in source")
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "main.go", mainSource, parser.ImportsOnly)
	if err != nil {
		return err
	}

	// Create a memory filesystem for the getter to store downloaded files (e.g. GOPATH).
	fs := memfs.New()

	// Send a message to the client that downloading step has started.
	send <- messages.Download{Starting: true}

	g := getter.New(fs, messages.DownloadWriter(send), []string{"jsgo"})

	var imports []string
	for _, spec := range f.Imports {
		p, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return err
		}
		imports = append(imports, p)
	}

	for _, p := range imports {
		// Start the download process - just like the "go get" command.
		if err := g.Get(ctx, p, false, false); err != nil {
			return err
		}
	}

	// Add a dummy package to the filesystem that we can build
	dir := filepath.Join("gopath", "src", "main")
	if err := fs.MkdirAll(dir, 0777); err != nil {
		return err
	}
	file, err := fs.Create(filepath.Join(dir, "main.go"))
	if err != nil {
		return err
	}
	if _, err := file.Write([]byte(mainSource)); err != nil {
		file.Close()
		return err
	}
	file.Close()

	// Send a message to the client that downloading step has finished.
	send <- messages.Download{Done: true}

	c := compile.New(assets.Assets, fs, send)

	if err := c.Playground(ctx, path); err != nil {
		return err
	}

	return nil

}
