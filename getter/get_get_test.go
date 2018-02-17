package getter

import (
	"os"
	"testing"

	"fmt"
	"path/filepath"

	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func TestClone(t *testing.T) {
	_, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:               "https://gitlab.com/agamigo/material",
		SingleBranch:      true,
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNew(t *testing.T) {
	fs := memfs.New()
	c := New(fs, os.Stdout)
	if err := c.Get("github.com/dave/material", false, false); err != nil {
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
