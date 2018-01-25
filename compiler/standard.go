package compiler

import (
	"fmt"

	"cloud.google.com/go/storage"

	"os"

	"context"

	"strings"

	"bytes"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/config"
	"github.com/gopherjs/gopherjs/compiler/prelude"
)

func (c *Cache) CompileStdLib() error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	bucket := client.Bucket("jsgo")

	fmt.Println("Storing prelude...")
	if err := storeJs(ctx, bucket, bytes.NewBuffer([]byte(prelude.Prelude)), "std/prelude.js"); err != nil {
		return nil
	}

	root := "pkg/"
	if !config.DEV {
		root = "pkg_min/"
	}
	done := map[string]bool{}
	var storeArchives func(path string) error
	storeArchives = func(path string) error {
		if done[path] {
			return nil
		}
		done[path] = true
		a, err := openStaticArchive(path)
		if err != nil {
			return err
		}
		if a != nil {
			fmt.Println("Storing", path)
			if err := storeStandard(ctx, bucket, path, a); err != nil {
				return err
			}
		}
		dir := fmt.Sprint(root, path)
		f, err := assets.Assets.Open(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		s, err := f.Stat()
		if err != nil {
			return err
		}
		if !s.IsDir() {
			return nil
		}
		fis, err := f.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fis {
			pkg := strings.TrimSuffix(fi.Name(), ".a")
			var newPath string
			if path == "" {
				newPath = pkg
			} else {
				newPath = path + "/" + pkg
			}
			if err := storeArchives(newPath); err != nil {
				return err
			}
		}
		return nil
	}
	if err := storeArchives(""); err != nil {
		return err
	}
	return nil
}
