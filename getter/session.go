package getter

import (
	"context"
	"go/build"
	"io"
	"os"
	"sync"

	"path"
	"path/filepath"
	"strings"

	"github.com/dave/jsgo/assets"
	"golang.org/x/sync/singleflight"
	"gopkg.in/src-d/go-billy.v4"
)

type Session struct {
	fs                billy.Filesystem
	log               io.Writer
	packageCache      map[string]*Package
	buildContext      *build.Context
	foldPath          map[string]string
	downloadCache     map[string]bool
	downloadRootCache map[string]*repoRoot
	repoPackages      map[string]*repoRoot
	fetchGroup        singleflight.Group
	fetchCacheMu      sync.Mutex
	fetchCache        map[string]fetchResult // key is metaImportsForPrefix's importPrefix
}

func New(fs billy.Filesystem, log io.Writer, buildTags []string) *Session {
	s := &Session{}

	s.fs = fs
	s.log = log
	s.packageCache = make(map[string]*Package)
	s.foldPath = make(map[string]string)
	s.downloadCache = make(map[string]bool)
	s.downloadRootCache = make(map[string]*repoRoot) // key is the root dir of the repo
	s.repoPackages = make(map[string]*repoRoot)      // key is the path of the package. NOTE: not all packages are included, but the ones we're interested in should be.
	s.fetchCache = make(map[string]fetchResult)
	s.buildContext = newBuildContext(fs, buildTags)
	return s

}

func (s *Session) Get(ctx context.Context, path string, update bool, insecure, single bool) error {
	var stk ImportStack
	return s.download(ctx, path, nil, &stk, update, insecure, single)
}

func newBuildContext(fs billy.Filesystem, buildTags []string) *build.Context {
	return &build.Context{
		GOARCH:      "amd64",   // target architecture
		GOOS:        "darwin",  // target operating system
		GOROOT:      "/goroot", // Go root
		GOPATH:      "/gopath", // Go path
		Compiler:    "gc",      // compiler to assume when computing target paths
		BuildTags:   append(buildTags, "js"),
		ReleaseTags: build.Default.ReleaseTags,

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
			if strings.HasPrefix(path, "/goroot/") {
				fi, err := assets.Assets.Stat(path)
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
			if strings.HasPrefix(path, "/goroot/") {
				fi, err := assets.Assets.ReadDir(path)
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
			if strings.HasPrefix(path, "/goroot/") {
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

// WithCancel executes the provided function, but returns early with true if the context cancellation
// signal was recieved.
func WithCancel(ctx context.Context, f func()) bool {
	finished := make(chan struct{})
	go func() {
		f()
		close(finished)
	}()
	select {
	case <-finished:
		return false
	case <-ctx.Done():
		return true
	}
}
