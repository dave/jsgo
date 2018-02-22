package getter

import (
	"os"
	"testing"

	"fmt"
	"path/filepath"

	"context"

	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

type debugWriter struct {
}

func (debugWriter) Write(p []byte) (n int, err error) {
	fmt.Printf("%#v\n", string(p))
	return len(p), nil
}

func TestClone(t *testing.T) {
	store, err := filesystem.NewStorage(memfs.New())
	if err != nil {
		t.Fatal(err)
	}
	_, err = git.Clone(store, memfs.New(), &git.CloneOptions{
		//URL: "https://go.googlesource.com/image",
		URL:               "https://github.com/dave/jstest",
		SingleBranch:      true,
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	// "https://go.googlesource.com/image":
	/*
		Counting objects: 264, done
		Finding sources: 100% (264/264)
		Total 264 (delta 27), reused 250 (delta 27)
	*/
	// "https://github.com/dave/jstest":
	/*
		Counting objects: 3, done.
		Compressing objects: 100% (2/2), done.
		Total 3 (delta 0), reused 3 (delta 0), pack-reused 0
	*/
}

func TestNew(t *testing.T) {
	fs := memfs.New()
	c := New(fs, os.Stdout)
	if err := c.Get(context.Background(), "github.com/moby/moby", false, false); err != nil {
		t.Fatal(err.Error())
	}
	var printDir func(string) error
	printDir = func(dir string) error {
		fis, err := fs.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, fi := range fis {
			fpath := filepath.Join(dir, fi.Name())
			fmt.Println(fpath)
			if fi.IsDir() {
				if err := printDir(fpath); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := printDir("/"); err != nil {
		t.Fatal(err.Error())
	}
}
