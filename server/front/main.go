package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"context"

	"github.com/dave/jsgo/server"
)

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/_ah/health", healthCheckHandler)
	log.Print("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	if err := server.ServeStatic("favicon.ico", w, req, "image/x-icon"); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	switch {
	case strings.HasSuffix(req.URL.Path, ".js"):
		serveJs(w, req)
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
}

var rootTpl = template.Must(template.New("root").Parse(`
<html>
	<head>
		<meta charset="utf-8">
	</head>
	<body id="wrapper">
		<span id="log">Loading...</span>
		<script src="https://cdn.jsgo.io/pkg/{{ .Path }}.{{ .Hash }}.js"></script>
	</body>
</html>`))

func serveRoot(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
	if path == "" {
		http.Redirect(w, req, "https://github.com/dave/jsgo", http.StatusFound)
	}
	max := hasQuery(req, "max")
	found, data, err := server.Lookup(context.Background(), path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if !found {
		http.Redirect(w, req, fmt.Sprintf("https://compile.jsgo.io/%s", path), http.StatusFound)
		return
	}
	var hash string
	if max {
		hash = fmt.Sprintf("%x", data.HashMax)
	} else {
		hash = fmt.Sprintf("%x", data.HashMin)
	}
	vars := rootVars{
		Path: path,
		Hash: hash,
	}
	if err := rootTpl.Execute(w, vars); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func hasQuery(req *http.Request, id string) bool {
	_, value := req.URL.Query()[id]
	return value
}
