package server

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/appengine"

	"path"

	"errors"

	"cloud.google.com/go/datastore"
	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/compile"
	"github.com/dave/jsgo/getter"
	"github.com/dustin/go-humanize"
	"github.com/shurcooL/httpgzip"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

const PROJECT_ID = "jsgo-192815"

const WriteTimeout = time.Second * 2

/*
func compileHandler(ws *websocket.Conn) {
	path := strings.TrimSuffix(strings.TrimPrefix(ws.Request().URL.Path, "/_compile/"), "/")
	if err := compile(path, ws); err != nil {
		fmt.Fprintln(ws, "error", err.Error())
		return
	}
}*/

func Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		serveCompilePost(w, req)
	} else {
		serveCompilePage(w, req)
	}
}

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

	data := Data{
		Time:    time.Now(),
		HashMin: hashMin,
		HashMax: hashMax,
	}

	if err := Save(ctx, path, data); err != nil {
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

func FaviconHandler(w http.ResponseWriter, req *http.Request) {
	if err := ServeStatic("favicon.ico", w, req, "image/x-icon"); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func serveCompilePage(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")

	if path == "" {
		http.Redirect(w, req, "https://github.com/dave/jsgo", http.StatusFound)
		return
	}

	found, data, err := Lookup(context.Background(), path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type vars struct {
		Found bool
		Path  string
		Last  string
		Root  string
	}

	v := vars{}
	v.Root = CompileRoot(req)
	v.Path = path
	if found {
		v.Found = true
		v.Last = humanize.Time(data.Time)
	}

	page := `
		<html>
			<head>
				<meta charset="utf-8">
			</head>
			<body id="wrapper">
				{{ if .Found }}
					<p>{{ .Path }} was last compiled {{ .Last }}.</p>
				{{ else }}
					<p>{{ .Path }} has never been compiled.</p>
				{{ end }}
				<p>
					<button id="btn">Compile now</button>
				</p>
				<pre id="log"></pre>
			</body>
			<script>
				document.getElementById("btn").onclick = function() {
					document.getElementById("log").innerHTML += "Compiling...\n";

					// Unbuffered HTTP method (doesn't work in App Engine):
					var xhr = new XMLHttpRequest();
					var url = "{{ .Root }}{{ .Path }}";
					xhr.open("POST", url, true);
					xhr.send();
					var last_index = 0;
					function parse() {
						var curr_index = xhr.responseText.length;
						if (last_index == curr_index) return; // No new data
						var s = xhr.responseText.substring(last_index, curr_index);
						last_index = curr_index;
						document.getElementById("log").innerHTML += s;
					}
					// Check for new content every 100ms
					var interval = setInterval(parse, 100);

					// WebSocket method (also doesn't work in App Engine):
					/*
					var socket = new WebSocket("ws://localhost:8080/_compile/{{ .Path }}");

					socket.onopen = function() {
						document.getElementById("log").innerHTML += "Socket opened\n";
					};
					socket.onmessage = function (e) {
						document.getElementById("log").innerHTML += e.data;
					}
					socket.onclose = function () {
						document.getElementById("log").innerHTML += "Socket closed\n";
					}
					*/
				};
			</script>
		</html>`

	tmpl, err := template.New("test").Parse(page)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err := tmpl.Execute(w, v); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func CompileRoot(req *http.Request) string {
	if req.Host == "jsgo.io" {
		return "https://compile.jsgo.io/"
	}
	return "/"
}

type Data struct {
	Time    time.Time
	HashMin []byte
	HashMax []byte
}

func Save(ctx context.Context, path string, data Data) error {
	client, err := datastore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return err
	}
	if _, err := client.Put(ctx, key(path), &data); err != nil {
		return err
	}
	return nil
}

func Lookup(ctx context.Context, path string) (bool, Data, error) {
	client, err := datastore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return false, Data{}, err
	}
	var data Data
	if err := client.Get(ctx, key(path), &data); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return false, Data{}, nil
		}
		return false, Data{}, err
	}
	return true, data, nil
}

func key(path string) *datastore.Key {
	return datastore.NameKey("package", path, nil)
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

	w.Header().Set("Cache-Control", "max-age=31536000")
	if mimeType == "" {
		w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(req.URL.Path)))
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
	case <-time.After(WriteTimeout):
		return errors.New("timeout")
	}
}

func WriteWithTimeout(w io.Writer, b []byte) error {
	return StreamWithTimeout(w, bytes.NewBuffer(b))
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
