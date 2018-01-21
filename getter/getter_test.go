package getter

import (
	"testing"

	"fmt"

	"path/filepath"

	"gopkg.in/src-d/go-billy.v4/memfs"
)

func TestNew(t *testing.T) {
	fs := memfs.New()
	c := New(fs)
	if err := c.Get("github.com/dave/brenda", true, false); err != nil {
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
