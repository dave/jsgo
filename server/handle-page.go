package server

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"runtime"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/store"
	humanize "github.com/dustin/go-humanize"
)

func (h *Handler) PageHandler(w http.ResponseWriter, req *http.Request) {
	if config.DEV {
		// always run the compile page in dev mode
		h.handleCompilePage(w, req)
		return
	}
	switch req.Host {
	case "play.jsgo.io":
		h.handlePlayPage(w, req)
		return
	case "compile.jsgo.io":
		h.handleCompilePage(w, req)
		return
	default:
		http.Error(w, fmt.Sprintf("unknown host %s", req.Host), 500)
		return
	}
}

func (h *Handler) handleCompilePage(w http.ResponseWriter, req *http.Request) {

	ctx, cancel := context.WithTimeout(req.Context(), config.PageTimeout)
	defer cancel()

	path := normalizePath(strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/"))

	if path == "" {
		http.Redirect(w, req, "https://github.com/dave/jsgo", http.StatusFound)
		return
	}

	found, data, err := store.Package(ctx, path)
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
		Protocol  string
	}

	v := vars{}
	v.PkgHost = config.PkgHost
	v.IndexHost = config.IndexHost
	v.Protocol = config.Protocol
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

	if err := compilePageTemplate.Execute(w, v); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

var compilePageTemplate = template.Must(template.New("main").Parse(`
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
		<a href="https://github.com/dave/jsgo" target="_blank">
			<img style="position: absolute; top: 0; right: 0; border: 0;" src="https://s3.amazonaws.com/github/ribbons/forkme_right_white_ffffff.png" alt="Fork me on GitHub">
		</a>
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
			completeLink.href = "{{ .Protocol }}://{{ .IndexHost }}/" + (short ? final.Short : final.Path) + (minify ? "" : "$max");
			completeLink.innerHTML = "{{ .IndexHost }}/" + (short ? final.Short : final.Path) + (minify ? "" : "$max");
			completeScript.value = "{{ .Protocol }}://{{ .PkgHost }}/" + final.Path + "." + (minify ? final.HashMin : final.HashMax) + ".js"
		}
		document.getElementById("minify-checkbox").onchange = refresh;
		document.getElementById("short-url-checkbox").onchange = refresh;
		document.getElementById("btn").onclick = function(event) {
			event.preventDefault();
			var socket = new WebSocket("{{ .Scheme }}://{{ .Host }}/_ws/");

			var headerPanel = document.getElementById("header-panel");
			var buttonPanel = document.getElementById("button-panel");
			var progressPanel = document.getElementById("progress-panel");
			var errorPanel = document.getElementById("error-panel");
			var completePanel = document.getElementById("complete-panel");
			var errorMessage = document.getElementById("error-message");
			
			var done = {};
			var complete = false;

			socket.onopen = function() {
				socket.send(JSON.stringify({
					"Type": "Compile",
					"Message": {
						"Path": "{{ .Path }}"
					}
				}));
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

func (h *Handler) handlePlayPage(w http.ResponseWriter, req *http.Request) {

	ctx, cancel := context.WithTimeout(req.Context(), config.PageTimeout)
	defer cancel()

	found, c, err := store.Package(ctx, "github.com/dave/play")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if !found {
		http.Error(w, "play package not found", 500)
		return
	}
	url := fmt.Sprintf("https://pkg.jsgo.io/github.com/dave/play.%s.js", c.Min.Main)

	v := struct {
		Script string
		Count  int
	}{
		Script: url,
		Count:  runtime.NumGoroutine(),
	}

	if err := playPageTemplate.Execute(w, v); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

var playPageTemplate = template.Must(template.New("main").Parse(`<html>
	<head>
		<meta charset="utf-8">
		<script async src="https://www.googletagmanager.com/gtag/js?id=UA-118676357-1"></script>
        <script>
            window.dataLayer = window.dataLayer || [];
            function gtag(){dataLayer.push(arguments);}
            gtag('js', new Date());
            gtag('config', 'UA-118676357-1');
        </script>
        <link href="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
        <script src="https://code.jquery.com/jquery-3.2.1.slim.min.js" integrity="sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN" crossorigin="anonymous"></script>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js" integrity="sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q" crossorigin="anonymous"></script>
        <script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js" integrity="sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl" crossorigin="anonymous"></script>
        <script src="https://cdnjs.cloudflare.com/ajax/libs/ace/1.3.3/ace.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/ace/1.3.3/ext-linking.js"></script>
	</head>
	<body id="wrapper" style="margin: 0;" data-count="{{ .Count }}">
		<div id="progress-holder" style="width: 100%; padding: 25%;">
			<div class="progress">
				<div id="progress-bar" class="progress-bar" role="progressbar" style="width: 0%" aria-valuenow="0" aria-valuemin="0" aria-valuemax="100"></div>
			</div>
		</div>
		<script>
			window.jsgoProgress = function(count, total) {
				var value = (count * 100.0) / (total * 1.0);
				var bar = document.getElementById("progress-bar");
				bar.style.width = value+"%";
				bar.setAttribute('aria-valuenow', value);
				if (count === total) {
					document.getElementById("progress-holder").style.display = "none";
				}
			}
		</script>
    	<script src="{{ .Script }}"></script>
	</body>
</html>`))
