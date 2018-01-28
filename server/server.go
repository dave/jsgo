package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/appengine"

	"cloud.google.com/go/datastore"

	pathpkg "path"

	"github.com/dave/jsgo/assets"
	"github.com/pkg/errors"

	"context"

	"github.com/dave/jsgo/compile"
	"github.com/dave/jsgo/getter"
	"github.com/dustin/go-humanize"
	"github.com/shurcooL/httpgzip"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
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
		if req.Method == "POST" {
			serveCompilePost(w, req)
		} else {
			serveCompile(w, req)
		}
	default:
		serveRoot(w, req)
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

func serveJs(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), ".js")
	fmt.Fprintln(w, "js", path)
}

type rootVars struct {
	Path string
	Hash string
	Min  string
}

var rootTpl = template.Must(template.New("root").Parse(`
<html>
	<head>
		<meta charset="utf-8">
		<link rel="icon" type="image/png" href="data:image/png;base64,iVBORw0KGgo=">
	</head>
	<body id="wrapper">
		<span id="log">Loading...</span>
		<script src="https://storage.googleapis.com/jsgo/pkg/{{ .Path }}.{{ .Hash }}{{ .Min }}.js"></script>
	</body>
</html>`))

func serveRoot(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
	max := hasQuery(req, "max")
	found, data, err := lookup(context.Background(), path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if !found {
		http.Redirect(w, req, fmt.Sprintf("/%s?compile", path), http.StatusFound)
		return
	}
	var hash, min string
	if max {
		min = ""
		hash = fmt.Sprintf("%x", data.HashMax)
	} else {
		min = ".min"
		hash = fmt.Sprintf("%x", data.HashMin)
	}
	vars := rootVars{
		Path: path,
		Hash: hash,
		Min:  min,
	}
	if err := rootTpl.Execute(w, vars); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func serveCompile(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")

	found, data, err := lookup(context.Background(), path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type vars struct {
		Found bool
		Path  string
		Last  string
	}

	v := vars{}
	v.Path = path
	if found {
		v.Found = true
		v.Last = humanize.Time(data.Time)
	}

	page := `
		<html>
			<head>
				<meta charset="utf-8">
				<link rel="icon" type="image/png" href="data:image/png;base64,iVBORw0KGgo=">
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

					// Unbuffered HTTP method (doesn't work in App Engine):
					var xhr = new XMLHttpRequest();
					var url = "/{{ .Path }}?compile";
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
	// TODO: https://friendlybit.com/js/partial-xmlhttprequest-responses/
	w.Write([]byte(strings.Repeat(".", 1024) + "\n"))
	logger := progressWriter{w}
	if err := doCompile(path, logger, req); err != nil {
		fmt.Fprintln(w, "error", err.Error())
		return
	}
}

func doCompile(path string, logger io.Writer, req *http.Request) error {

	fs := memfs.New()

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

	if err := save(ctx, path, data); err != nil {
		return err
	}

	fmt.Fprintln(logger, "Done!")

	return nil
}

type Data struct {
	Time    time.Time
	HashMin []byte
	HashMax []byte
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

func lookup(ctx context.Context, path string) (bool, Data, error) {
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

func serveStatic(name string, w http.ResponseWriter, req *http.Request) error {
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
