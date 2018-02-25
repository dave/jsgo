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

	"flag"

	"io/ioutil"

	"crypto/sha1"

	"github.com/dave/jennifer/jen"
	"github.com/dave/jsgo/builder"
	"github.com/dave/jsgo/builder/fscopy"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/compile"
	"github.com/gopherjs/gopherjs/compiler/prelude"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
)

func main() {

	flag.Parse()
	command := flag.Arg(0)

	switch command {
	case "js":
		if err := Js(); err != nil {
			log.Fatal(err)
		}
	case "src":
		if err := Src(); err != nil {
			log.Fatal(err)
		}
	case "prelude":
		if err := Prelude(); err != nil {
			log.Fatal(err)
		}
	}

}

func Src() error {
	fmt.Println("Loading...")
	root, err := getRootFilesystem()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	fmt.Println("Zipping...")
	var compress func(billy.Filesystem, string) error
	compress = func(fs billy.Filesystem, dir string) error {
		fis, err := fs.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, fi := range fis {
			fpath := filepath.Join(dir, fi.Name())
			if fi.IsDir() {
				if err := compress(fs, fpath); err != nil {
					return err
				}
				continue
			}
			if strings.HasSuffix(fpath, "_test.go") {
				continue
			}
			z, err := w.Create(fpath)
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
	if err := w.Flush(); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join("./assets/", config.AssetsFilename), buf.Bytes(), 0777); err != nil {
		return err
	}

	fmt.Println("Uploading...")
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(config.PkgBucket)
	if err := storeZip(ctx, bucket, bytes.NewBuffer(buf.Bytes()), config.AssetsFilename); err != nil {
		return err
	}
	fmt.Println("Done.")

	return nil
}

// Add dummy package prelude to the loader so prelude can be loaded like a package
const jsGoPrelude = `$load.prelude=function(){};`

func Prelude() error {
	fmt.Println("Storing prelude...")

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket(config.PkgBucket)

	store := func(contents string) ([]byte, error) {
		b := []byte(contents)
		s := sha1.New()
		if _, err := s.Write(b); err != nil {
			return nil, err
		}
		hash := s.Sum(nil)
		fname := fmt.Sprintf("prelude.%x.js", hash)
		if err := storeJs(ctx, bucket, bytes.NewBuffer(b), fname); err != nil {
			return nil, nil
		}
		return hash, nil
	}
	hashMin, err := store(prelude.PreludeMinified + jsGoPrelude)
	if err != nil {
		return err
	}
	hashMax, err := store(prelude.Prelude + jsGoPrelude)
	if err != nil {
		return err
	}

	/*
		const (
			PreludeMin = "..."
			PreludeMax = "..."
		)
	*/
	f := jen.NewFile("std")
	f.Const().Defs(
		jen.Id("PreludeMin").Op("=").Lit(fmt.Sprintf("%x", hashMin)),
		jen.Id("PreludeMax").Op("=").Lit(fmt.Sprintf("%x", hashMax)),
	)
	if err := f.Save("./builder/std/prelude.go"); err != nil {
		return err
	}
	fmt.Println("Done.")
	return nil
}

func Js() error {
	fmt.Println("Loading...")
	packages, err := getStandardLibraryPackages()
	if err != nil {
		return err
	}
	rootfs, err := getRootFilesystem()
	if err != nil {
		return err
	}
	pathfs := memfs.New()
	archivefs := memfs.New()

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	storer := compile.NewStorer(ctx, client, nil, 20)

	output := map[string]*builder.PackageHash{}

	buildAndSend := func(min bool) error {
		session := builder.NewSession(&builder.Options{
			Root:        rootfs,
			Path:        pathfs,
			Temporary:   archivefs,
			Unvendor:    true,
			Initializer: true,
			Minify:      min,
		})

		var minified = " (un-minified)"
		if min {
			minified = " (minified)"
		}

		sent := map[string]bool{}
		for _, p := range packages {
			fmt.Println("Compiling:", p+minified)
			if _, _, err := session.BuildImportPath(ctx, p); err != nil {
				return err
			}

			for _, archive := range session.Archives {
				path := archive.ImportPath
				if sent[path] {
					continue
				}
				fmt.Println("Storing:", path+minified)
				contents, hash, err := builder.GetPackageCode(ctx, archive, min, true)
				if err != nil {
					return err
				}
				storer.AddJs(path+minified, fmt.Sprintf("%s.%x.js", path, hash), contents)
				if output[path] == nil {
					output[path] = &builder.PackageHash{}
				}
				if min {
					output[path].HashMin = fmt.Sprintf("%x", hash)
				} else {
					output[path].HashMax = fmt.Sprintf("%x", hash)
				}
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

	fmt.Println("Compile finished. Waiting for storage operations...")
	storer.Wait()
	fmt.Println("Storage operations finished.")
	if storer.Err != nil {
		return err
	}

	fmt.Println("Saving index...")
	/*
		var Index = map[string]builder.PackageHash{
			{
				HashMax: "...",
				HashMin: "...",
			},
			...
		}
	*/
	f := jen.NewFile("std")
	f.Var().Id("Index").Op("=").Map(jen.String()).Qual("github.com/dave/jsgo/builder", "PackageHash").Values(jen.DictFunc(func(d jen.Dict) {
		for path, p := range output {
			d[jen.Lit(path)] = jen.Values(jen.Dict{
				jen.Id("HashMax"): jen.Lit(p.HashMax),
				jen.Id("HashMin"): jen.Lit(p.HashMin),
			})
		}
	}))
	if err := f.Save("./builder/std/index.go"); err != nil {
		return err
	}
	fmt.Println("Done.")

	return nil
}

func sendToStorage(ctx context.Context, bucket *storage.BucketHandle, path string, contents, hash []byte) error {
	fpath := fmt.Sprintf("%s.%x.js", path, hash)
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

func storeZip(ctx context.Context, bucket *storage.BucketHandle, reader io.Reader, filename string) error {
	wc := bucket.Object(filename).NewWriter(ctx)
	defer wc.Close()
	wc.ContentType = "application/zip"
	if _, err := io.Copy(wc, reader); err != nil {
		return err
	}
	return nil
}

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
