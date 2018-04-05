package main

import (
	"archive/zip"
	"context"
	"fmt"
	"go/build"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"

	"bytes"
	"io"

	"io/ioutil"

	"crypto/sha1"

	"encoding/gob"

	"github.com/dave/jennifer/jen"
	"github.com/dave/jsgo/builder"
	"github.com/dave/jsgo/builder/fscopy"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/session"
	"github.com/gopherjs/gopherjs/compiler"
	"github.com/gopherjs/gopherjs/compiler/prelude"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

func main() {

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	storer := compile.NewStorer(ctx, client, nil, 20)

	index := map[string]map[bool]string{}
	archives := map[string]map[bool]*compiler.Archive{}
	packages, err := getStandardLibraryPackages()
	if err != nil {
		log.Fatal(err)
	}
	root, err := getRootFilesystem()
	if err != nil {
		log.Fatal(err)
	}

	if err := CompileAndStoreJavascript(ctx, storer, packages, root, index, archives); err != nil {
		log.Fatal(err)
	}

	if err := Prelude(storer); err != nil {
		log.Fatal(err)
	}

	if err := CreateAssetsZip(storer, root, archives); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Waiting for storage operations...")
	storer.Wait()
	fmt.Println("Storage operations finished.")
	if storer.Err != nil {
		log.Fatal(err)
	}

}

func CompileAndStoreJavascript(ctx context.Context, storer *compile.Storer, packages []string, root billy.Filesystem, index map[string]map[bool]string, archives map[string]map[bool]*compiler.Archive) error {
	fmt.Println("Loading...")

	s, err := session.New(nil, nil, root)
	if err != nil {
		return err
	}

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
				storer.AddJs(path+minified, fmt.Sprintf("%s.%x.js", path, hash), contents)

				// NOTE: Archive binaries can change across compiles, so we can't take the hash of the
				// archive file. We use the hash of the JS instead. Could this cause a subtle bug when
				// GopherJS is upgraded? It would need the Archive to change, but the emitted JS to stay
				// exactly the same. Clients would get the old archive. If this is the case, we may have
				// to include a version in the hash. This would mean the the entire cache is invalidated
				// on every version increment instead of just the changed files.
				buf := &bytes.Buffer{}
				if err := compiler.WriteArchive(archive, buf); err != nil {
					return err
				}
				storer.AddArchive(path+" archive"+minified, fmt.Sprintf("%s.%x.a", path, hash), buf.Bytes())

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
	if err := f.Save("./builder/std/index.go"); err != nil {
		return err
	}
	fmt.Println("Done.")

	return nil
}

func CreateAssetsZip(storer *compile.Storer, root billy.Filesystem, archives map[string]map[bool]*compiler.Archive) error {
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
	if err := compress(osfs.New("./assets/static"), "/"); err != nil {
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
	if err := ioutil.WriteFile(filepath.Join("./assets/", config.AssetsFilename), buf.Bytes(), 0777); err != nil {
		return err
	}

	storer.AddZip("assets", config.AssetsFilename, buf.Bytes())

	return nil
}

func Prelude(storer *compile.Storer) error {
	store := func(suffix, contents string) ([]byte, error) {
		b := []byte(contents)
		s := sha1.New()
		if _, err := s.Write(b); err != nil {
			return nil, err
		}
		hash := s.Sum(nil)
		storer.AddJs("prelude"+suffix, fmt.Sprintf("prelude.%x.js", hash), b)
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
	if err := f.Save("./builder/std/prelude.go"); err != nil {
		return err
	}
	return nil
}

// Add dummy package prelude to the loader so prelude can be loaded like a package
const jsGoPrelude = `$load.prelude=function(){};`

func getRootFilesystem() (billy.Filesystem, error) {
	root := memfs.New()
	if err := fscopy.Copy("/src", "/goroot/src", osfs.New(build.Default.GOROOT), root); err != nil {
		return nil, err
	}
	if err := fscopy.Copy("/src/github.com/gopherjs/gopherjs/js", "/goroot/src/github.com/gopherjs/gopherjs/js", osfs.New(build.Default.GOPATH), root); err != nil {
		return nil, err
	}
	if err := fscopy.Copy("/src/github.com/gopherjs/gopherjs/nosync", "/goroot/src/github.com/gopherjs/gopherjs/nosync", osfs.New(build.Default.GOPATH), root); err != nil {
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
