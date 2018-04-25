package get

import (
	"context"

	configpkg "github.com/dave/jsgo/config"

	"errors"

	"fmt"

	"io"

	"bufio"

	"regexp"

	"strconv"

	"strings"

	"github.com/dave/jsgo/builder/copier"
	"github.com/dave/jsgo/getter/cache"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type vcsProvider interface {
	cmd() string
	ping(ctx context.Context, scheme, repo string) error
	schemes() []string
	create(ctx context.Context, url, dir string, fs billy.Filesystem) error
	download(ctx context.Context) error
	hash() string
}

type gitProvider struct {
	gitreq     *cache.Request
	repo       *git.Repository
	worktree   *git.Worktree
	hashString string
}

func (g *gitProvider) hash() string {
	return g.hashString
}

func (g *gitProvider) checkSize(ctx context.Context, url string) error {

	store, err := filesystem.NewStorage(memfs.New())
	if err != nil {
		return err
	}

	repo, err := git.Init(store, memfs.New())
	if err != nil {
		return err
	}

	r, err := repo.CreateRemote(&config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{url},
		Fetch: []config.RefSpec{config.RefSpec("refs/heads/*:refs/heads/*")},
	})
	if err != nil {
		return err
	}

	refs, err := r.List(&git.ListOptions{})
	if err != nil {
		return err
	}

	if len(refs) > configpkg.GitMaxRefs {
		return fmt.Errorf("repo is too big - ls-remote returned %d refs - max is %d", len(refs), configpkg.GitMaxRefs)
	}
	return nil
}

var progressRegex = regexp.MustCompile(`Counting objects: (\d+), done\.?\n$`)

func newProgressWatcher() (*progressWatcher, chan error) {
	r, w := io.Pipe()
	p := &progressWatcher{
		w: w,
	}
	errchan := make(chan error)
	go func() {
		defer close(errchan)
		buf := bufio.NewReader(r)
		s, err := buf.ReadString('\n')
		p.done = true
		if err != nil {
			errchan <- err
			return
		}
		matches := progressRegex.FindStringSubmatch(s)
		if len(matches) != 2 {
			errchan <- fmt.Errorf("error parsing git progress: %#v", strings.TrimSuffix(s, "\n"))
			return
		}
		objects, err := strconv.Atoi(matches[1])
		if err != nil {
			errchan <- fmt.Errorf("error parsing git progress: %#v", strings.TrimSuffix(s, "\n"))
			return
		}
		if objects > configpkg.GitMaxObjects {
			errchan <- fmt.Errorf("too many git objects (max %d): %d", configpkg.GitMaxObjects, objects)
			return
		}
	}()
	return p, errchan
}

type progressWatcher struct {
	w    io.Writer
	done bool
}

func (p *progressWatcher) Write(b []byte) (n int, err error) {
	if p.done {
		return
	}
	return p.w.Write(b)
}

func (g *gitProvider) create(ctx context.Context, url, dir string, fs billy.Filesystem) error {
	worktree, err := g.gitreq.Fetch(ctx, url)
	if err != nil {
		return err
	}
	if err := copier.Copy("/", dir, worktree, fs); err != nil {
		return err
	}
	return nil
}

// TODO: Do something about this (it's unused right now)
func (g *gitProvider) download(ctx context.Context) error {
	// git pull --ff-only
	// git submodule update --init --recursive
	ctx, cancel := context.WithTimeout(ctx, configpkg.GitPullTimeout)
	defer cancel()
	err := g.worktree.PullContext(ctx, &git.PullOptions{
		SingleBranch:      true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Force:             true,
	})
	if err != nil {
		if err == OutOfSpace {
			return errors.New("out of space pulling repo")
		}
		return err
	}

	// ... retrieves the branch pointed by HEAD
	ref, err := g.repo.Head()
	if err != nil {
		return err
	}

	// ... retrieves the commit history
	iter, err := g.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}

	c, err := iter.Next()
	if err != nil {
		return err
	}

	g.hashString = c.Hash.String()

	return nil
}

func (g *gitProvider) cmd() string {
	return "git"
}

func (g *gitProvider) schemes() []string {
	return []string{"git", "https", "http", "git+ssh", "ssh"}
}

func (g *gitProvider) ping(ctx context.Context, scheme, repo string) error {
	repository, _ := git.Init(memory.NewStorage(), nil)

	// Add a new remote, with the default fetch refspec
	remote, err := repository.CreateRemote(&config.RemoteConfig{
		Name: "example",
		URLs: []string{scheme + "://" + repo},
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, configpkg.GitListTimeout)
	defer cancel()

	if WithCancel(ctx, func() {
		_, err = remote.List(&git.ListOptions{})
	}) {
		return ctx.Err()
	}
	return err
}
