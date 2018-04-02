package server

import (
	"context"
	"fmt"
	"go/build"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/messages"
	"github.com/shurcooL/go/ctxhttp"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/helper/chroot"
	"gopkg.in/src-d/go-billy.v4/helper/mount"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

func playGet(ctx context.Context, info messages.Get, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	if strings.HasPrefix(info.Path, "p/") {
		// play link
		var httpClient = &http.Client{
			Timeout: config.HttpTimeout,
		}

		send(messages.Downloading{Message: info.Path})

		resp, err := ctxhttp.Get(ctx, httpClient, fmt.Sprintf("https://play.golang.org/%s.go", info.Path))
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
			info.Path: {
				"main.go": string(b),
			},
		}

		send(messages.GetComplete{Source: files})

		return nil
	}

	source := map[string]map[string]string{
		info.Path: {},
	}

	var fs billy.Filesystem
	var dir string

	// Look in the goroot for standard lib packages
	dirRoot := filepath.Join("goroot", "src", info.Path)
	if _, err := assets.Assets.Stat(dirRoot); err == nil {
		fs = assets.Assets
		dir = dirRoot
	} else {

		// Create a memory filesystem for the getter to store downloaded files (e.g. GOPATH).
		fs = memfs.New()

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
			if err := getter.New(fs, downloadWriter{send: send}, []string{"jsgo"}).Get(ctx, info.Path, false, false, true); err != nil {
				return err
			}
		}

		// Send a message to the client that downloading step has finished.
		send(messages.Downloading{Done: true})

		dir = filepath.Join("gopath", "src", info.Path)
	}

	fis, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if !strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), ".html") {
			continue
		}
		if strings.HasSuffix(fi.Name(), "_test.go") {
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
		source[info.Path][fi.Name()] = string(b)
	}

	send(messages.GetComplete{Source: source})

	return nil
}
