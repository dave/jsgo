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

	"cloud.google.com/go/datastore"

	"google.golang.org/appengine"

	"path"

	"errors"

	"regexp"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/builder"
	"github.com/dave/jsgo/builder/std"
	"github.com/dave/jsgo/compile"
	"github.com/dave/jsgo/getter"
	"github.com/dave/jsgo/server/logger"
	"github.com/dustin/go-humanize"
	"github.com/shurcooL/httpgzip"
	"golang.org/x/net/websocket"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

const PROJECT_ID = "jsgo-192815"

const WriteTimeout = time.Second * 2

func SocketHandler(ws *websocket.Conn) {
	path := strings.TrimSuffix(strings.TrimPrefix(ws.Request().URL.Path, "/_ws/"), "/")

	path = normalizePath(path)

	log := logger.New(ws)

	if err := doSocketCompile(ws, path, log); err != nil {
		log.Log(logger.Error, logger.ErrorPayload{
			Path:    path,
			Message: err.Error(),
		})
		return
	}
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
		data := CompileData{
			Path:    path,
			Time:    time.Now(),
			Success: false,
			Error:   err.Error(),
			Ip:      ws.Request().RemoteAddr,
		}
		Save(ctx, path, data) // ignore error
		return err
	}
	log.Log(logger.Download, logger.DownloadingPayload{Done: true})

	c := compile.New(assets.Assets, fs, log)
	hashMin, hashMax, outputMin, outputMax, err := c.Compile(ctx, path)
	if err != nil {
		data := CompileData{
			Path:    path,
			Time:    time.Now(),
			Success: false,
			Error:   err.Error(),
			Ip:      ws.Request().RemoteAddr,
		}
		Save(ctx, path, data) // ignore error
		return err
	}

	getCc := func(hash []byte, output *builder.CommandOutput) CompileContents {
		val := CompileContents{}
		val.Main = fmt.Sprintf("%x", hash)
		val.Prelude = std.PreludeHash
		for _, p := range output.Packages {
			val.Packages = append(val.Packages, CompilePackage{
				Path:     p.Path,
				Standard: p.Standard,
				Hash:     fmt.Sprintf("%x", p.Hash),
			})
		}
		return val
	}

	data := CompileData{
		Path:    path,
		Time:    time.Now(),
		Success: true,
		Min:     getCc(hashMin, outputMin),
		Max:     getCc(hashMax, outputMax),
		Ip:      ws.Request().RemoteAddr,
	}

	if err := Save(ctx, path, data); err != nil {
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

func FaviconHandler(w http.ResponseWriter, req *http.Request) {
	if err := ServeStatic("favicon.ico", w, req, "image/x-icon"); err != nil {
		http.Error(w, "error serving static file", 500)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

var gistWithUsernameRegex = regexp.MustCompile(`^gist\.github\.com/[A-Za-z0-9_.\-]+/([a-f0-9]+)(/[\p{L}0-9_.\-]+)*$`)
var githubUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]{0,38}$`)

func normalizePath(path string) string {

	// We should normalize gist urls by removing the username part
	if strings.HasPrefix(path, "gist.github.com/") {
		matches := gistWithUsernameRegex.FindStringSubmatch(path)
		if len(matches) > 1 {
			return fmt.Sprintf("gist.github.com/%s", matches[1])
		}
	}

	// Add github.com if the first part of the path is not a hostname and matches the github username regex
	if strings.Contains(path, "/") {
		firstPart := path[:strings.Index(path, "/")]
		if !strings.Contains(firstPart, ".") && githubUsernameRegex.MatchString(firstPart) {
			return fmt.Sprintf("github.com/%s", path)
		}
	}

	return path
}

func serveCompilePage(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")

	path = normalizePath(path)

	if path == "" {
		http.Redirect(w, req, "https://github.com/dave/jsgo", http.StatusFound)
		return
	}

	ctx := appengine.NewContext(req)
	found, data, err := Lookup(ctx, path)
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
				<style>
/*
 * Globals
 */

/* Links */
a,
a:focus,
a:hover {
  color: #fff;
}

/* Custom default button */
.btn-secondary,
.btn-secondary:hover,
.btn-secondary:focus {
  color: #333;
  text-shadow: none; /* Prevent inheritance from body */
  background-color: #fff;
  border: .05rem solid #fff;
}


/*
 * Base structure
 */

html,
body {
  height: 100%;
  background-color: #333;
}
body {
  color: #fff;
  text-align: center;
}
#error-panel {
	text-align: left;
}

/* Extra markup and styles for table-esque vertical and horizontal centering */
.site-wrapper {
  display: table;
  width: 100%;
  height: 100%; /* For at least Firefox */
  min-height: 100%;
  -webkit-box-shadow: inset 0 0 5rem rgba(0,0,0,.5);
          box-shadow: inset 0 0 5rem rgba(0,0,0,.5);
}
.site-wrapper-inner {
  display: table-cell;
  vertical-align: top;
}
.cover-container {
  margin-right: auto;
  margin-left: auto;
}

/* Padding for spacing */
.inner {
  padding: 2rem;
}


/*
 * Header
 */

.masthead {
  margin-bottom: 2rem;
}

.masthead-brand {
  margin-bottom: 0;
}

.nav-masthead .nav-link {
  padding: .25rem 0;
  font-weight: bold;
  color: rgba(255,255,255,.5);
  background-color: transparent;
  border-bottom: .25rem solid transparent;
}

.nav-masthead .nav-link:hover,
.nav-masthead .nav-link:focus {
  border-bottom-color: rgba(255,255,255,.25);
}

.nav-masthead .nav-link + .nav-link {
  margin-left: 1rem;
}

.nav-masthead .active {
  color: #fff;
  border-bottom-color: #fff;
}

@media (min-width: 48em) {
  .masthead-brand {
    float: left;
  }
  .nav-masthead {
    float: right;
  }
}


/*
 * Cover
 */

.cover {
  padding: 0 1.5rem;
}
.cover .btn-lg {
  padding: .75rem 1.25rem;
  font-weight: bold;
}


/*
 * Footer
 */

.mastfoot {
  color: rgba(255,255,255,.5);
}


/*
 * Affix and center
 */

@media (min-width: 40em) {
  /* Pull out the header and footer */
  .masthead {
    position: fixed;
    top: 0;
  }
  .mastfoot {
    position: fixed;
    bottom: 0;
  }
  /* Start the vertical centering */
  .site-wrapper-inner {
    vertical-align: middle;
  }
  /* Handle the widths */
  .masthead,
  .mastfoot,
  .cover-container {
    width: 100%; /* Must be percentage or pixels for horizontal alignment */
  }
}

@media (min-width: 62em) {
  .masthead,
  .mastfoot,
  .cover-container {
    width: 42rem;
  }
}
				</style>
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

					var downloadItem = document.getElementById("download-item");
					var compileItem = document.getElementById("compile-item");
					var storeItem = document.getElementById("store-item");
					var indexItem = document.getElementById("index-item");

					var downloadSpan = document.getElementById("download-span");
					var compileSpan = document.getElementById("compile-span");
					var storeSpan = document.getElementById("store-span");
					var indexSpan = document.getElementById("index-span");

					var completeLink = document.getElementById("complete-link");
					var completeScript = document.getElementById("complete-script");
					var errorMessage = document.getElementById("error-message");

					socket.onopen = function() {
						buttonPanel.style.display = "none";
						progressPanel.style.display = "";
					};
					socket.onmessage = function (e) {
						var message = JSON.parse(e.data)
						switch (message.type) {
						case "download":
							downloadItem.style.display = "";
							if (message.payload.done) {
								downloadSpan.innerHTML = "Done";
							} else if (message.payload.path) {
								downloadSpan.innerHTML = message.payload.path;
							}
							break;
						case "compile":
							compileItem.style.display = "";
							if (message.payload.done) {
								compileSpan.innerHTML = "Done";
							} else if (message.payload.path) {
								compileSpan.innerHTML = message.payload.path;
							}
							break;
						case "store":
							storeItem.style.display = "";
							if (message.payload.done) {
								storeSpan.innerHTML = "Done";
							} else if (message.payload.path) {
								storeSpan.innerHTML = message.payload.path;
							}
							break;
						case "index":
							indexItem.style.display = "";
							if (message.payload.done) {
								indexSpan.innerHTML = "Done";
							} else if (message.payload.path) {
								indexSpan.innerHTML = message.payload.path;
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

type CompileData struct {
	Path    string
	Time    time.Time
	Success bool
	Error   string
	Min     CompileContents
	Max     CompileContents
	Ip      string
}

type CompileContents struct {
	Main     string
	Prelude  string
	Packages []CompilePackage
}

type CompilePackage struct {
	Path     string
	Standard bool
	Hash     string
}

func Save(ctx context.Context, path string, data CompileData) error {
	client, err := datastore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return err
	}
	if _, err := client.Put(ctx, compileKey(), &data); err != nil {
		return err
	}
	if data.Success {
		if _, err := client.Put(ctx, packageKey(path), &data); err != nil {
			return err
		}
	}
	return nil
}

func Lookup(ctx context.Context, path string) (bool, CompileData, error) {
	client, err := datastore.NewClient(ctx, PROJECT_ID)
	if err != nil {
		return false, CompileData{}, err
	}
	var data CompileData

	/*q := datastore.NewQuery("Compile").Filter("Success =", true).Filter("Path =", path).Order("-Time").Limit(1)

	if _, err := client.Run(ctx, q).Next(&data); err != nil {
		if err == iterator.Done {
			return false, CompileData{}, nil
		}
		return false, CompileData{}, err
	}*/

	if err := client.Get(ctx, packageKey(path), &data); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return false, CompileData{}, nil
		}
		return false, CompileData{}, err
	}
	return true, data, nil
}

func compileKey() *datastore.Key {
	return datastore.IncompleteKey("Compile", nil)
}

func packageKey(path string) *datastore.Key {
	return datastore.NameKey("Package", path, nil)
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

func hasQuery(req *http.Request, id string) bool {
	_, value := req.URL.Query()[id]
	return value
}
