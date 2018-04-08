package gitfetcher

import (
	"context"
	"errors"
	"os"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/gitcache"
	"gopkg.in/src-d/go-billy-siva.v4"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const FNAME = "repo.bin"

type Fetcher struct {
	Cache, Gcs gitcache.Persister
}

func (f *Fetcher) Fetch(ctx context.Context, url string) (billy.Filesystem, error) {

	persisted := memfs.New()

	sfs, err := sivafs.NewFilesystem(persisted, FNAME, memfs.New())
	if err != nil {
		return nil, err
	}

	store, err := filesystem.NewStorage(sfs)
	if err != nil {
		return nil, err
	}
	worktree := memfs.New()

	exists, err := load(ctx, f.Cache, url, persisted)
	if err != nil {
		return nil, err
	}

	if !exists {
		exists, err = load(ctx, f.Gcs, url, persisted)
		if err != nil {
			return nil, err
		}
	}

	var changed bool

	if exists {
		// Opening git repo
		repo, err := git.Open(store, worktree)
		if err != nil {
			return nil, err
		}

		// Get the origin remote (all repos have origin?)
		remote, err := repo.Remote("origin")
		if err != nil {
			return nil, err
		}

		// Get a list of references from the remote
		refs, err := remote.List(&git.ListOptions{})
		if err != nil {
			return nil, err
		}

		// Find the HEAD reference. If we can't find it, return an error.
		rs := memory.ReferenceStorage{}
		for _, ref := range refs {
			rs[ref.Name()] = ref
		}
		originHead, err := storer.ResolveReference(rs, plumbing.HEAD)
		if err != nil {
			return nil, err
		}
		if originHead == nil {
			return nil, errors.New("HEAD not found")
		}

		// We only need to do a full Fetch if the head has changed. Compare with repo.Head().
		repoHead, err := repo.Head()
		if err != nil {
			return nil, err
		}
		if originHead.Hash() != repoHead.Hash() {

			// repo has changed - this will mean it's saved after the operation
			changed = true

			// Do a full Fetch. Can this be made faster with options?
			if err := repo.Fetch(&git.FetchOptions{
				Force:    true,
				Progress: os.Stdout,
			}); err != nil && err != git.NoErrAlreadyUpToDate {
				return nil, err
			}
		}

		// Get the worktree, and do a hard reset to the HEAD from origin.
		w, err := repo.Worktree()
		if err != nil {
			return nil, err
		}
		if err := w.Reset(&git.ResetOptions{
			Commit: originHead.Hash(),
			Mode:   git.HardReset,
		}); err != nil {
			return nil, err
		}

	} else {

		// repo has changed - this will mean it's saved after the operation
		changed = true

		// Clone the repo
		if _, err := git.Clone(store, worktree, &git.CloneOptions{
			URL:          url,
			Tags:         git.NoTags,
			SingleBranch: true,
			Progress:     os.Stdout,
		}); err != nil {
			return nil, err
		}
	}

	if err := sfs.Sync(); err != nil {
		return nil, err
	}
	// we don't want the context to be cancelled half way through saving, so let's create a new one:
	gitctx, _ := context.WithTimeout(context.Background(), config.GitSaveTimeout)
	if changed {
		go save(gitctx, f.Gcs, url, persisted)
	}
	go save(gitctx, f.Cache, url, persisted)

	return worktree, nil
}

func save(ctx context.Context, saver gitcache.Persister, url string, fs billy.Filesystem) error {
	s, err := fs.Stat(FNAME)
	if err != nil {
		return err
	}
	// open the persisted git file for reading
	persisted, err := fs.Open(FNAME)
	if err != nil {
		return err
	}
	defer persisted.Close()
	if err := saver.Save(ctx, url, uint64(s.Size()), persisted); err != nil {
		return err
	}
	return nil
}

func load(ctx context.Context, loader gitcache.Persister, url string, fs billy.Filesystem) (found bool, err error) {
	// open / create the persisted git file for writing
	persisted, err := fs.Create(FNAME)
	if err != nil {
		return false, err
	}
	defer persisted.Close()
	return loader.Load(ctx, url, persisted)
}
