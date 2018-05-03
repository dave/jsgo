package get

import (
	"context"

	configpkg "github.com/dave/jsgo/config"

	"errors"

	"github.com/dave/jsgo/builder/copier"
	"github.com/dave/jsgo/getter/cache"
	"gopkg.in/src-d/go-billy.v4"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
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
