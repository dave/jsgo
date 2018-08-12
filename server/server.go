package server

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	pathpkg "path"

	"errors"

	"context"

	"sync"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/frizz"
	"github.com/dave/jsgo/server/jsgo"
	"github.com/dave/jsgo/server/play"
	"github.com/dave/jsgo/server/store"
	"github.com/dave/jsgo/server/wasm"
	"github.com/dave/patsy"
	"github.com/dave/patsy/vos"
	"github.com/dave/services"
	"github.com/dave/services/database/gcsdatabase"
	"github.com/dave/services/database/localdatabase"
	"github.com/dave/services/fetcher/gitfetcher"
	"github.com/dave/services/fetcher/localfetcher"
	"github.com/dave/services/fileserver/cachefileserver"
	"github.com/dave/services/fileserver/gcsfileserver"
	"github.com/dave/services/fileserver/localfileserver"
	"github.com/dave/services/getter/cache"
	"github.com/dave/services/queue"
	"github.com/dave/services/tracker"
	"github.com/gorilla/websocket"
	"github.com/shurcooL/httpgzip"
	"gopkg.in/src-d/go-billy.v4"
)

func init() {
	assets.Init()
}

func New(shutdown chan struct{}) *Handler {
	var c *cache.Cache
	var fileserver services.Fileserver
	var database services.Database
	if config.LOCAL {
		fileserver = localfileserver.New(config.LocalFileserverTempDir, config.Static, config.Host, config.Bucket)
		database = localdatabase.New(config.LocalFileserverTempDir)
		fetcherResolver := localfetcher.New()
		c = cache.New(
			database,
			fetcherResolver,
			fetcherResolver,
			config.HintsKind,
		)
	} else {
		storageClient, err := storage.NewClient(context.Background())
		if err != nil {
			panic(err)
		}

		datastoreClient, err := datastore.NewClient(context.Background(), config.ProjectID)
		if err != nil {
			panic(err)
		}

		database = gcsdatabase.New(datastoreClient)
		fileserver = gcsfileserver.New(storageClient, config.Buckets)
		c = cache.New(
			database,
			gitfetcher.New(
				cachefileserver.New(1024*1024*1042, 100*1024*1024),
				fileserver,
				config.GitFetcherConfig,
			),
			nil,
			config.HintsKind,
		)
	}
	h := &Handler{
		mux:        http.NewServeMux(),
		shutdown:   shutdown,
		Queue:      queue.New(config.MaxConcurrentCompiles, config.MaxQueue),
		Waitgroup:  &sync.WaitGroup{},
		Cache:      c,
		Fileserver: fileserver,
		Database:   database,
	}
	h.mux.HandleFunc("/", h.PageHandler)
	h.mux.HandleFunc("/_script.js", h.ScriptHandler)
	h.mux.HandleFunc("/_script.js.map", h.ScriptHandler)
	h.mux.HandleFunc("/_info/", tracker.Handler)

	h.mux.HandleFunc("/_jsgo/", h.SocketHandler(&jsgo.Handler{h.Cache, h.Fileserver, h.Database}))
	h.mux.HandleFunc("/_play/", h.SocketHandler(&play.Handler{h.Cache, h.Fileserver, h.Database}))
	h.mux.HandleFunc("/_frizz/", h.SocketHandler(&frizz.Handler{h.Cache, h.Fileserver, h.Database}))
	h.mux.HandleFunc("/_wasm/", h.SocketHandler(&wasm.Handler{h.Cache, h.Fileserver, h.Database}))

	//h.mux.HandleFunc("/_ws/", h.SocketHandler)
	//h.mux.HandleFunc("/_pg/", h.SocketHandler)
	h.mux.HandleFunc("/favicon.ico", h.IconHandler)
	h.mux.HandleFunc("/compile.css", h.CssHandler)
	h.mux.HandleFunc("/_ah/health", h.HealthCheckHandler)
	if config.LOCAL {
		dir, err := patsy.Dir(vos.Os(), "github.com/dave/jsgo/assets/static/")
		if err != nil {
			panic(err)
		}
		h.mux.Handle("/_local/", http.FileServer(http.Dir(dir)))
	}
	return h
}

type Handler struct {
	Cache      *cache.Cache
	Fileserver services.Fileserver
	Database   services.Database
	Waitgroup  *sync.WaitGroup
	Queue      *queue.Queue
	mux        *http.ServeMux
	shutdown   chan struct{}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *Handler) storeError(ctx context.Context, err error, req *http.Request) {

	if err == queue.TooManyItemsQueued {
		// If the server is getting flooded by a DOS, this will prevent database flooding
		return
	}

	// ignore errors when logging an error
	store.StoreError(ctx, h.Database, store.Error{
		Time:  time.Now(),
		Error: err.Error(),
		Ip:    req.Header.Get("X-Forwarded-For"),
	})

}

func (h *Handler) IconHandler(w http.ResponseWriter, req *http.Request) {
	if err := ServeStatic(req.URL.Path, w, req, "image/x-icon"); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func (h *Handler) CssHandler(w http.ResponseWriter, req *http.Request) {
	if err := ServeStatic(req.URL.Path, w, req, "text/css"); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func ServeStatic(name string, w http.ResponseWriter, req *http.Request, mimeType string) error {
	var file billy.File
	var err error
	file, err = assets.Assets.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, req)
			return nil
		}
		http.Error(w, fmt.Sprintf("error opening %s", name), 500)
		return nil
	}
	defer file.Close()

	w.Header().Set("Cache-Control", "public,max-age=31536000,immutable")
	if mimeType == "" {
		w.Header().Set("Content-Type", mime.TypeByExtension(pathpkg.Ext(req.URL.Path)))
	} else {
		w.Header().Set("Content-Type", mimeType)
	}

	_, noCompress := file.(httpgzip.NotWorthGzipCompressing)
	gzb, isGzb := file.(httpgzip.GzipByter)

	if isGzb && !noCompress && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		if err := WriteWithTimeout(w, gzb.GzipBytes()); err != nil {
			http.Error(w, fmt.Sprintf("error streaming gzipped %s", name), 500)
			return err
		}
	} else {
		if err := StreamWithTimeout(w, file); err != nil {
			http.Error(w, fmt.Sprintf("error streaming %s", name), 500)
			return err
		}
	}
	return nil

}

func StreamWithTimeout(w io.Writer, r io.Reader) error {
	c := make(chan error, 1)
	go func() {
		_, err := io.Copy(w, r)
		c <- err
	}()
	select {
	case err := <-c:
		if err != nil {
			return err
		}
		return nil
	case <-time.After(config.WriteTimeout):
		return errors.New("timeout")
	}
}

func WriteWithTimeout(w io.Writer, b []byte) error {
	return StreamWithTimeout(w, bytes.NewBuffer(b))
}

type Pather interface {
	Path() string
}
