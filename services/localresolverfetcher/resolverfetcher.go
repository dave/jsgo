package localfetcher

import (
	"context"
	"os"

	"go/build"
	"path/filepath"

	"fmt"

	"sync"

	"github.com/dave/jsgo/builder/copier"
	"golang.org/x/sync/singleflight"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
)

func New() *ResolverFetcher {
	rf := &ResolverFetcher{}
	rf.init()
	return rf
}

// FetcherResolver satisfies the Fetcher and Resolver interfaces in order to use the local GOPATH for
// git operations during testing and offline development.
type ResolverFetcher struct {
	reposM    sync.RWMutex
	repos     map[string]string // url -> dir
	packagesM sync.RWMutex
	packages  map[string]string // package path -> url
	group     singleflight.Group
}

func (f *ResolverFetcher) saveRepo(url, dir string) {
	f.reposM.Lock()
	defer f.reposM.Unlock()
	f.repos[url] = dir
}

func (f *ResolverFetcher) getRepo(url string) (string, bool) {
	f.reposM.RLock()
	defer f.reposM.RUnlock()
	dir, found := f.repos[url]
	return dir, found
}

func (f *ResolverFetcher) savePackage(path, url string) {
	f.packagesM.Lock()
	defer f.packagesM.Unlock()
	f.packages[path] = url
}

func (f *ResolverFetcher) getPackage(path string) (string, bool) {
	f.packagesM.RLock()
	defer f.packagesM.RUnlock()
	url, found := f.packages[path]
	return url, found
}

func (f *ResolverFetcher) init() error {
	// be careful not to init twice at the same time
	_, err, _ := f.group.Do("init", func() (interface{}, error) {
		f.repos = map[string]string{}
		f.packages = map[string]string{}
		src := filepath.Join(build.Default.GOPATH, "src")
		// scan gopath and make a list of all the repos
		if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			packagePath := filepath.ToSlash(rel)

			if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}
			r, err := git.PlainOpen(path)
			if err != nil {
				return err
			}
			o, err := r.Remote("origin")
			if err != nil {
				return err
			}
			f.saveRepo(o.Config().URLs[0], path)
			f.savePackage(packagePath, o.Config().URLs[0])

			//fmt.Println("Detected local git repo", o.Config().URLs[0], "at", path)

			// ignore any subdirs if this dir is a repo
			return filepath.SkipDir

		}); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (f *ResolverFetcher) Resolve(ctx context.Context, path string) (url string, root string, err error) {
	find := func() (string, string, bool) {
		p := path
		for {
			// assume the package path is of the form foo/bar/baz, and step backwards until we find a
			// matching repo e.g. first try foo/bar/baz, then foo/bar, then foo.
			url, ok := f.getPackage(p)
			if ok {
				return url, p, true
			}
			p = filepath.Dir(p)
			if p == "" || p == "." || p == "/" {
				return "", "", false
			}
		}
	}

	url, root, ok := find()
	if !ok {
		// initialise again in case we have done a manual "go get" while the server is running
		if err := f.init(); err != nil {
			return "", "", err
		}
		url, root, ok = find()
		if !ok {
			return "", "", fmt.Errorf("%s not found", path)
		}
	}

	return url, root, nil
}

func (f *ResolverFetcher) Fetch(ctx context.Context, url string) (billy.Filesystem, error) {

	dir, ok := f.getRepo(url)
	if !ok {
		// initialise again in case we have done a manual "go get" while the server is running
		if err := f.init(); err != nil {
			return nil, err
		}
		dir, ok = f.getRepo(url)
		if !ok {
			return nil, fmt.Errorf("local repo %s not found", url)
		}
	}

	fs := memfs.New()

	r, err := git.PlainOpen(dir)
	if err != nil {
		return nil, err
	}

	wt, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	if err := copier.Copy("/", "/", wt.Filesystem, fs); err != nil {
		return nil, err
	}

	return fs, nil
}
