package main

import (
	"context"
	"log"

	"cloud.google.com/go/storage"

	"path/filepath"

	"os"

	"io/ioutil"

	"archive/zip"
	"io"

	"fmt"

	"bytes"

	"github.com/dave/patsy"
	"github.com/dave/patsy/vos"
)

func main() {

	root, err := patsy.Dir(vos.Os(), "github.com/dave/jsgo/assets/static")
	if err != nil {
		log.Fatalln(err)
	}

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	fmt.Println("Zipping...")
	var compress func(dir string) error
	compress = func(dir string) error {
		fis, err := ioutil.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, fi := range fis {
			fpath := filepath.Join(dir, fi.Name())
			if fi.IsDir() {
				if err := compress(fpath); err != nil {
					return err
				}
				continue
			}
			rel, err := filepath.Rel(root, fpath)
			if err != nil {
				return err
			}
			z, err := w.Create(rel)
			if err != nil {
				return err
			}
			f, err := os.Open(fpath)
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
	if err := compress(root); err != nil {
		log.Fatalln(err)
	}
	w.Flush()
	w.Close()

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	bucket := client.Bucket("jsgo")

	fmt.Println("Storing local copy...")
	if err := ioutil.WriteFile(filepath.Join(root, "../assets.zip"), buf.Bytes(), 0666); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Uploading to GCS...")
	if err := storeZip(ctx, bucket, buf, "assets.zip"); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Done!")

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
