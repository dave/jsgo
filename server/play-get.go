package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/gitcache"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/session"
	"github.com/shurcooL/go/ctxhttp"
	"gopkg.in/src-d/go-billy.v4"
)

func playGet(ctx context.Context, info messages.Get, req *http.Request, send func(message messages.Message), receive chan messages.Message, cache *gitcache.Cache) error {
	s := session.New(nil, assets.Assets)
	g := getter.New(s, downloadWriter{send: send}, cache.NewRequest(false))
	_, err := getSource(ctx, g, s, info.Path, send)
	if err != nil {
		return err
	}
	return nil
}

func getSource(ctx context.Context, g *getter.Getter, s *session.Session, path string, send func(message messages.Message)) (map[string]map[string]string, error) {

	if strings.HasPrefix(path, "p/") {
		send(messages.Downloading{Message: path})
		source, err := getGolangPlaygroundSource(ctx, path)
		if err != nil {
			return nil, err
		}
		send(messages.Downloading{Done: true})
		send(messages.GetComplete{Source: source})
		return source, nil
	}

	root := filepath.Join("goroot", "src", path)
	if _, err := assets.Assets.Stat(root); err == nil {
		// Look in the goroot for standard lib packages
		source, err := getSourceFiles(assets.Assets, path, root)
		if err != nil {
			return nil, err
		}
		send(messages.GetComplete{Source: source})
		return source, nil
	}

	// Send a message to the client that downloading step has started.
	send(messages.Downloading{Starting: true})

	// Start the download process - just like the "go get" command.
	// Don't need to give git hints here because only one package will be downloaded
	if err := g.Get(ctx, path, false, false, true); err != nil {
		return nil, err
	}

	source, err := getSourceFiles(s.GoPath(), path, filepath.Join("gopath", "src", path))
	if err != nil {
		return nil, err
	}

	// Send a message to the client that downloading step has finished.
	send(messages.Downloading{Done: true})
	send(messages.GetComplete{Source: source})

	return source, nil
}

func getSourceFiles(fs billy.Filesystem, path, dir string) (map[string]map[string]string, error) {
	source := map[string]map[string]string{}
	fis, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			f.Close()
			return nil, err
		}
		f.Close()
		if source[path] == nil {
			source[path] = map[string]string{}
		}
		source[path][fi.Name()] = string(b)
	}
	return source, nil
}

func getGolangPlaygroundSource(ctx context.Context, path string) (map[string]map[string]string, error) {
	var httpClient = &http.Client{
		Timeout: config.HttpTimeout,
	}
	resp, err := ctxhttp.Get(ctx, httpClient, fmt.Sprintf("https://play.golang.org/%s.go", path))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error %d", resp.StatusCode)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	source := map[string]map[string]string{
		path: {
			"main.go": string(b),
		},
	}
	return source, nil
}
