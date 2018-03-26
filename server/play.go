package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"cloud.google.com/go/storage"

	"fmt"

	"go/parser"
	"go/token"

	"strconv"

	"path/filepath"

	"go/build"

	"strings"

	"go/ast"

	"bytes"
	"crypto/sha1"
	"encoding/json"
	"io"

	"os"

	"io/ioutil"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"golang.org/x/net/context/ctxhttp"
	"gopkg.in/src-d/go-billy.v4/helper/chroot"
	"gopkg.in/src-d/go-billy.v4/helper/mount"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

func playgroundCompile(ctx context.Context, path string, req *http.Request, send func(messages.Message), receive chan messages.Message) {
	if err := playgroundCompiler(ctx, path, req, send, receive); err != nil {
		sendAndStoreError(ctx, send, path, err, req)
		return
	}
}

func playgroundCompiler(ctx context.Context, path string, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {
	select {
	case m := <-receive:
		switch m := m.(type) {
		case messages.Update:
			return playgroundUpdate(ctx, m, path, req, send, receive)
		case messages.Share:
			return playgroundShare(ctx, m, path, req, send, receive)
		case messages.Get:
			return playgroundGet(ctx, m, path, req, send, receive)
		default:
			return fmt.Errorf("invalid init message %T", m)
		}
	case <-time.After(config.WebsocketInstructionTimeout):
		return errors.New("timed out waiting for instruction from client")
	}
}

func playgroundGet(ctx context.Context, info messages.Get, path string, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	// TODO: fix this
	path = info.Path

	if strings.HasPrefix(path, "p/") {
		// play link
		var httpClient = &http.Client{
			Timeout: config.HttpTimeout,
		}

		send(messages.Downloading{Message: path})

		resp, err := ctxhttp.Get(ctx, httpClient, fmt.Sprintf("https://play.golang.org/%s.go", path))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("error %d", resp.StatusCode)
		}
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		send(messages.Downloading{Done: true})

		files := map[string]map[string]string{
			path: {
				"main.go": string(b),
			},
		}

		send(messages.GetComplete{Source: files})

		return nil
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
		// Start the download process - just like the "go get" command.
		if err := getter.New(fs, downloadWriter{send: send}, []string{"jsgo"}).Get(ctx, path, false, false, true); err != nil {
			return err
		}
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})

	source := map[string]map[string]string{
		path: {},
	}

	dir := filepath.Join("gopath", "src", path)
	fis, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if !strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), ".html") {
			continue
		}
		f, err := fs.Open(filepath.Join(dir, fi.Name()))
		if err != nil {
			return err
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			f.Close()
			return err
		}
		f.Close()
		source[path][fi.Name()] = string(b)
	}

	send(messages.GetComplete{Source: source})

	return nil
}

func playgroundShare(ctx context.Context, info messages.Share, path string, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	send(messages.Storing{Starting: true})

	if config.UseLocal {
		// dummy for local dev
		send(messages.ShareComplete{Hash: "56f9ea337c5f39631fa095e789e44957344e498f"})
		return nil
	}

	buf := &bytes.Buffer{}
	sha := sha1.New()
	w := io.MultiWriter(buf, sha)
	if err := json.NewEncoder(w).Encode(info); err != nil {
		return err
	}
	hash := sha.Sum(nil)

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	storer := compile.NewStorer(ctx, client, send, config.ConcurrentStorageUploads)
	storer.AddSrc("source", fmt.Sprintf("%x.json", hash), buf.Bytes())
	storer.Wait()

	send(messages.Storing{Done: true})

	send(messages.ShareComplete{Hash: fmt.Sprintf("%x", hash)})

	return nil
}

func playgroundUpdate(ctx context.Context, info messages.Update, path string, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {
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
