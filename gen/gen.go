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
	"github.com/dave/jsgo/builder/std"
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
			z, err := w.Create(fpath)
			if err != nil {
				return err
			}
			if strings.HasSuffix(fpath, "_test.go") {
				continue
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
	if err := ioutil.WriteFile("./assets/assets.zip", buf.Bytes(), 0777); err != nil {
		return err
	}

	fmt.Println("Uploading...")
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket("jsgo")
	if err := storeZip(ctx, bucket, bytes.NewBuffer(buf.Bytes()), "assets.zip"); err != nil {
		return err
	}
	fmt.Println("Done.")

	return nil
}

func Prelude() error {
	fmt.Println("Storing prelude...")
	b := []byte(prelude.Prelude)
	s := sha1.New()
	if _, err := s.Write(b); err != nil {
		return err
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket("jsgo")

	hash := fmt.Sprintf("%x", s.Sum(nil))

	fname := fmt.Sprintf("sys/prelude.%s.js", hash)
	if err := storeJs(ctx, bucket, bytes.NewBuffer(b), fname); err != nil {
		return nil
	}

	/*
		const PreludeHash = "..."
	*/
	f := jen.NewFile("std")
	f.Const().Id("PreludeHash").Op("=").Lit(hash)
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
	bucket := client.Bucket("jsgo")

	sessionMin := builder.NewSession(&builder.Options{
		Root:        rootfs,
		Path:        pathfs,
		Temporary:   archivefs,
		Unvendor:    true,
		Initializer: true,
		Minify:      true,
	})
	sessionMax := builder.NewSession(&builder.Options{
		Root:        rootfs,
		Path:        pathfs,
		Temporary:   archivefs,
		Unvendor:    true,
		Initializer: true,
		Minify:      false,
	})

	for _, p := range packages {
		fmt.Println("Compiling", p)
		if _, err := sessionMin.BuildImportPath(p); err != nil {
			return err
		}
		if _, err := sessionMax.BuildImportPath(p); err != nil {
			return err
		}
	}

	output := map[string]std.Package{}
	for key := range sessionMin.Archives {
		archiveMin := sessionMin.Archives[key]
		archiveMax := sessionMax.Archives[key]
		if archiveMin == nil || archiveMax == nil || archiveMin.ImportPath != archiveMax.ImportPath {
			return fmt.Errorf("archives %s don't match!", key)
		}
		path := archiveMin.ImportPath
		fmt.Println("Storing", path)
		contentsMin, hashMin, err := builder.GetPackageCode(archiveMin, true, true)
		if err != nil {
			return err
		}
		contentsMax, hashMax, err := builder.GetPackageCode(archiveMax, false, true)
		if err != nil {
			return err
		}
		if err := sendToStorage(ctx, bucket, path, contentsMin, hashMin, true); err != nil {
			return err
		}
		if err := sendToStorage(ctx, bucket, path, contentsMax, hashMax, false); err != nil {
			return err
		}
		output[path] = std.Package{
			HashMin: fmt.Sprintf("%x", hashMin),
			HashMax: fmt.Sprintf("%x", hashMax),
		}
	}

	/*
		var Index = map[string]Package{
			{
				HashMax: "...",
				HashMin: "...",
			},
			...
		}
	*/
	f := jen.NewFile("std")
	f.Var().Id("Index").Op("=").Map(jen.String()).Id("Package").Values(jen.DictFunc(func(d jen.Dict) {
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

func sendToStorage(ctx context.Context, bucket *storage.BucketHandle, path string, contents, hash []byte, minified bool) error {
	min := ""
	if minified {
		min = ".min"
	}
	fpath := fmt.Sprintf("sys/%s.%x%s.js", path, hash, min)
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
		"builtin":        true,
		"internal/cpu":   true,
		"net/http/pprof": true,
		"plugin":         true,
		"runtime/cgo":    true,
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
