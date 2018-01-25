package compiler

import (
	"go/parser"
	"go/types"
	"os"

	"cloud.google.com/go/storage"

	"sync"

	"encoding/gob"
	"fmt"

	"io"

	"strings"

	"crypto/sha1"
	"encoding/json"

	"text/template"

	"bytes"

	"context"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/common"
	"github.com/dave/jsgo/config"
	"github.com/gopherjs/gopherjs/compiler"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/loader"
	"gopkg.in/src-d/go-billy.v4"
)

const VERSION = "9"

func New(fs billy.Filesystem) *Cache {
	c := &Cache{
		fs:        fs,
		archivesM: &sync.RWMutex{},
		archives:  make(map[string]*compiler.Archive),
	}
	return c
}

type Cache struct {
	fs         billy.Filesystem
	archivesM  *sync.RWMutex
	archives   map[string]*compiler.Archive
	prog       *loader.Program
	ordered    []*ArchiveInfo
	info       map[string]*ArchiveInfo
	mainJs     []byte
	mainJsHash []byte
}

func (c *Cache) Hash(path string) []byte {
	return c.mainJsHash
}

type ArchiveInfo struct {
	Path     string
	Standard bool
	Archive  *compiler.Archive
	Hash     []byte
	FullHash []byte
}

func (c *Cache) Store(ctx context.Context, path string, logger io.Writer) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket("jsgo")
	for _, a := range c.ordered {
		if a.Standard {
			continue
		}
		fmt.Fprintf(logger, "Storing %s\n", a.Path)
		if err := storeArchive(ctx, bucket, a); err != nil {
			return err
		}
	}
	if err := c.storeMainJs(ctx, bucket, path); err != nil {
		return err
	}
	return nil
}

func (c *Cache) storeMainJs(ctx context.Context, bucket *storage.BucketHandle, path string) error {

	min := ".min"
	if config.DEV {
		min = ""
	}

	fname := fmt.Sprintf("js/%s/main.%x%s.js", path, c.mainJsHash, min)
	if err := storeJs(ctx, bucket, bytes.NewBuffer(c.mainJs), fname); err != nil {
		return err
	}
	return nil
}

func storeArchive(ctx context.Context, bucket *storage.BucketHandle, archive *ArchiveInfo) error {
	buf := &bytes.Buffer{}
	if err := writeArchive(buf, archive.Archive); err != nil {
		return err
	}

	min := ".min"
	if config.DEV {
		min = ""
	}

	fname := fmt.Sprintf("js/%s/package.%x%s.js", archive.Path, archive.FullHash, min)

	if err := storeJs(ctx, bucket, buf, fname); err != nil {
		return err
	}
	return nil
}

func storeStandard(ctx context.Context, bucket *storage.BucketHandle, path string, archive *compiler.Archive) error {
	buf := &bytes.Buffer{}
	if err := writeArchive(buf, archive); err != nil {
		return err
	}

	min := ".min"
	if config.DEV {
		min = ""
	}

	fname := fmt.Sprintf("std/%s/package%s.js", path, min)

	if err := storeJs(ctx, bucket, buf, fname); err != nil {
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

func (c *Cache) Compile(path string, logger io.Writer, hashes map[string]string) error {

	conf := loader.Config{}
	conf.Import(path)
	conf.ParserMode = parser.ParseComments
	conf.Build = common.NewBuildContext(c.fs, true)

	prog, err := conf.Load()
	if err != nil {
		return err
	}
	c.prog = prog

	var importContext *compiler.ImportContext
	importContext = &compiler.ImportContext{
		Packages: make(map[string]*types.Package),
		Import: func(path string) (*compiler.Archive, error) {

			// find in local cache
			if a, ok := c.getArchive(path); ok {
				return a, nil
			}

			pi := c.prog.Package(path)
			importContext.Packages[path] = pi.Pkg

			// find in standard library cache
			a, err := openStaticArchive(path)
			if err != nil {
				return nil, err
			}

			if a != nil {
				c.setArchive(path, a)
				return a, nil
			}

			fmt.Fprintf(logger, "Compiling %s\n", path)

			// compile package
			minify := !config.DEV
			a, err = compiler.Compile(path, pi.Files, c.prog.Fset, importContext, minify)
			if err != nil {
				return nil, err
			}
			c.setArchive(path, a)
			return a, nil
		},
	}

	if _, err := importContext.Import(path); err != nil {
		return err
	}

	if err := c.orderImports(path); err != nil {
		return err
	}

	if err := c.assignHashes(path, hashes); err != nil {
		return err
	}

	//for path, archive := range c.info {
	//	fmt.Println(path, fmt.Sprintf("%x", archive.Hash), archive.Revision)
	//}

	if err := c.renderMain(path); err != nil {
		return err
	}

	return nil
}

type MainVars struct {
	Min  bool
	Path string
	Json string
}

type PkgJson struct {
	Path string `json:"path"`
	Hash string `json:"hash,omitempty"`
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
load("https://storage.googleapis.com/jsgo/std/prelude.js");
$pkgs.forEach(function(pkg){
	var hash = pkg.hash ? "." + pkg.hash : "";
	var dir = pkg.hash ? "js" : "std";
	var min = $min ? ".min" : "";
	load("https://storage.googleapis.com/jsgo/" + dir + "/" + pkg.path + "/package" + hash + min + ".js");
});
`))

func (c *Cache) renderMain(path string) error {
	var pkgs []PkgJson
	for _, a := range c.ordered {
		if a.Archive == nil {
			continue
		}
		p := PkgJson{
			Path: a.Path,
			Hash: fmt.Sprintf("%x", a.FullHash),
		}
		pkgs = append(pkgs, p)
	}
	b, err := json.Marshal(pkgs)
	if err != nil {
		return err
	}
	m := MainVars{
		Min:  !config.DEV,
		Path: path,
		Json: string(b),
	}
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, m); err != nil {
		return err
	}
	c.mainJs = buf.Bytes()
	s := sha1.New()
	if _, err := s.Write(buf.Bytes()); err != nil {
		return err
	}
	c.mainJsHash = s.Sum(nil)
	return nil
}

// TODO: automate this? Add GopherJS version?
const STDLIB_HASH = "go version go1.9 darwin/amd64"

func (c *Cache) assignHashes(path string, repoHashes map[string]string) error {
	var getHash func(path string) ([]byte, error)
	getHash = func(path string) ([]byte, error) {
		archive, ok := c.info[path]
		if !ok {
			return nil, nil
		}
		if archive.Standard {
			return nil, nil
		}
		pi := c.prog.Package(path)
		sig := Sig{
			Version: VERSION,
			Path:    path,
			Name:    pi.Pkg.Name(),
			Defs:    map[string]bool{},
			Imports: map[string][]byte{},
		}
		for _, p := range pi.Pkg.Imports() {
			h, err := getHash(p.Path())
			if err != nil {
				return nil, err
			}
			sig.Imports[p.Path()] = h
		}
		for id, ob := range pi.Defs {
			if !id.IsExported() {
				continue
			}
			sig.Defs[ob.String()] = true
		}
		sha := sha1.New()
		if err := json.NewEncoder(sha).Encode(sig); err != nil {
			return nil, err
		}
		archive.Hash = sha.Sum(nil)
		revision, ok := repoHashes[path]
		if !ok {
			return nil, fmt.Errorf("can't find repo revision hash for %s", path)
		}
		fmt.Fprint(sha, "\n", revision)
		archive.FullHash = sha.Sum(nil)
		return archive.Hash, nil
	}
	if _, err := getHash(path); err != nil {
		return err
	}
	return nil
}

type Sig struct {
	Version string
	Path    string
	Name    string
	Defs    map[string]bool
	Imports map[string][]byte
}

func (c *Cache) getArchive(path string) (*compiler.Archive, bool) {
	c.archivesM.RLock()
	defer c.archivesM.RUnlock()
	a, ok := c.archives[path]
	return a, ok
}

func (c *Cache) setArchive(path string, a *compiler.Archive) {
	c.archivesM.Lock()
	defer c.archivesM.Unlock()
	c.archives[path] = a
}

func openStaticArchive(path string) (*compiler.Archive, error) {

	var filename string
	if config.DEV {
		filename = fmt.Sprintf("pkg/%s.a", path)
	} else {
		filename = fmt.Sprintf("pkg_min/%s.a", path)
	}

	f, err := assets.Assets.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	defer f.Close()

	archive := new(compiler.Archive)
	if err := gob.NewDecoder(f).Decode(archive); err != nil {
		return nil, errors.WithStack(err)
	}
	return archive, nil

}

func (c *Cache) orderImports(path string) error {

	getInfo := func(path string) (*ArchiveInfo, error) {
		pkg, err := openStaticArchive(path)
		if err != nil {
			return nil, err
		}
		return &ArchiveInfo{
			Path:     path,
			Standard: true,
			Archive:  pkg,
		}, nil
	}

	var err error
	var pkgJs, pkgNosync, pkgRuntime, pkgRuntimeSys *ArchiveInfo

	if pkgJs, err = getInfo("github.com/gopherjs/gopherjs/js"); err != nil {
		return err
	}
	if pkgNosync, err = getInfo("github.com/gopherjs/gopherjs/nosync"); err != nil {
		return err
	}
	if pkgRuntime, err = getInfo("runtime"); err != nil {
		return err
	}
	if pkgRuntimeSys, err = getInfo("runtime/internal/sys"); err != nil {
		return err
	}

	c.info = map[string]*ArchiveInfo{
		"github.com/gopherjs/gopherjs/js":     pkgJs,
		"github.com/gopherjs/gopherjs/nosync": pkgNosync,
		"runtime/internal/sys":                pkgRuntimeSys,
		"runtime":                             pkgRuntime,
	}
	c.ordered = []*ArchiveInfo{pkgJs, pkgNosync, pkgRuntimeSys, pkgRuntime}

	done := map[string]struct{}{
		"github.com/gopherjs/gopherjs/js":     {},
		"github.com/gopherjs/gopherjs/nosync": {},
		"runtime":              {},
		"runtime/internal/sys": {},
	}

	var orderImports func(string) error
	orderImports = func(path string) error {

		done[path] = struct{}{}

		pi := c.prog.Package(path)

		if len(pi.Files) == 0 {
			return nil
		}

		std := strings.HasPrefix(c.prog.Fset.File(pi.Files[0].Pos()).Name(), "/goroot/")

		for _, child := range pi.Pkg.Imports() {
			if _, ok := done[child.Path()]; ok {
				continue
			}
			if err := orderImports(child.Path()); err != nil {
				return err
			}
		}

		var arch *compiler.Archive

		if std {
			var err error
			arch, err = openStaticArchive(path)
			if err != nil {
				return err
			}
			if arch == nil {
				// some packages e.g. internal/cpu, runtime/internal/atomic etc. are not needed by
				// gopherjs, so we don't have them in the static archive cache
				return nil
			}
		} else {
			var ok bool
			arch, ok = c.archives[path]
			if !ok {
				return fmt.Errorf("can't find import %s", path)
			}
		}

		info := &ArchiveInfo{Path: path, Standard: std, Archive: arch}
		c.ordered = append(c.ordered, info)
		c.info[path] = info
		return nil
	}
	if err := orderImports(path); err != nil {
		return err
	}
	return nil
}

func writeArchive(w io.Writer, archive *compiler.Archive) error {

	selection := make(map[*compiler.Decl]struct{})
	for _, d := range archive.Declarations {
		selection[d] = struct{}{}
	}

	fmt.Fprintf(w, `$initialisers["%s"] = function () {`, archive.ImportPath)

	if err := compiler.WritePkgCode(archive, selection, false, &compiler.SourceMapFilter{Writer: w}); err != nil {
		return err
	}

	fmt.Fprint(w, "};")

	return nil
}
