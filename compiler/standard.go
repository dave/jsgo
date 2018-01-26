package compiler

import (
	"fmt"

	"cloud.google.com/go/storage"

	"os"

	"context"

	"strings"

	"bytes"

	"github.com/dave/jsgo/assets"
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

	done := map[string]bool{}
	var storeArchives func(path string, minify bool) error
	storeArchives = func(path string, minify bool) error {
		root := "pkg/"
		if minify {
			root = "pkg_min/"
		}
		if done[path] {
			return nil
		}
		done[path] = true
		a, err := openStaticArchive(path)
		if err != nil {
			return err
		}
		if a != nil {
			fmt.Println(path)
			if err := storeStandard(ctx, bucket, path, a); err != nil {
				return err
			}
		}
		dir := fmt.Sprint(root, path)
		s, err := assets.Assets.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !s.IsDir() {
			return nil
		}
		fis, err := assets.Assets.ReadDir(dir)
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
			if err := storeArchives(newPath, minify); err != nil {
				return err
			}
		}
		return nil
	}
	fmt.Println("Sending non-minified JS to GCS...")
	if err := storeArchives("", false); err != nil {
		return err
	}
	done = map[string]bool{}
	fmt.Println("Sending minified JS to GCS...")
	if err := storeArchives("", true); err != nil {
		return err
	}
	return nil
}
