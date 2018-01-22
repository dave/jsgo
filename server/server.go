package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/datastore"

	pathpkg "path"

	"github.com/dave/jsgo/assets"
	"github.com/pkg/errors"

	"context"

	"github.com/shurcooL/httpgzip"
)

const PROJECT_ID = "jsgo-192815"

const writeTimeout = time.Second * 2

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/_ah/health", healthCheckHandler)
	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	if err := serveStatic("favicon.ico", w, req); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	switch {
	case strings.HasSuffix(req.URL.Path, ".js"):
		serveJs(w, req)
	case hasQuery(req, "compile"):
		serveCompile(w, req)
	default:
		serveRoot(w, req)
	}
}

func serveJs(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), ".js")
	fmt.Fprintln(w, "js", path)
}

func serveRoot(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")

	if err := save(context.Background(), path, Data{time.Now(), fmt.Sprintf("Bar: %s", path)}); err != nil {
		fmt.Fprintln(w, "error", err.Error())
	}

	fmt.Fprintln(w, "root", path)
}

func serveCompile(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")

	data, err := lookup(context.Background(), path)
	if err != nil {
		fmt.Fprintln(w, "error", err.Error())
	}
	fmt.Fprintln(w, "data", data)

	fmt.Fprintln(w, "compile", path)
}

type Data struct {
	Time time.Time
	Hash string
}

func save(ctx context.Context, path string, data Data) error {
	client, err := datastore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return err
	}
	if _, err := client.Put(ctx, key(path), &data); err != nil {
		return err
	}
	return nil
}

func lookup(ctx context.Context, path string) (Data, error) {
	client, err := datastore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return Data{}, err
	}
	var data Data
	if err := client.Get(ctx, key(path), &data); err != nil {
		return Data{}, err
	}
	return data, nil
}

func key(path string) *datastore.Key {
	return datastore.NameKey("package", path, nil)
}

func serveStatic(name string, w http.ResponseWriter, req *http.Request) error {
	var file http.File
	var err error
	file, err = assets.Assets.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			// Special case: in /static/pkg/ we don't want 404 errors because we can't stop them from
			// popping up in the js console. Instead, deiver a 200 with a zero lenth body.
			if strings.HasPrefix(req.URL.Path, "/static/pkg/") {
				if err := writeWithTimeout(w, []byte{}); err != nil {
					return err
				}
				return nil
			}
			http.NotFound(w, req)
			return nil
		}
		http.Error(w, fmt.Sprintf("error opening %s", name), 500)
		return nil
	}
	defer file.Close()

	w.Header().Set("Cache-Control", "max-age=31536000")
	w.Header().Set("Content-Type", mime.TypeByExtension(pathpkg.Ext(req.URL.Path)))

	_, noCompress := file.(httpgzip.NotWorthGzipCompressing)
	gzb, isGzb := file.(httpgzip.GzipByter)

	if isGzb && !noCompress && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		if err := writeWithTimeout(w, gzb.GzipBytes()); err != nil {
			http.Error(w, fmt.Sprintf("error streaming gzipped %s", name), 500)
			return err
		}
	} else {
		if err := streamWithTimeout(w, file); err != nil {
			http.Error(w, fmt.Sprintf("error streaming %s", name), 500)
			return err
		}
	}
	return nil

}

func streamWithTimeout(w io.Writer, r io.Reader) error {
	c := make(chan error, 1)
	go func() {
		_, err := io.Copy(w, r)
		c <- err
	}()
	select {
	case err := <-c:
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	case <-time.After(writeTimeout):
		return errors.New("timeout")
	}
}

func writeWithTimeout(w io.Writer, b []byte) error {
	return streamWithTimeout(w, bytes.NewBuffer(b))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func hasQuery(req *http.Request, id string) bool {
	_, value := req.URL.Query()[id]
	return value
}
