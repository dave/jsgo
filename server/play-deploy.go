package server

import (
	"context"
	"net/http"

	"path/filepath"
	"strings"

	"fmt"

	"time"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/builder/std"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/server/store"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

func playDeploy(ctx context.Context, info messages.Deploy, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	if info.Source[info.Main] == nil {
		return fmt.Errorf("can't find main package %s in source", info.Main)
	}

	// Create a memory filesystem for the getter to store downloaded files (e.g. GOPATH).
	fs := memfs.New()

	source := map[string]bool{}

	for path, files := range info.Source {
		source[path] = true
		if err := storeTemporaryPackage(fs, path, info.Main, files); err != nil {
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
	output, err := compile.New(assets.Assets, fs, send).Compile(ctx, "main", compileWriter{send: send}, true, source)
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

func storeTemporaryPackage(fs billy.Filesystem, path, mainPkg string, source map[string]string) error {
	// Add a dummy package to the filesystem that we can build
	var dir string
	if path == mainPkg {
		dir = filepath.Join("gopath", "src", "main")
	} else {
		dir = filepath.Join("gopath", "src", "main", "vendor", path)
	}
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
