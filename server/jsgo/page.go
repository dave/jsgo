package jsgo

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/store"
	"github.com/dave/services"
	"github.com/dustin/go-humanize"
)

func Page(w http.ResponseWriter, req *http.Request, database services.Database) {

	ctx, cancel := context.WithTimeout(req.Context(), config.PageTimeout)
	defer cancel()

	path := normalizePath(strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/"))

	if path == "" {
		http.Redirect(w, req, "https://github.com/dave/jsgo", http.StatusFound)
		return
	}

	var found bool
	var data store.CompileData
	var err error
	if config.LOCAL {
		found = false
	} else {
		found, data, err = store.Package(ctx, database, path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	type vars struct {
		Found         bool
		Path          string
		Last          string
		Host          string
		Scheme        string
		PkgHost       string
		IndexHost     string
		PkgProtocol   string
		IndexProtocol string
	}

	v := vars{}
	v.PkgHost = config.Host[config.Pkg]
	v.IndexHost = config.Host[config.Index]
	v.PkgProtocol = config.Protocol[config.Pkg]
	v.IndexProtocol = config.Protocol[config.Index]
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

func asset(url string) string {
	if config.LOCAL {
		return "/_local" + url[strings.LastIndex(url, "/"):]
	}
	return url
}

var compilePageTemplate = template.Must(template.New("main").Funcs(template.FuncMap{"Asset": asset}).Parse(`
<html>
	<head>
		<meta charset="utf-8">
		<link href="{{ Asset "https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css" }}" rel="stylesheet" integrity="sha384-Gn5384xqQ1aoWXA+058RXPxPg6fy4IWvTNh0E263XmFcJlSAwiGgFAW/dAiS6JXm" crossorigin="anonymous">
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
			<img style="position: absolute; top: 0; right: 0; border: 0;" src="{{ Asset "https://s3.amazonaws.com/github/ribbons/forkme_right_white_ffffff.png" }}" alt="Fork me on GitHub">
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
			completeLink.href = "{{ .IndexProtocol }}://{{ .IndexHost }}/" + (short ? final.Short : final.Path) + (minify ? "" : "$max");
			completeLink.innerHTML = "{{ .IndexHost }}/" + (short ? final.Short : final.Path) + (minify ? "" : "$max");
			completeScript.value = "{{ .PkgProtocol }}://{{ .PkgHost }}/" + final.Path + "." + (minify ? final.HashMin : final.HashMax) + ".js"
		}
		document.getElementById("minify-checkbox").onchange = refresh;
		document.getElementById("short-url-checkbox").onchange = refresh;
		document.getElementById("btn").onclick = function(event) {
			event.preventDefault();
			var socket = new WebSocket("{{ .Scheme }}://{{ .Host }}/_jsgo/");

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
