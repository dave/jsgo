package getter

import (
	"log"
	"os"
	"testing"

	"fmt"
	"path/filepath"

	"context"

	"runtime"

	"github.com/dave/jsgo/config"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

func TestClone(t *testing.T) {
	for {

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		log.Println("mem.Alloc", mem.Alloc)
		log.Println("mem.TotalAlloc", mem.TotalAlloc)
		log.Println("mem.HeapAlloc", mem.HeapAlloc)
		log.Println("mem.HeapSys", mem.HeapSys)

		store, err := filesystem.NewStorage(NewWriteLimitedFilesystem(memfs.New(), config.GitMaxBytes))
		if err != nil {
			t.Fatal(err)
		}
		_, err = git.Clone(store, memfs.New(), &git.CloneOptions{
			URL:               "https://github.com/kubernetes/kubernetes",
			SingleBranch:      true,
			Depth:             1,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func TestNew(t *testing.T) {
	fs := memfs.New()
	c := New(fs, os.Stdout)
	if err := c.Get(context.Background(), "github.com/dave/material", false, false); err != nil {
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
