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
	"github.com/dave/jsgo/common"
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
	bucketCdn := client.Bucket("cdn.jsgo.io")
	bucketIndex := client.Bucket("jsgo.io")

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

	if archiveMin.Name != "main" {
		return nil, nil, fmt.Errorf("can't compile - %s is not a main package", path)
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
		if err := sendToStorage(ctx, bucketCdn, poMin.Path, poMin.Contents, poMin.Hash); err != nil {
			return nil, nil, err
		}
		if err := sendToStorage(ctx, bucketCdn, poMax.Path, poMax.Contents, poMax.Hash); err != nil {
			return nil, nil, err
		}
	}

	hashMin, err := genMain(ctx, bucketCdn, outputMin, true)
	if err != nil {
		return nil, nil, err
	}
	hashMax, err := genMain(ctx, bucketCdn, outputMax, false)
	if err != nil {
		return nil, nil, err
	}

	fmt.Fprintln(c.log, "\nGenerating index...")
	if err := genIndex(ctx, bucketIndex, path, hashMin, true); err != nil {
		return nil, nil, err
	}
	if err := genIndex(ctx, bucketIndex, path, hashMax, false); err != nil {
		return nil, nil, err
	}
	fmt.Fprintln(c.log, path)

	return hashMin, hashMax, nil

}

type IndexVars struct {
	Path   string
	Hash   string
	Script string
}

var indexTpl = template.Must(template.New("main").Parse(`
<html>
	<head>
		<meta charset="utf-8">
	</head>
	<body id="wrapper">
		<span id="log">Loading...</span>
		<script src="{{ .Script }}"></script>
	</body>
</html>
`))

func genIndex(ctx context.Context, bucket *storage.BucketHandle, path string, hash []byte, min bool) error {

	v := IndexVars{
		Path:   path,
		Hash:   fmt.Sprintf("%x", hash),
		Script: fmt.Sprintf("https://cdn.jsgo.io/pkg/%s.%x.js", path, hash),
	}

	buf := &bytes.Buffer{}
	if err := indexTpl.Execute(buf, v); err != nil {
		return err
	}

	fullpath := path
	if !min {
		fullpath = fmt.Sprintf("%s$max", path)
	}

	if err := sendIndex(ctx, bucket, fullpath, buf.Bytes()); err != nil {
		return err
	}

	return nil

}

func sendIndex(ctx context.Context, bucket *storage.BucketHandle, path string, contents []byte) error {

	// For URLs of the form jsgo.io/path
	if err := storeHtml(ctx, bucket, bytes.NewBuffer(contents), path); err != nil {
		return err
	}

	// For URLs of the form jsgo.io/path/
	fpath := fmt.Sprintf("%s/index.html", path)
	if err := storeHtml(ctx, bucket, bytes.NewBuffer(contents), fpath); err != nil {
		return err
	}

	return nil
}

func storeHtml(ctx context.Context, bucket *storage.BucketHandle, reader io.Reader, filename string) error {
	wc := bucket.Object(filename).NewWriter(ctx)
	defer wc.Close()
	wc.ContentType = "text/html"
	wc.CacheControl = "no-cache"
	if _, err := io.Copy(wc, reader); err != nil {
		return err
	}
	return nil
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
	if _, err := s.Write([]byte{common.HASH_VERSION}); err != nil {
		return nil, err
	}

	hash := s.Sum(nil)

	if err := sendToStorage(ctx, bucket, output.Path, buf.Bytes(), hash); err != nil {
		return nil, err
	}

	return hash, nil
}

func sendToStorage(ctx context.Context, bucket *storage.BucketHandle, path string, contents, hash []byte) error {
	fpath := fmt.Sprintf("pkg/%s.%x.js", path, hash)
	if err := storeJs(ctx, bucket, bytes.NewBuffer(contents), fpath); err != nil {
		return err
	}
	return nil
}

func storeJs(ctx context.Context, bucket *storage.BucketHandle, reader io.Reader, filename string) error {
	wc := bucket.Object(filename).NewWriter(ctx)
	defer wc.Close()
	wc.ContentType = "application/javascript"
	wc.CacheControl = "public, max-age=31536000"
	if _, err := io.Copy(wc, reader); err != nil {
		return err
	}
	return nil
}

type MainVars struct {
	Prelude string
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
var $mainPkg;
var $load = {};
(function(){
	var count = 0;
	var total = 0;
	var path = "{{ .Path }}";
	var info = {{ .Json }};
	var logger = function(s) { var log = document.getElementById("log"); if (log) { log.innerHTML = s; } }
	var finished = function() {
		logger("Running...");
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
		logger("Loading " + count + " / " + total);
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
	get("https://cdn.jsgo.io/std/prelude.{{ .Prelude }}.js");
	for (var i = 0; i < info.length; i++) {
		get("https://cdn.jsgo.io/" + (info[i].std ? "std" : "pkg") + "/" + info[i].path + "." + info[i].hash + ".js");
	}
})();
`))
