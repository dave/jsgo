package server

import (
	"context"
	"errors"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"gopkg.in/src-d/go-billy.v4/helper/chroot"
	"gopkg.in/src-d/go-billy.v4/helper/mount"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

func playUpdate(ctx context.Context, info messages.Update, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	mainPackageSource, ok := info.Source["main"]
	if !ok {
		return errors.New("can't find main package in source")
	}

	files, err := parseForImports(mainPackageSource)
	if err != nil {
		return err
	}

	// Create a memory filesystem for the getter to store downloaded files (e.g. GOPATH).
	fs := memfs.New()

	if config.UseLocal {
		// KLUDGE JUST FOR TESTING IN LOCAL MODE: "main" dir will be created in gopath/src. Remove it
		// before starting.
		if err := os.RemoveAll(filepath.Join(build.Default.GOPATH, "src", "main")); err != nil {
			return err
		}

		local := osfs.New(filepath.Join(build.Default.GOPATH, "src"))
		mounted := mount.New(fs, filepath.Join("gopath", "src"), local)
		fs = chroot.New(mounted, "/")
	}

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	if !config.UseLocal {
		g := getter.New(fs, downloadWriter{send: send}, []string{"jsgo"})

		imports := map[string]bool{}
		for _, f := range files {
			for _, spec := range f.Imports {
				p, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					return err
				}
				imports[p] = true
			}
		}

		for p := range imports {
			// Start the download process - just like the "go get" command.
			if err := g.Get(ctx, p, false, false, false); err != nil {
				return err
			}
		}
	}

	if err := storeTemporaryPackage(fs, "main", mainPackageSource); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	c := compile.New(assets.Assets, fs, send)

	if err := c.Update(ctx, info, updateWriter{send: send}); err != nil {
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
