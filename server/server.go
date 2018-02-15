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

	pathpkg "path"

	"errors"

	"regexp"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/builder"
	"github.com/dave/jsgo/builder/std"
	"github.com/dave/jsgo/compile"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/logger"
	"github.com/dave/jsgo/server/queue"
	"github.com/dave/jsgo/server/store"
	"github.com/dustin/go-humanize"
	"github.com/shurcooL/httpgzip"
	"golang.org/x/net/websocket"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

var queuer = queue.New(config.MaxConcurrentCompiles, config.MaxQueue)

func SocketHandler(ws *websocket.Conn) {
	//ctx := context.Background()

	//ctx, cancel := context.WithTimeout(ctx)

	path := strings.TrimSuffix(strings.TrimPrefix(ws.Request().URL.Path, "/_ws/"), "/")

	path = normalizePath(path)

	log := logger.New(ws)

	/*
		defer func() {
			if r := recover(); r != nil {
				// TODO: Write error to database here
				log.Log(logger.Error, logger.ErrorPayload{
					Path:    path,
					Message: fmt.Sprintf("Panic recovered: %s", r),
				})
			}
		}()
	*/

	start, end, err := queuer.Slot(func(position int) { log.Log(logger.Queue, logger.QueuePayload{Position: position}) })
	if err != nil {
		log.Log(logger.Error, logger.ErrorPayload{
			Path:    path,
			Message: err.Error(),
		})
		return
	}
	defer close(end)
	<-start
	log.Log(logger.Queue, logger.QueuePayload{Done: true})

	if err := doSocketCompile(ws, path, log); err != nil {
		log.Log(logger.Error, logger.ErrorPayload{
			Path:    path,
			Message: err.Error(),
		})
		return
	}
}

func doSocketCompile(ws *websocket.Conn, path string, log *logger.Logger) error {

	fs := memfs.New()

	ctx := context.Background()

	log.Log(logger.Download, logger.DownloadingPayload{Done: false})
	downloadLogger := funcWriter{func(b []byte) error {
		return log.Log(logger.Download, logger.DownloadingPayload{
			Path: string(b),
		})
	}}
	g := getter.New(fs, downloadLogger)
	if err := g.Get(path, false, false); err != nil {
		data := store.CompileData{
			Path:    path,
			Time:    time.Now(),
			Success: false,
			Error:   err.Error(),
			Ip:      ws.Request().Header.Get("X-Forwarded-For"),
		}
		store.Save(ctx, path, data) // ignore error
		return err
	}
	log.Log(logger.Download, logger.DownloadingPayload{Done: true})

	c := compile.New(assets.Assets, fs, log)
	hashMin, hashMax, outputMin, outputMax, err := c.Compile(ctx, path)
	if err != nil {
		data := store.CompileData{
			Path:    path,
			Time:    time.Now(),
			Success: false,
			Error:   err.Error(),
			Ip:      ws.Request().Header.Get("X-Forwarded-For"),
		}
		store.Save(ctx, path, data) // ignore error
		return err
	}

	getCc := func(hash []byte, output *builder.CommandOutput) store.CompileContents {
		val := store.CompileContents{}
		val.Main = fmt.Sprintf("%x", hash)
		val.Prelude = std.PreludeHash
		for _, p := range output.Packages {
			val.Packages = append(val.Packages, store.CompilePackage{
				Path:     p.Path,
				Standard: p.Standard,
				Hash:     fmt.Sprintf("%x", p.Hash),
			})
		}
		return val
	}

	data := store.CompileData{
		Path:    path,
		Time:    time.Now(),
		Success: true,
		Min:     getCc(hashMin, outputMin),
		Max:     getCc(hashMax, outputMax),
		Ip:      ws.Request().Header.Get("X-Forwarded-For"),
	}

	if err := store.Save(ctx, path, data); err != nil {
		return err
	}

	log.Log(logger.Complete, logger.CompletePayload{
		Path:    path,
		Short:   strings.TrimPrefix(path, "github.com/"),
		HashMin: fmt.Sprintf("%x", hashMin),
		HashMax: fmt.Sprintf("%x", hashMax),
	})

	return nil
}

func Handler(w http.ResponseWriter, req *http.Request) {
	serveCompilePage(w, req)
}

func IconHandler(w http.ResponseWriter, req *http.Request) {
	if err := ServeStatic(req.URL.Path, w, req, "image/x-icon"); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func CssHandler(w http.ResponseWriter, req *http.Request) {
	if err := ServeStatic(req.URL.Path, w, req, "text/css"); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func normalizePath(path string) string {

	// We should normalize gist urls by removing the username part
	if strings.HasPrefix(path, "gist.github.com/") {
		matches := gistWithUsername.FindStringSubmatch(path)
		if len(matches) > 1 {
			return fmt.Sprintf("gist.github.com/%s", matches[1])
		}
	}

	// Add github.com if the first part of the path is not a hostname and matches the github username regex
	if strings.Contains(path, "/") {
		firstPart := path[:strings.Index(path, "/")]
		if !strings.Contains(firstPart, ".") && githubUsername.MatchString(firstPart) {
			return fmt.Sprintf("github.com/%s", path)
		}
	}

	return path
}

var gistWithUsername = regexp.MustCompile(`^gist\.github\.com/[A-Za-z0-9_.\-]+/([a-f0-9]+)(/[\p{L}0-9_.\-]+)*$`)
var githubUsername = regexp.MustCompile(`^[a-zA-Z0-9\-]{0,38}$`)

func serveCompilePage(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")

	path = normalizePath(path)

	if path == "" {
		http.Redirect(w, req, "https://github.com/dave/jsgo", http.StatusFound)
		return
	}

	ctx := appengine.NewContext(req)
	found, data, err := store.Lookup(ctx, path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type vars struct {
		Found  bool
		Path   string
		Last   string
		Host   string
		Scheme string
	}

	v := vars{}
	v.Host = req.Host
	v.Path = path
	if req.Host == "compile.jsgo.io" {
		v.Scheme = "wss"
	} else {
		v.Scheme = "ws"
	}
	if found {
		v.Found = true
		v.Last = humanize.Time(data.Time)
	}

	page := `
		<html>
			<head>
				<meta charset="utf-8">
				<link href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
				<link href="/compile.css" rel="stylesheet">
			</head>
			<body>
			    <div class="site-wrapper">
					<div class="site-wrapper-inner">
						<div class="cover-container">
							<div class="masthead clearfix">
								<div class="inner">
									<h3 class="masthead-brand">jsgo</h3>
									<nav class="nav nav-masthead">
										<a class="nav-link active" href="">Compile</a>
										<a class="nav-link" href="https://github.com/dave/jsgo">Info</a>
									</nav>
								</div>
							</div>

							<div id="header-panel" class="inner cover">
								<h1 class="cover-heading">Compile</h1>
								<p class="lead">
									{{ .Path }}
									{{ if .Found }} was compiled {{ .Last }} {{ end }}
								</p>
								<p class="lead" id="button-panel">
									<a href="#" class="btn btn-lg btn-secondary" id="btn">Compile</a>
								</p>
							</div>

							<div id="progress-panel" style="display: none;">
								<table class="table table-dark">
									<tbody>
										<tr id="queue-item" style="display: none;">
											<th scope="row" class="w-25">Queued:</th>
											<td class="w-75"><span id="queue-span"></span></td>
										</tr>
										<tr id="download-item" style="display: none;">
											<th scope="row" class="w-25">Downloading:</th>
											<td class="w-75"><span id="download-span"></span></td>
										</tr>
										<tr id="compile-item" style="display: none;">
											<th scope="row" class="w-25">Compiling:</th>
											<td class="w-75"><span id="compile-span"></span></td>
										</tr>
										<tr id="store-item" style="display: none;">
											<th scope="row" class="w-25">Storing:</th>
											<td class="w-75"><span id="store-span"></span></td>
										</tr>
										<tr id="index-item" style="display: none;">
											<th scope="row" class="w-25">Index:</th>
											<td class="w-75"><span id="index-span"></span></td>
										</tr>
									</tbody>
								</table>
							</div>
							<div id="error-panel" style="display: none;" class="alert alert-warning" role="alert">
								<h4 class="alert-heading">Error</h4>
								<pre id="error-message"></pre>
							</div>
							<div id="complete-panel" style="display: none;">
								<div class="inner cover">
									<h1 class="cover-heading">
										Complete!
									</h1>

									<h3><small class="text-muted">Link</small></h3>
									<p>
										<a id="complete-link" href=""></a>
									</p>

									<h3><small class="text-muted">Script</small></h3>
									<p>
										<input id="complete-script" type="text" onclick="this.select()" class="form-control" />
									</p>

									<p>
										<input type="checkbox" id="minify-checkbox" checked> <label for="minify-checkbox" class="text-muted">Minify</label>
									</p>
								</div>
							</div>
						</div>
					</div>
				</div>
			</body>
			<script>
				var complete = {};
				document.getElementById("minify-checkbox").onchange = function() {
					var value = document.getElementById("minify-checkbox").checked;
					var completeLink = document.getElementById("complete-link");
					var completeScript = document.getElementById("complete-script");
					completeLink.href = "https://jsgo.io/" + complete.short + (value ? "" : "$max");
					completeLink.innerHTML = "jsgo.io/" + complete.short + (value ? "" : "$max");
					completeScript.value = "https://cdn.jsgo.io/pkg/" + complete.path + "." + (value ? complete.hashmin : complete.hashmax) + ".js"
				}
				document.getElementById("btn").onclick = function(event) {
					event.preventDefault();
					var socket = new WebSocket("{{ .Scheme }}://{{ .Host }}/_ws/{{ .Path }}");

					var headerPanel = document.getElementById("header-panel");
					var buttonPanel = document.getElementById("button-panel");
					var progressPanel = document.getElementById("progress-panel");
					var errorPanel = document.getElementById("error-panel");
					var completePanel = document.getElementById("complete-panel");

					var completeLink = document.getElementById("complete-link");
					var completeScript = document.getElementById("complete-script");
					var errorMessage = document.getElementById("error-message");
					
					var done = {};

					socket.onopen = function() {
						buttonPanel.style.display = "none";
						progressPanel.style.display = "";
					};
					socket.onmessage = function (e) {
						var message = JSON.parse(e.data)
						switch (message.type) {
						case "queue":
						case "download":
						case "compile":
						case "store":
						case "index":
							if (done[message.type]) {
								// Messages might arrive out of order... Once we get a "done", ignore 
								// any more.
								break;
							}
							var item = document.getElementById(message.type+"-item");
							var span = document.getElementById(message.type+"-span");
							item.style.display = "";
							if (message.payload.done) {
								span.innerHTML = "Done";
								done[message.type] = true;
							} else if (message.payload.path) {
								span.innerHTML = message.payload.path;
							} else if (message.payload.position) {
								span.innerHTML = "Position " + message.payload.position;
							} else {
								span.innerHTML = "Starting";
							}
							break;
						case "complete":
							completePanel.style.display = "";
							progressPanel.style.display = "none";
							headerPanel.style.display = "none";
							complete = message.payload;
							completeLink.href = "https://jsgo.io/" + message.payload.short
							completeLink.innerHTML = "jsgo.io/" + message.payload.short
							completeScript.value = "https://cdn.jsgo.io/pkg/" + message.payload.path + "." + message.payload.hashmin + ".js"
							break;
						case "error":
							errorPanel.style.display = "";
							errorMessage.innerHTML = message.payload.message;
							break;
						}
					}
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

type funcWriter struct {
	f func(b []byte) error
}

func (f funcWriter) Write(b []byte) (n int, err error) {
	if err := f.f(b); err != nil {
		return 0, err
	}
	return len(b), nil
}
