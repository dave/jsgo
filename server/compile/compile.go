package compile

import (
	"context"
	"io"
	"text/template"

	"fmt"

	"bytes"

	"encoding/json"

	"crypto/sha1"

	"path/filepath"

	"os"

	"io/ioutil"

	"strings"

	"sync"

	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/jsgo/messages"
	"github.com/dave/services"
	"github.com/dave/services/builder"
	"github.com/dave/services/fileserver/constor"
	"github.com/dave/services/session"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

type Compiler struct {
	session *session.Session
	send    func(services.Message)
}

func New(session *session.Session, send func(services.Message)) *Compiler {
	c := &Compiler{}
	c.session = session
	c.send = send
	return c
}

type CompileOutput struct {
	*builder.CommandOutput
	MainHash, IndexHash []byte
}

// Compile compiles path. If provided, source specifies the source packages. Including std lib packages
// in source forces them to be compiled (if they are not included the pre-compiled Archives are used).
func (c *Compiler) Compile(ctx context.Context, path string, play bool) (map[bool]*CompileOutput, error) {

	storer := constor.New(ctx, c.session.Fileserver, config.ConcurrentStorageUploads)
	defer storer.Close()

	c.send(messages.Compiling{Starting: true})
	c.send(messages.Storing{Starting: true})

	wg := &sync.WaitGroup{}

	outputs := map[bool]*builder.CommandOutput{}
	mainHashes := map[bool][]byte{}
	indexHashes := map[bool][]byte{}

	var outer error
	do := func(min bool) {
		defer wg.Done()

		var err error
		var data *builder.PackageData

		data, outputs[min], err = c.compileAndStore(ctx, path, storer, min)
		if err != nil {
			outer = err
			return
		}

		c.send(messages.Compiling{Message: "Loader"})

		mainHashes[min], err = c.genMain(ctx, storer, outputs[min], min)
		if err != nil {
			outer = err
			return
		}

		c.send(messages.Compiling{Message: "Index"})

		tpl, err := c.getIndexTpl(data.Dir)
		if err != nil {
			outer = err
			return
		}

		indexHashes[min], err = c.genIndex(storer, tpl, path, mainHashes[min], min, play)
		if err != nil {
			outer = err
			return
		}
	}

	wg.Add(1)
	go do(true)
	if !play {
		// TODO: make this configurable
		wg.Add(1)
		go do(false)
	}
	wg.Wait()
	if outer != nil {
		return nil, outer
	}
	c.send(messages.Compiling{Done: true})

	storer.Wait()
	if storer.Err != nil {
		return nil, storer.Err
	}
	c.send(messages.Storing{Done: true})

	out := map[bool]*CompileOutput{}
	for min := range outputs {
		out[min] = &CompileOutput{
			CommandOutput: outputs[min],
			MainHash:      mainHashes[min],
			IndexHash:     indexHashes[min],
		}
	}

	return out, nil

}

type compileWriter struct {
	send func(services.Message)
}

func (w compileWriter) Write(b []byte) (n int, err error) {
	w.send(messages.Compiling{Message: strings.TrimSuffix(string(b), "\n")})
	return len(b), nil
}

func (c *Compiler) defaultOptions(log io.Writer, min bool) *builder.Options {
	return &builder.Options{
		Temporary:   memfs.New(),
		Unvendor:    true,
		Initializer: true,
		Log:         log,
		Verbose:     true,
		Minify:      min,
		Standard:    std.Index,
	}
}

func (c *Compiler) compileAndStore(ctx context.Context, path string, storer *constor.Storer, min bool) (*builder.PackageData, *builder.CommandOutput, error) {

	b := builder.New(c.session, c.defaultOptions(compileWriter{c.send}, min))

	data, archive, err := b.BuildImportPath(ctx, path)
	if err != nil {
		return nil, nil, err
	}

	if archive.Name != "main" {
		return nil, nil, fmt.Errorf("can't compile - %s is not a main package", path)
	}

	output, err := b.WriteCommandPackage(ctx, archive)
	if err != nil {
		return nil, nil, err
	}

	for _, po := range output.Packages {
		if !po.Store {
			continue
		}
		storer.Add(constor.Item{
			Message:   po.Path,
			Name:      fmt.Sprintf("%s.%x.js", po.Path, po.Hash),
			Contents:  po.Contents,
			Bucket:    config.PkgBucket,
			Mime:      constor.MimeJs,
			Count:     true,
			Immutable: true,
			Changed: func(done bool) {
				messages.SendStoring(c.send, storer.Stats)
			},
		})
	}

	return data, output, nil
}

func (c *Compiler) getIndexTpl(dir string) (*template.Template, error) {
	fs := c.session.Filesystem(dir)
	fname := filepath.Join(dir, "index.jsgo.html")
	_, err := fs.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return indexTemplate, nil
		}
		return nil, err
	}
	f, err := fs.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	tpl, err := template.New("main").Parse(string(b))
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

type IndexVars struct {
	Path   string
	Hash   string
	Script string
}

var indexTemplate = template.Must(template.New("main").Parse(`
<html>
	<head>
		<meta charset="utf-8">
	</head>
	<body id="wrapper">
		<span id="jsgo-progress-span"></span>
		<script>
			window.jsgoProgress = function(count, total) {
				if (count === total) {
					document.getElementById("jsgo-progress-span").style.display = "none";
				} else {
					document.getElementById("jsgo-progress-span").innerHTML = count + "/" + total;
				}
			}
		</script>
		<script src="{{ .Script }}"></script>
	</body>
</html>
`))

func (c *Compiler) genIndex(storer *constor.Storer, tpl *template.Template, path string, loaderHash []byte, min, play bool) ([]byte, error) {

	v := IndexVars{
		Path:   path,
		Hash:   fmt.Sprintf("%x", loaderHash),
		Script: fmt.Sprintf("%s://%s/%s.%x.js", config.Protocol, config.PkgHost, path, loaderHash),
	}

	buf := &bytes.Buffer{}
	sha := sha1.New()

	if err := tpl.Execute(io.MultiWriter(buf, sha), v); err != nil {
		return nil, err
	}

	indexHash := sha.Sum(nil)

	if play {
		storer.Add(constor.Item{
			Message:   "Index",
			Name:      fmt.Sprintf("%x", indexHash),
			Contents:  buf.Bytes(),
			Bucket:    config.IndexBucket,
			Mime:      constor.MimeHtml,
			Count:     true,
			Immutable: true,
			Changed: func(done bool) {
				messages.SendStoring(c.send, storer.Stats)
			},
		})
		storer.Add(constor.Item{
			Message:   "",
			Name:      fmt.Sprintf("%x/index.html", indexHash),
			Contents:  buf.Bytes(),
			Bucket:    config.IndexBucket,
			Mime:      constor.MimeHtml,
			Count:     true,
			Immutable: true,
			Changed: func(done bool) {
				messages.SendStoring(c.send, storer.Stats)
			},
		})
	} else {
		fullpath := path
		if !min {
			fullpath = fmt.Sprintf("%s$max", path)
		}
		shortpath := strings.TrimPrefix(fullpath, "github.com/")

		storer.Add(constor.Item{
			Message:   "Index",
			Name:      shortpath,
			Contents:  buf.Bytes(),
			Bucket:    config.IndexBucket,
			Mime:      constor.MimeHtml,
			Count:     false,
			Immutable: false,
		})
		storer.Add(constor.Item{
			Message:   "",
			Name:      fmt.Sprintf("%s/index.html", shortpath),
			Contents:  buf.Bytes(),
			Bucket:    config.IndexBucket,
			Mime:      constor.MimeHtml,
			Count:     false,
			Immutable: false,
		})

		if shortpath != fullpath {
			storer.Add(constor.Item{
				Message:   "",
				Name:      fullpath,
				Contents:  buf.Bytes(),
				Bucket:    config.IndexBucket,
				Mime:      constor.MimeHtml,
				Count:     false,
				Immutable: false,
			})
			storer.Add(constor.Item{
				Message:   "",
				Name:      fmt.Sprintf("%s/index.html", fullpath),
				Contents:  buf.Bytes(),
				Bucket:    config.IndexBucket,
				Mime:      constor.MimeHtml,
				Count:     false,
				Immutable: false,
			})
		}
	}

	return indexHash, nil

}

func (c *Compiler) genMain(ctx context.Context, storer *constor.Storer, output *builder.CommandOutput, min bool) ([]byte, error) {

	preludeHash := std.Prelude[min]
	pkgs := []PkgJson{
		{
			// Always include the prelude dummy package first
			Path: "prelude",
			Hash: preludeHash,
		},
	}
	for _, po := range output.Packages {
		pkgs = append(pkgs, PkgJson{
			Path: po.Path,
			Hash: fmt.Sprintf("%x", po.Hash),
		})
	}

	pkgJson, err := json.Marshal(pkgs)
	if err != nil {
		return nil, err
	}

	m := MainVars{
		Protocol: config.Protocol,
		PkgHost:  config.PkgHost,
		Path:     output.Path,
		Json:     string(pkgJson),
	}

	buf := &bytes.Buffer{}
	var tmpl *template.Template
	if min {
		tmpl = mainTemplateMinified
	} else {
		tmpl = mainTemplate
	}
	if err := tmpl.Execute(buf, m); err != nil {
		return nil, err
	}

	s := sha1.New()
	if _, err := s.Write(buf.Bytes()); err != nil {
		return nil, err
	}

	hash := s.Sum(nil)

	var message string
	if min {
		message = "Loader (minified)"
	} else {
		message = "Loader (un-minified)"
	}
	storer.Add(constor.Item{
		Message:   message,
		Name:      fmt.Sprintf("%s.%x.js", output.Path, hash),
		Contents:  buf.Bytes(),
		Bucket:    config.PkgBucket,
		Mime:      constor.MimeJs,
		Count:     true,
		Immutable: true,
		Changed: func(done bool) {
			messages.SendStoring(c.send, storer.Stats)
		},
	})

	return hash, nil
}

type MainVars struct {
	Path     string
	Json     string
	PkgHost  string
	Protocol string
}

type PkgJson struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// minify with https://skalman.github.io/UglifyJS-online/

var mainTemplateMinified = template.Must(template.New("main").Parse(
	`"use strict";var $mainPkg,$load={};!function(){for(var n=0,t=0,e={{ .Json }},o=(document.getElementById("log"),function(){n++,window.jsgoProgress&&window.jsgoProgress(n,t),n==t&&function(){for(var n=0;n<e.length;n++)$load[e[n].path]();$mainPkg=$packages["{{ .Path }}"],$synthesizeMethods(),$packages.runtime.$init(),$go($mainPkg.$init,[]),$flushConsole()}()}),a=function(n){t++;var e=document.createElement("script");e.src=n,e.onload=o,e.onreadystatechange=o,document.head.appendChild(e)},s=0;s<e.length;s++)a("{{ .Protocol }}://{{ .PkgHost }}/"+e[s].path+"."+e[s].hash+".js")}();`,
))
var mainTemplate = template.Must(template.New("main").Parse(`"use strict";
var $mainPkg;
var $load = {};
(function(){
	var count = 0;
	var total = 0;
	var path = "{{ .Path }}";
	var info = {{ .Json }};
	var log = document.getElementById("log");
	var finished = function() {
		for (var i = 0; i < info.length; i++) {
			$load[info[i].path]();
		}
		$mainPkg = $packages[path];
		$synthesizeMethods();
		$packages["runtime"].$init();
		$go($mainPkg.$init, []);
		$flushConsole();
	}
	var done = function() {
		count++;
		if (window.jsgoProgress) { window.jsgoProgress(count, total); }
		if (count == total) { finished(); }
	}
	var get = function(url) {
		total++;
		var tag = document.createElement('script');
		tag.src = url;
		tag.onload = done;
		tag.onreadystatechange = done;
		document.head.appendChild(tag);
	}
	for (var i = 0; i < info.length; i++) {
		get("{{ .Protocol }}://{{ .PkgHost }}/" + info[i].path + "." + info[i].hash + ".js");
	}
})();`))
