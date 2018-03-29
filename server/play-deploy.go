package server

import (
	"context"
	"errors"
	"net/http"

	"path/filepath"
	"strings"

	"fmt"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

func playDeploy(ctx context.Context, info messages.Deploy, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

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

	// Start the compile process - this compiles to JS and sends the files to a GCS bucket.
	output, err := compile.New(assets.Assets, fs, send).Compile(ctx, "main", compileWriter{send: send}, true)
	if err != nil {
		return err
	}

	// TODO: store in database

	// Send a message to the client that the process has successfully finished
	// TODO: make minify configurable
	send(messages.DeployComplete{
		Main:  fmt.Sprintf("%x", output[true].MainHash),
		Index: fmt.Sprintf("%x", output[true].IndexHash),
	})

	return nil
}

func storeTemporaryPackage(fs billy.Filesystem, path string, source map[string]string) error {
	// Add a dummy package to the filesystem that we can build
	dir := filepath.Join("gopath", "src", path)
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
	for name, contents := range source {
		if !strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, ".inc.js") && !strings.HasSuffix(name, ".jsgo.html") {
			continue
		}
		if err := createFile(name, contents); err != nil {
			return err
		}
	}
	return nil
}
