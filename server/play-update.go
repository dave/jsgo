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

	fset := token.NewFileSet()
	var files []*ast.File
	for name, contents := range mainPackageSource {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, contents, parser.ImportsOnly)
		if err != nil {
			return err
		}
		files = append(files, f)
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

	// Add a dummy package to the filesystem that we can build
	dir := filepath.Join("gopath", "src", "main")
	if err := fs.MkdirAll(dir, 0777); err != nil {
		return err
	}
	createFile := func(name, contents string) error {
		file, err := fs.Create(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err := file.Write([]byte(contents)); err != nil {
			return err
		}
		return nil
	}
	for name, contents := range mainPackageSource {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if err := createFile(name, contents); err != nil {
			return err
		}
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
