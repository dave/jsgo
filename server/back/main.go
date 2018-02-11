package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/appengine"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/server"

	"github.com/dave/jsgo/compile"
	"github.com/dave/jsgo/getter"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

func main() {
	port := "8080"
	if fromEnv := os.Getenv("PORT"); fromEnv != "" {
		port = fromEnv
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/_ah/health", healthCheckHandler)
	log.Print("Listening on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	if err := server.ServeStatic("favicon.ico", w, req, "image/x-icon"); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		serveCompilePost(w, req)
	} else {
		server.ServeCompile(w, req)
	}
}

type progressWriter struct {
	w http.ResponseWriter
}

func (p progressWriter) Write(b []byte) (n int, err error) {
	i, err := p.w.Write(b)
	if err != nil {
		return i, err
	}
	if f, ok := p.w.(http.Flusher); ok {
		f.Flush()
	}
	return i, nil
}

/*
func compileHandler(ws *websocket.Conn) {
	path := strings.TrimSuffix(strings.TrimPrefix(ws.Request().URL.Path, "/_compile/"), "/")
	if err := compile(path, ws); err != nil {
		fmt.Fprintln(ws, "error", err.Error())
		return
	}
}*/

func serveCompilePost(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
	w.Write([]byte(strings.Repeat(" ", 1024) + "\n"))
	logger := progressWriter{w}
	if err := doCompile(path, logger, req); err != nil {
		fmt.Fprintln(w, "error", err.Error())
		return
	}
}

func doCompile(path string, logger io.Writer, req *http.Request) error {

	fs := memfs.New()

	fmt.Fprintln(logger, "Downloading source...")
	g := getter.New(fs, logger)
	if err := g.Get(path, true, false); err != nil {
		return err
	}

	ctx := appengine.NewContext(req)

	c := compile.New(assets.Assets, fs, logger)
	hashMin, hashMax, err := c.Compile(ctx, path)
	if err != nil {
		return err
	}

	data := server.Data{
		Time:    time.Now(),
		HashMin: hashMin,
		HashMax: hashMax,
	}

	if err := server.Save(ctx, path, data); err != nil {
		return err
	}

	fmt.Fprintln(logger, "\nPage:")
	fmt.Fprintf(logger, "https://jsgo.io/%s (minified)\n", path)
	fmt.Fprintf(logger, "https://jsgo.io/%s$max (non-minified)\n", path)

	fmt.Fprintln(logger, "\nJavascript:")
	fmt.Fprintf(logger, "https://cdn.jsgo.io/pkg/%s.%x.js (minified)\n", path, hashMin)
	fmt.Fprintf(logger, "https://cdn.jsgo.io/pkg/%s.%x.js (non-minified)\n", path, hashMax)

	fmt.Fprintln(logger, "\nCompile link:")
	fmt.Fprintf(logger, "https://compile.jsgo.io/%s\n", path)

	fmt.Fprintln(logger, "\nDone!")

	return nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func hasQuery(req *http.Request, id string) bool {
	_, value := req.URL.Query()[id]
	return value
}
