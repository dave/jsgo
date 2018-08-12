package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/dave/frizz/models"
	"github.com/dave/jennifer/jen"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/dave/jsgo/server/frizz/gotypes/convert"
	"github.com/dave/services"
	"github.com/dave/services/builder"
	"github.com/dave/services/constor"
	"github.com/dave/services/deployer"
	"github.com/dave/services/fileserver/gcsfileserver"
	"github.com/dave/services/fileserver/localfileserver"
	"github.com/dave/services/fsutil"
	"github.com/dave/services/session"
	"github.com/dave/services/srcimporter"
	"github.com/dave/stablegob"
	"github.com/gopherjs/gopherjs/compiler"
	"github.com/gopherjs/gopherjs/compiler/prelude"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

func init() {
	gotypes.RegisterTypesStablegob()
}

func main() {

	ctx := context.Background()

	var fileserver services.Fileserver
	if config.LOCAL {
		fileserver = localfileserver.New(config.LocalFileserverTempDir, nil, nil, nil)
	} else {
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Fatal(err)
		}
		defer client.Close()
		fileserver = gcsfileserver.New(client, config.Buckets)
	}

	storer := constor.New(ctx, fileserver, nil, 20)

	archives := map[string]map[bool]*compiler.Archive{}
	source := map[string]map[string]string{}
	packages, err := getStandardLibraryPackages()
	if err != nil {
		log.Fatal(err)
	}
	root, err := getRootFilesystem()
	if err != nil {
		log.Fatal(err)
	}

	if err := StoreSource(ctx, storer, packages, root, source); err != nil {
		log.Fatal(err)
	}

	if err := ScanAndStoreTypes(ctx, storer, packages, root, source); err != nil {
		log.Fatal(err)
	}

	if err := CompileAndStoreJavascript(ctx, storer, packages, root, archives); err != nil {
		log.Fatal(err)
	}

	if err := Prelude(storer); err != nil {
		log.Fatal(err)
	}

	if err := CreateAssetsZip(storer, root, archives); err != nil {
		log.Fatal(err)
	}

	if err := Wasm(storer); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Waiting for storage operations...")
	if err := storer.Wait(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Storage operations finished.")

}

func StoreSource(ctx context.Context, storer *constor.Storer, packages []string, root billy.Filesystem, source map[string]map[string]string) error {
	index := map[string]string{}
	for _, path := range packages {
		source[path] = map[string]string{}
		dir := filepath.Join("goroot", "src", path)
		fis, err := root.ReadDir(dir)
		if err != nil {
			return err
		}
		if len(fis) == 0 {
			panic("0 files in " + path)
		}
		for _, fi := range fis {
			if !strings.HasSuffix(fi.Name(), ".go") {
				// include all .go files
				continue
			}
			err := func() error {
				f, err := root.Open(filepath.Join(dir, fi.Name()))
				if err != nil {
					return err
				}
				defer f.Close()
				b, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}
				source[path][fi.Name()] = string(b)
				return nil
			}()
			if err != nil {
				return err
			}
		}
		if len(source[path]) == 0 {
			panic("0 files in " + path)
		}
		sp := models.SourcePack{
			Path:  path,
			Files: source[path],
		}

		sha := sha1.New()
		buf := &bytes.Buffer{}
		mw := io.MultiWriter(sha, buf)
		if err := json.NewEncoder(mw).Encode(sp); err != nil {
			return err
		}
		hash := fmt.Sprintf("%x", sha.Sum(nil))

		storer.Add(constor.Item{
			Message:   path,
			Bucket:    config.Bucket[config.Pkg],
			Name:      fmt.Sprintf("%s.%s.json", path, hash), // Note: hash is a string
			Contents:  buf.Bytes(),
			Mime:      constor.MimeJson,
			Immutable: true,
			Count:     true,
			Send:      true,
		})

		index[path] = hash
	}

	fmt.Println("Saving source index...")
	/*
		var Source = map[string]string{
			"<path>": "<hash>",
		},
	*/
	f := jen.NewFile("std")
	f.Var().Id("Source").Op("=").Map(jen.String()).String().Values(jen.DictFunc(func(d jen.Dict) {
		for path, hash := range index {
			d[jen.Lit(path)] = jen.Lit(hash)
		}
	}))
	if err := f.Save("../assets/std/source.go"); err != nil {
		return err
	}
	fmt.Println("Done.")
	return nil
}

func ScanAndStoreTypes(ctx context.Context, storer *constor.Storer, stdPackages []string, root billy.Filesystem, source map[string]map[string]string) error {
	fmt.Println("Scanning for objects...")

	s := session.New([]string{}, root, nil, nil, config.ValidExtensions)

	hashes := map[string]string{}

	//stdPackages = []string{"os/user"}

	for _, path := range stdPackages {
		packages := map[string]*types.Package{}
		fset := token.NewFileSet()
		bctx := s.BuildContext(session.DefaultType, "")
		tc := &types.Config{
			FakeImportC: true,
			Importer:    srcimporter.New(bctx, fset, packages),
		}
		ti := &types.Info{
			Types: map[ast.Expr]types.TypeAndValue{},
			Defs:  map[*ast.Ident]types.Object{},
		}

		bp, err := bctx.Import(path, filepath.Join(bctx.GOROOT, "src", path), 0)
		if err != nil {
			return nil
		}

		parsed := []*ast.File{}

		sort.Slice(bp.GoFiles, func(i, j int) bool { return bp.GoFiles[i] < bp.GoFiles[j] })

		for _, fname := range bp.GoFiles {
			dir := filepath.Join(bctx.GOROOT, "src", path)
			fpath := filepath.Join(dir, fname)
			f, err := bctx.OpenFile(fpath)
			if err != nil {
				return err
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				f.Close()
				return err
			}
			f.Close()
			if !strings.HasSuffix(fname, ".go") || strings.HasSuffix(fname, "_test.go") {
				continue
			}
			match, err := bctx.MatchFile(filepath.Join(bctx.GOROOT, "src", path), fname)
			if err != nil {
				return err
			}
			if !match {
				continue
			}
			astfile, err := parser.ParseFile(fset, fpath, b, 0)
			if err != nil {
				return err
			}
			parsed = append(parsed, astfile)
		}

		p, err := tc.Check(path, fset, parsed, ti)
		if err != nil {
			return err
		}

		if p == nil {
			continue
		}

		objects := map[string]map[string]gotypes.Object{}
		for _, name := range p.Scope().Names() {
			v := p.Scope().Lookup(name)
			if v == nil {
				continue
			}
			if !v.Exported() {
				continue
			}

			object := convert.Object(v)
			name := object.Object().Name
			_, file := filepath.Split(fset.File(v.Pos()).Name())

			if objects[file] == nil {
				objects[file] = map[string]gotypes.Object{}
			}
			objects[file][name] = object
		}
		pp := models.ObjectPack{
			Path:    p.Path(),
			Name:    p.Name(),
			Objects: objects,
		}
		sha := sha1.New()
		buf := &bytes.Buffer{}
		mw := io.MultiWriter(sha, buf)
		if err := stablegob.NewEncoder(mw).Encode(pp); err != nil {
			return err
		}
		hash := fmt.Sprintf("%x", sha.Sum(nil))
		hashes[path] = hash
		storer.Add(constor.Item{
			Message:   p.Path(),
			Bucket:    config.Bucket[config.Pkg],
			Name:      fmt.Sprintf("%s.%s.objects.gob", p.Path(), hash), // Note: hash is a string
			Contents:  buf.Bytes(),
			Mime:      constor.MimeBin,
			Immutable: true,
			Count:     true,
			Send:      true,
		})
		fmt.Printf("%s: %d objects\n", path, len(objects))

	}

	fmt.Println("Saving index...")
	/*
		var Objects = map[string]string{
			"<path>": "<hash>",
			...
		}
	*/
	f := jen.NewFile("std")
	f.Var().Id("Objects").Op("=").Map(jen.String()).String().Values(jen.DictFunc(func(d jen.Dict) {
		for path, hash := range hashes {
			d[jen.Lit(path)] = jen.Lit(hash)
		}
	}))
	if err := f.Save("../assets/std/objects.go"); err != nil {
		return err
	}

	fmt.Println("Done")
	return nil
}

func CompileAndStoreJavascript(ctx context.Context, storer *constor.Storer, packages []string, root billy.Filesystem, archives map[string]map[bool]*compiler.Archive) error {
	fmt.Println("Loading...")

	s := session.New(nil, root, nil, nil, nil)

	index := map[string]map[bool]string{}

	buildAndSend := func(min bool) error {
		b := builder.New(s, &builder.Options{Unvendor: true, Initializer: true, Minify: min})

		var minified = " (un-minified)"
		if min {
			minified = " (minified)"
		}

		sent := map[string]bool{}
		for _, p := range packages {
			fmt.Println("Compiling:", p+minified)
			if _, _, err := b.BuildImportPath(ctx, p); err != nil {
				return err
			}

			for _, archive := range b.Archives {
				path := builder.UnvendorPath(archive.ImportPath)
				if sent[path] {
					continue
				}
				fmt.Println("Storing:", path+minified)

				contents, hash, err := builder.GetPackageCode(ctx, archive, min, true)
				if err != nil {
					return err
				}
				storer.Add(constor.Item{
					Message:   path + minified,
					Name:      fmt.Sprintf("%s.%x.js", path, hash),
					Contents:  contents,
					Bucket:    config.Bucket[config.Pkg],
					Mime:      constor.MimeJs,
					Count:     true,
					Immutable: true,
				})

				// NOTE: Archive binaries can change across compiles, so we can't take the hash of the
				// archive file. We use the hash of the JS instead. Could this cause a subtle bug when
				// GopherJS is upgraded? It would need the Archive to change, but the emitted JS to stay
				// exactly the same. Clients would get the old archive. If this is the case, we may have
				// to include a version in the hash. This would mean the the entire cache is invalidated
				// on every version increment instead of just the changed files.

				// We strip most of the contents of the standard library archive files because we don't
				// need to recreate the full JS from these files. Instead we use the stripped archive
				// files in the compile process, and we use the JS files stored on the CDN. Thus we
				// benefit from browser caching.

				// NOTE: After stripping the contents, the archives are actually binary consistent...
				// we could change this to use separate hashes for js and archive... However that would
				// be fiddly so I'm not going to do it now. TODO: revisit this?

				buf := &bytes.Buffer{}
				if err := compiler.WriteArchive(deployer.StripArchive(archive), buf); err != nil {
					return err
				}
				storer.Add(constor.Item{
					Message:   path + " archive" + minified,
					Name:      fmt.Sprintf("%s.%x.ax", path, hash),
					Contents:  buf.Bytes(),
					Bucket:    config.Bucket[config.Pkg],
					Mime:      constor.MimeBin,
					Count:     true,
					Immutable: true,
				})

				if index[path] == nil {
					index[path] = make(map[bool]string, 2)
				}
				index[path][min] = fmt.Sprintf("%x", hash)

				if archives[path] == nil {
					archives[path] = make(map[bool]*compiler.Archive, 2)
				}
				archives[path][min] = archive

				sent[path] = true
			}
		}

		return nil
	}

	if err := buildAndSend(true); err != nil {
		return err
	}
	if err := buildAndSend(false); err != nil {
		return err
	}

	fmt.Println("Saving index...")
	/*
		var Index = map[string]map[bool]string{
			{
				false: "...",
				true: "...",
			},
			...
		}
	*/
	f := jen.NewFile("std")
	f.Var().Id("Index").Op("=").Map(jen.String()).Map(jen.Bool()).String().Values(jen.DictFunc(func(d jen.Dict) {
		for path, p := range index {
			d[jen.Lit(path)] = jen.Values(jen.Dict{
				jen.Lit(false): jen.Lit(p[false]),
				jen.Lit(true):  jen.Lit(p[true]),
			})
		}
	}))
	if err := f.Save("../assets/std/index.go"); err != nil {
		return err
	}
	fmt.Println("Done.")

	return nil
}

func CreateAssetsZip(storer *constor.Storer, root billy.Filesystem, archives map[string]map[bool]*compiler.Archive) error {
	fmt.Println("Loading...")
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	fmt.Println("Zipping...")
	var compress func(billy.Filesystem, string) error
	compress = func(fs billy.Filesystem, dir string) error {

		destinationDir := dir
		unvendoredImportPath := builder.UnvendorPath(strings.TrimPrefix(filepath.ToSlash(dir), "/goroot/src/"))
		if archives[unvendoredImportPath] != nil {
			// only unvendor the packages that correspond to archives
			destinationDir = filepath.Join("/", "goroot", "src", unvendoredImportPath)
		}

		fis, err := fs.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, fi := range fis {
			fpath := filepath.Join(dir, fi.Name())
			fpathDestination := filepath.Join(destinationDir, fi.Name())
			if fi.IsDir() {
				if err := compress(fs, fpath); err != nil {
					return err
				}
				continue
			}
			if strings.HasSuffix(fpath, "_test.go") {
				continue
			}
			z, err := w.Create(fpathDestination)
			if err != nil {
				return err
			}
			f, err := fs.Open(fpath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(z, f); err != nil {
				f.Close()
				return err
			}
			f.Close()
			w.Flush()
		}
		return nil
	}
	if err := compress(root, "/"); err != nil {
		return err
	}
	if err := compress(osfs.New("../assets/static"), "/"); err != nil {
		return err
	}

	z, err := w.Create("/archives.gob")
	if err != nil {
		return err
	}
	if err := gob.NewEncoder(z).Encode(archives); err != nil {
		return err
	}

	if err := w.Flush(); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join("../assets/", config.AssetsFilename), buf.Bytes(), 0777); err != nil {
		return err
	}

	storer.Add(constor.Item{
		Message:   "assets",
		Name:      config.AssetsFilename,
		Contents:  buf.Bytes(),
		Bucket:    config.Bucket[config.Pkg],
		Mime:      constor.MimeZip,
		Count:     false,
		Immutable: false,
	})

	return nil
}

func Wasm(storer *constor.Storer) error {
	store := func(message, fpath string) ([]byte, error) {
		buf := &bytes.Buffer{}
		sha := sha1.New()
		w := io.MultiWriter(buf, sha)
		f, err := os.Open(fpath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			return nil, err
		}
		hash := sha.Sum(nil)
		storer.Add(constor.Item{
			Message:   message,
			Name:      fmt.Sprintf("wasm_exec.%x.js", hash),
			Contents:  buf.Bytes(),
			Bucket:    config.Bucket[config.Pkg],
			Mime:      constor.MimeJs,
			Count:     true,
			Immutable: true,
		})
		return hash, nil
	}
	hashMax, err := store("wasm", "../assets/static/wasm/wasm_exec.js")
	if err != nil {
		return err
	}
	hashMin, err := store("wasm (minified)", "../assets/static/wasm/wasm_exec.min.js")
	if err != nil {
		return err
	}

	/*
		var Wasm = map[bool]string{
			false: "...",
			true: "...",
		}
	*/
	f := jen.NewFile("std")
	f.Var().Id("Wasm").Op("=").Map(jen.Bool()).String().Values(jen.Dict{
		jen.Lit(false): jen.Lit(fmt.Sprintf("%x", hashMax)),
		jen.Lit(true):  jen.Lit(fmt.Sprintf("%x", hashMin)),
	})
	if err := f.Save("../assets/std/wasm.go"); err != nil {
		return err
	}
	return nil
}

func Prelude(storer *constor.Storer) error {
	store := func(suffix, contents string) ([]byte, error) {
		b := []byte(contents)
		s := sha1.New()
		if _, err := s.Write(b); err != nil {
			return nil, err
		}
		hash := s.Sum(nil)
		storer.Add(constor.Item{
			Message:   "prelude" + suffix,
			Name:      fmt.Sprintf("prelude.%x.js", hash),
			Contents:  b,
			Bucket:    config.Bucket[config.Pkg],
			Mime:      constor.MimeJs,
			Count:     true,
			Immutable: true,
		})
		return hash, nil
	}
	hashMin, err := store(" (minified)", prelude.Minified+jsGoPrelude)
	if err != nil {
		return err
	}
	hashMax, err := store(" (non-minified)", prelude.Prelude+jsGoPrelude)
	if err != nil {
		return err
	}

	/*
		var Prelude = map[bool]string{
			false: "...",
			true: "...",
		}
	*/
	f := jen.NewFile("std")
	f.Var().Id("Prelude").Op("=").Map(jen.Bool()).String().Values(jen.Dict{
		jen.Lit(false): jen.Lit(fmt.Sprintf("%x", hashMax)),
		jen.Lit(true):  jen.Lit(fmt.Sprintf("%x", hashMin)),
	})
	if err := f.Save("../assets/std/prelude.go"); err != nil {
		return err
	}
	return nil
}

// Add dummy package prelude to the loader so prelude can be loaded like a package
/*
const jsGoPrelude = `$load.prelude=function(){};$done();`
*/
const jsGoPrelude = `$load.prelude=function(){};`

func getRootFilesystem() (billy.Filesystem, error) {
	root := memfs.New()
	if err := fsutil.Copy(root, "/goroot/src", osfs.New(build.Default.GOROOT), "/src"); err != nil {
		return nil, err
	}
	if err := fsutil.Copy(root, "/goroot/src/github.com/gopherjs/gopherjs/js", osfs.New(build.Default.GOPATH), "/src/github.com/gopherjs/gopherjs/js"); err != nil {
		return nil, err
	}
	if err := fsutil.Copy(root, "/goroot/src/github.com/gopherjs/gopherjs/nosync", osfs.New(build.Default.GOPATH), "/src/github.com/gopherjs/gopherjs/nosync"); err != nil {
		return nil, err
	}
	return root, nil
}

func getStandardLibraryPackages() ([]string, error) {
	cmd := exec.Command("go", "list", "./...")
	cmd.Env = []string{
		fmt.Sprintf("GOPATH=%s", build.Default.GOPATH),
		fmt.Sprintf("GOROOT=%s", build.Default.GOROOT),
	}
	cmd.Dir = filepath.Join(build.Default.GOROOT, "src")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	all := strings.Split(strings.TrimSpace(string(b)), "\n")
	excluded := map[string]bool{
		"builtin":                true,
		"internal/cpu":           true,
		"net/http/pprof":         true,
		"plugin":                 true,
		"runtime/cgo":            true,
		"os/signal/internal/pty": true,
	}
	var filtered []string
	for _, p := range all {
		if excluded[p] {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered, nil
}

func printDir(fs billy.Filesystem, dir string) error {
	fis, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		fpath := filepath.Join(dir, fi.Name())
		fmt.Println(fpath)
		if fi.IsDir() {
			if err := printDir(fs, fpath); err != nil {
				return err
			}
		}
	}
	return nil
}
