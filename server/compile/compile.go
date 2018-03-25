package compile

import (
	"context"
	"io"
	"text/template"

	"cloud.google.com/go/storage"

	"fmt"

	"bytes"

	"encoding/json"

	"crypto/sha1"

	"path/filepath"

	"os"

	"io/ioutil"

	"strings"

	"github.com/dave/jsgo/builder"
	"github.com/dave/jsgo/builder/std"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/messages"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

type Compiler struct {
	root, path, temp billy.Filesystem
	send             func(messages.Message)
	log              io.Writer
}

func New(goroot, gopath billy.Filesystem, send func(messages.Message)) *Compiler {
	c := &Compiler{}
	c.root = goroot
	c.path = gopath
	c.temp = memfs.New()
	c.send = send
	return c
}

type CompileOutput struct {
	*builder.CommandOutput
	Hash []byte
}

func (c *Compiler) Compile(ctx context.Context, path string, log io.Writer) (min, max *CompileOutput, err error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()

	storer := NewStorer(ctx, client, c.send, config.ConcurrentStorageUploads)
	defer storer.Close()

	c.send(messages.Compiling{Starting: true})
	c.send(messages.Storing{Starting: true})

	data, outputMin, err := c.compileAndStore(ctx, path, storer, log, true)
	if err != nil {
		return nil, nil, err
	}
	_, outputMax, err := c.compileAndStore(ctx, path, storer, log, false)
	if err != nil {
		return nil, nil, err
	}

	c.send(messages.Compiling{Message: "Loader"})
	hashMin, err := genMain(ctx, storer, outputMin, true)
	if err != nil {
		return nil, nil, err
	}
	hashMax, err := genMain(ctx, storer, outputMax, false)
	if err != nil {
		return nil, nil, err
	}

	c.send(messages.Compiling{Message: "Index"})

	tpl, err := c.getIndexTpl(data.Dir)
	if err != nil {
		return nil, nil, err
	}

	if err := genIndex(storer, tpl, path, hashMin, true); err != nil {
		return nil, nil, err
	}
	if err := genIndex(storer, tpl, path, hashMax, false); err != nil {
		return nil, nil, err
	}

	c.send(messages.Compiling{Done: true})
	storer.Wait()
	if storer.Err != nil {
		fmt.Println("detected fail")
		return nil, nil, storer.Err
	}
	c.send(messages.Storing{Done: true})

	return &CompileOutput{CommandOutput: outputMin, Hash: hashMin}, &CompileOutput{CommandOutput: outputMax, Hash: hashMax}, nil

}

func (c *Compiler) defaultOptions(log io.Writer, min bool) *builder.Options {
	return &builder.Options{
		Root:        c.root,
		Path:        c.path,
		Temporary:   c.temp,
		Unvendor:    true,
		Initializer: true,
		Log:         log,
		Verbose:     true,
		Minify:      min,
		Standard:    std.Index,
		BuildTags:   []string{"jsgo"},
	}
}

func (c *Compiler) compileAndStore(ctx context.Context, path string, storer *Storer, log io.Writer, min bool) (*builder.PackageData, *builder.CommandOutput, error) {

	session := builder.NewSession(c.defaultOptions(log, min))

	data, archive, err := session.BuildImportPath(ctx, path)
	if err != nil {
		return nil, nil, err
	}

	if archive.Name != "main" {
		return nil, nil, fmt.Errorf("can't compile - %s is not a main package", path)
	}

	output, err := session.WriteCommandPackage(ctx, archive)
	if err != nil {
		return nil, nil, err
	}

	for _, po := range output.Packages {
		if po.Standard {
			continue
		}
		storer.AddJs(fmt.Sprintf("%s (minified)", po.Path), fmt.Sprintf("%s.%x.js", po.Path, po.Hash), po.Contents)
	}

	return data, output, nil
}

func (c *Compiler) getIndexTpl(dir string) (*template.Template, error) {
	fname := filepath.Join(dir, "index.jsgo.html")
	_, err := c.path.Stat(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return indexTemplate, nil
		}
		return nil, err
	}
	f, err := c.path.Open(fname)
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

func genIndex(storer *Storer, tpl *template.Template, path string, hash []byte, min bool) error {

	v := IndexVars{
		Path:   path,
		Hash:   fmt.Sprintf("%x", hash),
		Script: fmt.Sprintf("https://%s/%s.%x.js", config.PkgHost, path, hash),
	}

	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, v); err != nil {
		return err
	}

	fullpath := path
	if !min {
		fullpath = fmt.Sprintf("%s$max", path)
	}

	shortpath := strings.TrimPrefix(fullpath, "github.com/")

	var message string
	if min {
		message = "Index (minified)"
	} else {
		message = "Index (un-minified)"
	}
	storer.AddHtml(message, shortpath, buf.Bytes())
	storer.AddHtml("", fmt.Sprintf("%s/index.html", shortpath), buf.Bytes())

	if shortpath != fullpath {
		storer.AddHtml("", fullpath, buf.Bytes())
		storer.AddHtml("", fmt.Sprintf("%s/index.html", fullpath), buf.Bytes())
	}

	return nil

}

func genMain(ctx context.Context, storer *Storer, output *builder.CommandOutput, min bool) ([]byte, error) {

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
		PkgHost: config.PkgHost,
		Path:    output.Path,
		Json:    string(pkgJson),
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
	storer.AddJs(message, fmt.Sprintf("%s.%x.js", output.Path, hash), buf.Bytes())

	return hash, nil
}

type MainVars struct {
	Path    string
	Json    string
	PkgHost string
}

type PkgJson struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// minify with https://skalman.github.io/UglifyJS-online/

var mainTemplateMinified = template.Must(template.New("main").Parse(
	`"use strict";var $mainPkg,$load={};!function(){for(var n=0,t=0,e={{ .Json }},o=(document.getElementById("log"),function(){n++,window.jsgoProgress&&window.jsgoProgress(n,t),n==t&&function(){for(var n=0;n<e.length;n++)$load[e[n].path]();$mainPkg=$packages["{{ .Path }}"],$synthesizeMethods(),$packages.runtime.$init(),$go($mainPkg.$init,[]),$flushConsole()}()}),a=function(n){t++;var e=document.createElement("script");e.src=n,e.onload=o,e.onreadystatechange=o,document.head.appendChild(e)},s=0;s<e.length;s++)a("https://{{ .PkgHost }}/"+e[s].path+"."+e[s].hash+".js")}();`,
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
		get("https://{{ .PkgHost }}/" + info[i].path + "." + info[i].hash + ".js");
	}
})();`))
