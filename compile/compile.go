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

	"errors"

	"github.com/dave/jsgo/builder"
	"github.com/dave/jsgo/builder/std"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

type Compiler struct {
	root, path, temp billy.Filesystem
	log              io.Writer
}

func New(goroot, gopath billy.Filesystem, log io.Writer) *Compiler {
	c := &Compiler{}
	c.root = goroot
	c.path = gopath
	c.temp = memfs.New()
	c.log = log
	return c
}

func (c *Compiler) Compile(ctx context.Context, path string) (min, max []byte, err error) {
	fmt.Fprintln(c.log, "\nCompiling...")
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()
	bucket := client.Bucket("jsgo")

	options := func(min bool, verbose bool) *builder.Options {
		return &builder.Options{
			Root:        c.root,
			Path:        c.path,
			Temporary:   c.temp,
			Unvendor:    true,
			Initializer: true,
			Log:         c.log,
			Verbose:     verbose,
			Minify:      min,
			Standard:    std.Index,
		}
	}

	sessionMin := builder.NewSession(options(true, true))
	sessionMax := builder.NewSession(options(false, false))

	archiveMin, err := sessionMin.BuildImportPath(path)
	if err != nil {
		return nil, nil, err
	}
	archiveMax, err := sessionMax.BuildImportPath(path)
	if err != nil {
		return nil, nil, err
	}

	outputMin, err := sessionMin.WriteCommandPackage(archiveMin)
	if err != nil {
		return nil, nil, err
	}
	outputMax, err := sessionMax.WriteCommandPackage(archiveMax)
	if err != nil {
		return nil, nil, err
	}

	if len(outputMin.Packages) != len(outputMax.Packages) {
		return nil, nil, errors.New("minified output has different number of packages to non-minified")
	}

	fmt.Fprintln(c.log, "\nStoring...")
	for i := range outputMin.Packages {
		poMin := outputMin.Packages[i]
		poMax := outputMax.Packages[i]
		if poMin.Path != poMax.Path {
			return nil, nil, errors.New("minified output has different order of packages to non-minified")
		}
		if poMin.Standard {
			continue
		}
		fmt.Fprintln(c.log, poMin.Path)
		if err := sendToStorage(ctx, bucket, poMin.Path, poMin.Contents, poMin.Hash, true); err != nil {
			return nil, nil, err
		}
		if err := sendToStorage(ctx, bucket, poMax.Path, poMax.Contents, poMax.Hash, false); err != nil {
			return nil, nil, err
		}
	}

	hashMin, err := genMain(ctx, bucket, outputMin, true)
	if err != nil {
		return nil, nil, err
	}
	hashMax, err := genMain(ctx, bucket, outputMax, false)
	if err != nil {
		return nil, nil, err
	}
	return hashMin, hashMax, nil

}

func genMain(ctx context.Context, bucket *storage.BucketHandle, output *builder.CommandOutput, min bool) ([]byte, error) {

	var pkgs []PkgJson
	for _, po := range output.Packages {
		pkgs = append(pkgs, PkgJson{
			Path:     po.Path,
			Hash:     fmt.Sprintf("%x", po.Hash),
			Standard: po.Standard,
		})
	}

	pkgJson, err := json.Marshal(pkgs)
	if err != nil {
		return nil, err
	}

	m := MainVars{
		Prelude: std.PreludeHash,
		Path:    output.Path,
		Min:     min,
		Json:    string(pkgJson),
	}

	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, m); err != nil {
		return nil, err
	}

	s := sha1.New()
	if _, err := s.Write(buf.Bytes()); err != nil {
		return nil, err
	}

	hash := s.Sum(nil)

	if err := sendToStorage(ctx, bucket, output.Path, buf.Bytes(), hash, min); err != nil {
		return nil, err
	}

	return hash, nil
}

func sendToStorage(ctx context.Context, bucket *storage.BucketHandle, path string, contents, hash []byte, minified bool) error {
	min := ""
	if minified {
		min = ".min"
	}
	fpath := fmt.Sprintf("pkg/%s.%x%s.js", path, hash, min)
	if err := storeJs(ctx, bucket, bytes.NewBuffer(contents), fpath); err != nil {
		return err
	}
	return nil
}

func storeJs(ctx context.Context, bucket *storage.BucketHandle, reader io.Reader, filename string) error {
	wc := bucket.Object(filename).NewWriter(ctx)
	defer wc.Close()
	wc.ContentType = "application/javascript"
	if _, err := io.Copy(wc, reader); err != nil {
		return err
	}
	return nil
}

type MainVars struct {
	Prelude string
	Min     bool
	Path    string
	Json    string
}

type PkgJson struct {
	Path     string `json:"path"`
	Hash     string `json:"hash,omitempty"`
	Standard bool   `json:"std,omitempty"`
}

var tpl = template.Must(template.New("main").Parse(`
"use strict";
var $initialisers = {};
var $mainPkg;
var $min = {{ .Min }};
var $path = "{{ .Path }}";
var $pkgs = {{ .Json }};
var $progressCount = 0;
var $progressTotal = 0;
var logger = function(s) {
	document.getElementById("log").innerHTML = s;
}
var finished = function() {
	logger("Initialising...");
	$pkgs.forEach(function(pkg){
		$initialisers[pkg.path]();
	});
	$mainPkg = $packages[$path];
	$synthesizeMethods();
	$packages["runtime"].$init();
	$go($mainPkg.$init, []);
	$flushConsole();
}
var load = function(url) {
	$progressTotal++;
    var tag = document.createElement('script');
    tag.src = url;
	var done = function() {
		$progressCount++;
		logger("Loading " + $progressCount + " / " + $progressTotal);
		if ($progressCount == $progressTotal) {
			finished();
		}
	}
    tag.onload = done;
    tag.onreadystatechange = done;
    document.head.appendChild(tag);
}
load("https://storage.googleapis.com/jsgo/sys/prelude.{{ .Prelude }}.js");
$pkgs.forEach(function(pkg){
	var dir = pkg.std ? "sys" : "pkg";
	var min = $min ? ".min" : "";
	load("https://storage.googleapis.com/jsgo/" + dir + "/" + pkg.path + "." + pkg.hash + min + ".js");
});
`))
