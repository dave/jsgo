package server

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	pathpkg "path"

	"errors"

	"regexp"

	"context"

	"sync"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/server/queue"
	"github.com/dave/jsgo/server/store"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/websocket"
	"github.com/shurcooL/httpgzip"
	"gopkg.in/src-d/go-billy.v4"
)

func New(shutdown chan struct{}) *Handler {
	h := &Handler{
		mux:       http.NewServeMux(),
		shutdown:  shutdown,
		Queue:     queue.New(config.MaxConcurrentCompiles, config.MaxQueue),
		Waitgroup: &sync.WaitGroup{},
	}
	h.mux.HandleFunc("/", h.PageHandler)
	//h.mux.HandleFunc("/_foo/", h.Foo)
	h.mux.HandleFunc("/_ws/", h.SocketHandler)
	h.mux.HandleFunc("/_pg/", h.SocketHandler)
	h.mux.HandleFunc("/favicon.ico", h.IconHandler)
	h.mux.HandleFunc("/compile.css", h.CssHandler)
	h.mux.HandleFunc("/_ah/health", h.HealthCheckHandler)
	h.mux.HandleFunc("/_go", h.GoCheckHandler)
	return h
}

func (h *Handler) Foo(w http.ResponseWriter, req *http.Request) {

	// creates the in-cluster config
	cfg, err := rest.InClusterConfig()
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	fmt.Fprintf(w, "There are %d pods in the cluster:\n\n", len(pods.Items))

	for k, v := range pods.Items {
		fmt.Fprintf(w, "Pod %d:\n", k)
		fmt.Fprintf(w, "TypeMeta: %#v\n", v.TypeMeta)
		fmt.Fprintf(w, "ObjectMeta: %#v\n", v.ObjectMeta)
		fmt.Fprintf(w, "Spec: %#v\n", v.Spec)
		fmt.Fprintf(w, "Status: %#v\n", v.Status)
		fmt.Fprintln(w)
	}

	// Examples for error handling:
	// - Use helper functions like e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	/*
		_, err = clientset.CoreV1().Pods("default").Get("example-xxxxx", metav1.GetOptions{})
		if kerrors.IsNotFound(err) {
			fmt.Printf("Pod not found\n")
		} else if statusError, isStatus := err.(*kerrors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod\n")
		}

		time.Sleep(10 * time.Second)
	*/

}

type Handler struct {
	Waitgroup *sync.WaitGroup
	Queue     *queue.Queue
	mux       *http.ServeMux
	shutdown  chan struct{}
}

func (h *Handler) PageHandler(w http.ResponseWriter, req *http.Request) {

	ctx, cancel := context.WithTimeout(req.Context(), config.PageTimeout)
	defer cancel()

	path := normalizePath(strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/"))

	if path == "" {
		http.Redirect(w, req, "https://github.com/dave/jsgo", http.StatusFound)
		return
	}

	found, data, err := store.Lookup(ctx, path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type vars struct {
		Found     bool
		Path      string
		Last      string
		Host      string
		Scheme    string
		PkgHost   string
		IndexHost string
	}

	v := vars{}
	v.PkgHost = config.PkgHost
	v.IndexHost = config.IndexHost
	v.Host = req.Host
	v.Path = path
	if req.Host == config.CompileHost {
		v.Scheme = "wss"
	} else {
		v.Scheme = "ws"
	}
	if found {
		v.Found = true
		v.Last = humanize.Time(data.Time)
	}

	if err := pageTemplate.Execute(w, v); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

var pageTemplate = template.Must(template.New("main").Parse(`
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

					<div id="complete-panel" style="display: none;">
						<div class="inner cover">
							<h1 class="cover-heading">
								Complete!
							</h1>

							<h3><small class="text-muted">Link</small></h3>
							<p>
								<a id="complete-link" href=""></a>
							</p>

							<h3><small class="text-muted">Loader JS</small></h3>
							<p>
								<input id="complete-script" type="text" onclick="this.select()" class="form-control" />
							</p>

							<p>
								<small>
									<input type="checkbox" id="minify-checkbox" checked> <label for="minify-checkbox" class="text-muted">Minify</label>
								</small>
								<small id="short-url-checkbox-holder">
									<input type="checkbox" id="short-url-checkbox" checked> <label for="short-url-checkbox" class="text-muted">Short URL</label>
								</small>
							</p>
							
						</div>
					</div>

					<div id="progress-panel" style="display: none;">
						<table class="table table-dark">
							<tbody>
								<tr id="queueing-item" style="display: none;">
									<th scope="row" class="w-25">Queued:</th>
									<td class="w-75"><span id="queueing-span"></span></td>
								</tr>
								<tr id="downloading-item" style="display: none;">
									<th scope="row" class="w-25">Downloading:</th>
									<td class="w-75"><span id="downloading-span"></span></td>
								</tr>
								<tr id="compiling-item" style="display: none;">
									<th scope="row" class="w-25">Compiling:</th>
									<td class="w-75"><span id="compiling-span"></span></td>
								</tr>
								<tr id="storing-item" style="display: none;">
									<th scope="row" class="w-25">Storing:</th>
									<td class="w-75"><span id="storing-span"></span></td>
								</tr>
							</tbody>
						</table>
					</div>
					<div id="error-panel" style="display: none;" class="alert alert-warning" role="alert">
						<h4 class="alert-heading">Error</h4>
						<pre id="error-message"></pre>
					</div>
				</div>
			</div>
		</div>
	</body>
	<script>
		var final = {};
		var refresh = function() {
			var minify = document.getElementById("minify-checkbox").checked;
			var short = document.getElementById("short-url-checkbox").checked;
			var completeLink = document.getElementById("complete-link");
			var completeScript = document.getElementById("complete-script");
			var shortUrlCheckboxHolder = document.getElementById("short-url-checkbox-holder");
			
			shortUrlCheckboxHolder.style.display = (final.Short == final.Path) ? "none" : "";
			completeLink.href = "https://{{ .IndexHost }}/" + (short ? final.Short : final.Path) + (minify ? "" : "$max");
			completeLink.innerHTML = "{{ .IndexHost }}/" + (short ? final.Short : final.Path) + (minify ? "" : "$max");
			completeScript.value = "https://{{ .PkgHost }}/" + final.Path + "." + (minify ? final.HashMin : final.HashMax) + ".js"
		}
		document.getElementById("minify-checkbox").onchange = refresh;
		document.getElementById("short-url-checkbox").onchange = refresh;
		document.getElementById("btn").onclick = function(event) {
			event.preventDefault();
			var socket = new WebSocket("{{ .Scheme }}://{{ .Host }}/_ws/{{ .Path }}");

			var headerPanel = document.getElementById("header-panel");
			var buttonPanel = document.getElementById("button-panel");
			var progressPanel = document.getElementById("progress-panel");
			var errorPanel = document.getElementById("error-panel");
			var completePanel = document.getElementById("complete-panel");
			var errorMessage = document.getElementById("error-message");
			
			var done = {};
			var complete = false;

			socket.onopen = function() {
				buttonPanel.style.display = "none";
				progressPanel.style.display = "";
			};
			socket.onmessage = function (e) {
				var payload = JSON.parse(e.data)
				switch (payload.Type) {
				case "Queueing":
				case "Downloading":
				case "Compiling":
				case "Storing":
					if (done[payload.Type]) {
						// Messages might arrive out of order... Once we get a "done", ignore 
						// any more.
						break;
					}
					var item = document.getElementById(payload.Type.toLowerCase()+"-item");
					var span = document.getElementById(payload.Type.toLowerCase()+"-span");
					item.style.display = "";
					if (payload.Message.Done) {
						span.innerHTML = "Done";
						done[payload.Type] = true;
					} else if (payload.Message.Starting) {
						span.innerHTML = "Starting";
					} else if (payload.Message.Message) {
						span.innerHTML = payload.Message.Message;
					} else if (payload.Message.Position) {
						span.innerHTML = "Position " + payload.Message.Position;
					} else if (payload.Message.Finished !== undefined) {
						span.innerHTML = payload.Message.Finished + " finished, " + payload.Message.Unchanged + " unchanged, " + payload.Message.Remain + " remain.";
					} else {
						span.innerHTML = "Starting";
					}
					break;
				case "Complete":
					complete = true;
					final = payload.Message;
					completePanel.style.display = "";
					progressPanel.style.display = "none";
					headerPanel.style.display = "none";
					refresh();
					break;
				case "Error":
					if (complete) {
						break;
					}
					complete = true;
					errorPanel.style.display = "";
					errorMessage.innerHTML = payload.Message.Message;
					break;
				}
				socket.onclose = function() {
					if (complete) {
						return;
					}
					errorPanel.style.display = "";
					errorMessage.innerHTML = "server disconnected";
				}
			}
		};
	</script>
</html>
`))

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

const (
	TypeCompile = iota
	TypePlaygroundCompile
	TypePlaygroundDeploy
)

func (h *Handler) SocketHandler(w http.ResponseWriter, req *http.Request) {

	h.Waitgroup.Add(1)
	defer h.Waitgroup.Done()

	ctx, cancel := context.WithTimeout(req.Context(), config.CompileTimeout)
	defer cancel()

	var compileType int
	var prefix string

	if strings.HasPrefix(req.URL.Path, "/_pg/") {
		compileType = TypePlaygroundCompile
		prefix = "/_pg/"
	} else {
		compileType = TypeCompile
		prefix = "/_ws/"
	}

	path := normalizePath(strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, prefix), "/"))

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		storeError(ctx, path, fmt.Errorf("upgrading request to websocket: %v", err), req)
		return
	}

	var sendWg sync.WaitGroup
	sendChan := make(chan messages.Message, 256)
	receive := make(chan messages.Message, 256)

	send := func(message messages.Message) {
		sendWg.Add(1)
		sendChan <- message
	}

	defer func() {
		// wait for sends to finish before closing websocket
		sendWg.Wait()
		conn.Close()
	}()

	// Recover from any panic and log the error.
	defer func() {
		if r := recover(); r != nil {
			sendAndStoreError(ctx, send, path, errors.New(fmt.Sprintf("panic recovered: %s", r)), req)
		}
	}()

	// Set up a ticker to ping the client regularly
	go func() {
		ticker := time.NewTicker(config.WebsocketPingPeriod)
		defer func() {
			ticker.Stop()
			cancel()
		}()
		for {
			select {
			case message, ok := <-sendChan:
				func() {
					defer sendWg.Done()
					conn.SetWriteDeadline(time.Now().Add(config.WebsocketWriteTimeout))
					if !ok {
						// The send channel was closed.
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
					b, err := messages.Marshal(message)
					if err != nil {
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
					if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
						// Error writing message, close and exit
						conn.WriteMessage(websocket.CloseMessage, []byte{})
						return
					}
				}()
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(config.WebsocketWriteTimeout))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// React to pongs from the client
	go func() {
		defer func() {
			cancel()
		}()
		conn.SetReadDeadline(time.Now().Add(config.WebsocketPongTimeout))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(config.WebsocketPongTimeout))
			return nil
		})
		for {
			_, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					// Don't bother storing an error if the client disconnects gracefully
					break
				}
				storeError(ctx, path, err, req)
				break
			}
			message, err := messages.Unmarshal(messageBytes)
			if err != nil {
				storeError(ctx, path, err, req)
				break
			}
			select {
			case receive <- message:
			default:
			}
		}
	}()

	// React to the server shutdown signal
	go func() {
		select {
		case <-h.shutdown:
			sendAndStoreError(ctx, send, path, errors.New("server shut down"), req)
			cancel()
		case <-ctx.Done():
		}
	}()

	// Request a slot in the queue...
	start, end, err := h.Queue.Slot(func(position int) {
		send(messages.Queueing{Position: position})
	})
	if err != nil {
		sendAndStoreError(ctx, send, path, err, req)
		return
	}

	// Signal to the queue that processing has finished.
	defer close(end)

	// Wait for the slot to become available.
	select {
	case <-start:
		// continue
	case <-ctx.Done():
		return
	}

	// Send a message to the client that queue step has finished.
	send(messages.Queueing{Done: true})

	switch compileType {
	case TypeCompile:
		jsgoCompile(ctx, path, req, send)
	case TypePlaygroundCompile:
		playgroundCompile(ctx, path, req, send, receive)
	}

	return
}

func sendAndStoreError(ctx context.Context, send func(messages.Message), path string, err error, req *http.Request) {
	storeError(ctx, path, err, req)
	sendError(send, path, err)
}

func sendError(send func(messages.Message), path string, err error) {
	if p, ok := err.(Pather); ok {
		path = p.Path()
	}
	send(messages.Error{
		Path:    path,
		Message: err.Error(),
	})
}

func storeError(ctx context.Context, path string, err error, req *http.Request) {

	if err == queue.TooManyItemsQueued {
		// If the server is getting flooded by a DOS, this will prevent database flooding
		return
	}

	if p, ok := err.(Pather); ok {
		path = p.Path()
	}

	// ignore errors when logging an error
	store.Save(ctx, path, store.CompileData{
		Path:    path,
		Time:    time.Now(),
		Success: false,
		Error:   err.Error(),
		Ip:      req.Header.Get("X-Forwarded-For"),
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

func (h *Handler) GoCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, runtime.NumGoroutine())
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
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

type downloadWriter struct {
	send func(messages.Message)
}

func (w downloadWriter) Write(b []byte) (n int, err error) {
	w.send(messages.Downloading{Message: strings.TrimSuffix(string(b), "\n")})
	return len(b), nil
}

type Pather interface {
	Path() string
}
