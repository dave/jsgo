package common

import (
	"go/build"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dave/jsgo/assets"

	"gopkg.in/src-d/go-billy.v4"
)

func NewBuildContext(fs billy.Filesystem, root bool) *build.Context {
	return &build.Context{
		GOARCH:    build.Default.GOARCH, // target architecture
		GOOS:      build.Default.GOOS,   // target operating system
		GOROOT:    "/goroot",            // Go root
		GOPATH:    "/gopath",            // Go path
		Compiler:  "gc",                 // compiler to assume when computing target paths
		BuildTags: []string{"js"},

		// JoinPath joins the sequence of path fragments into a single path.
		// If JoinPath is nil, Import uses filepath.Join.
		JoinPath: path.Join,

		// SplitPathList splits the path list into a slice of individual paths.
		// If SplitPathList is nil, Import uses filepath.SplitList.
		SplitPathList: func(list string) []string {
			if list == "" {
				return nil
			}
			return strings.Split(list, "/")
		},

		// IsAbsPath reports whether path is an absolute path.
		// If IsAbsPath is nil, Import uses filepath.IsAbs.
		IsAbsPath: path.IsAbs,

		// IsDir reports whether the path names a directory.
		// If IsDir is nil, Import calls os.Stat and uses the result's IsDir method.
		IsDir: func(path string) bool {
			if strings.HasPrefix(path, "/goroot/") && root {
				f, err := assets.Assets.Open(path)
				if err != nil {
					return false
				}
				defer f.Close()
				fi, err := f.Stat()
				if err != nil {
					return false
				}
				return fi.IsDir()
			}
			fi, err := fs.Stat(path)
			return err == nil && fi.IsDir()
		},

		// HasSubdir reports whether dir is lexically a subdirectory of
		// root, perhaps multiple levels below. It does not try to check
		// whether dir exists.
		// If so, HasSubdir sets rel to a slash-separated path that
		// can be joined to root to produce a path equivalent to dir.
		// If HasSubdir is nil, Import uses an implementation built on
		// filepath.EvalSymlinks.
		HasSubdir: func(root, dir string) (rel string, ok bool) {
			const sep = string(filepath.Separator)
			root = filepath.Clean(root)
			if !strings.HasSuffix(root, sep) {
				root += sep
			}
			dir = filepath.Clean(dir)
			if !strings.HasPrefix(dir, root) {
				return "", false
			}
			return filepath.ToSlash(dir[len(root):]), true
		},

		// ReadDir returns a slice of os.FileInfo, sorted by Name,
		// describing the content of the named directory.
		// If ReadDir is nil, Import uses ioutil.ReadDir.
		ReadDir: func(path string) ([]os.FileInfo, error) {
			if strings.HasPrefix(path, "/goroot/") && root {
				f, err := assets.Assets.Open(path)
				if err != nil {
					return nil, err
				}
				defer f.Close()
				fi, err := f.Readdir(-1)
				if err != nil {
					return nil, err
				}
				return fi, nil
			}
			return fs.ReadDir(path)
		},

		// OpenFile opens a file (not a directory) for reading.
		// If OpenFile is nil, Import uses os.Open.
		OpenFile: func(path string) (io.ReadCloser, error) {
			if strings.HasPrefix(path, "/goroot/") && root {
				f, err := assets.Assets.Open(path)
				if err != nil {
					return nil, err
				}
				return f, nil
			}
			return fs.Open(path)
		},
	}
}
