package server

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"strings"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

func playUpdate(ctx context.Context, info messages.Update, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	if info.Source["main"] == nil {
		return errors.New("can't find main package in source")
	}

	// Create a memory filesystem for the getter to store downloaded files (e.g. GOPATH).
	fs := memfs.New()

	for path, source := range info.Source {
		if err := storeTemporaryPackage(fs, path, source); err != nil {
			return err
		}
	}

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	// Start the download process - just like the "go get" command.
	if err := getter.New(fs, downloadWriter{send: send}, []string{"jsgo"}).Get(ctx, "main", false, false, false); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	if err := compile.New(assets.Assets, fs, send).Update(ctx, info, updateWriter{send: send}); err != nil {
		return err
	}

	return nil

}

type updateWriter struct {
	send func(messages.Message)
}

func (w updateWriter) Write(b []byte) (n int, err error) {
	w.send(messages.Updating{Message: strings.TrimSuffix(string(b), "\n")})
	return len(b), nil
}

func parseForImports(source map[string]string) ([]*ast.File, error) {
	fset := token.NewFileSet()
	var files []*ast.File
	for name, contents := range source {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, contents, parser.ImportsOnly)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}
