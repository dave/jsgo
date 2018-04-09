package main

import (
	"path/filepath"

	"os"

	"io"

	"fmt"

	"errors"

	"gopkg.in/src-d/go-billy-siva.v4"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

const REPO = "https://github.com/dave/jstest.git"
const FNAME = "repo.bin"

func main() {
	if err := Get(REPO); err != nil {
		fmt.Println(err)
	}
}

func Get(url string) error {
	persisted := memfs.New()

	sfs, err := sivafs.NewFilesystem(persisted, FNAME, memfs.New())
	if err != nil {
		return err
	}

	store, err := filesystem.NewStorage(sfs)
	if err != nil {
		return err
	}
	worktree := memfs.New()

	exists, err := load(persisted)
	if err != nil {
		return err
	}

	var changed bool

	if exists {
		// Opening git repo
		repo, err := git.Open(store, worktree)
		if err != nil {
			return err
		}

		// Get the origin remote (all repos have origin?)
		remote, err := repo.Remote("origin")
		if err != nil {
			return err
		}

		// Get a list of references from the remote
		refs, err := remote.List(&git.ListOptions{})
		if err != nil {
			return err
		}

		// Find the HEAD reference. If we can't find it, return an error.
		rs := memory.ReferenceStorage{}
		for _, ref := range refs {
			rs[ref.Name()] = ref
		}
		originHead, err := storer.ResolveReference(rs, plumbing.HEAD)
		if err != nil {
			return err
		}
		if originHead == nil {
			return errors.New("HEAD not found")
		}

		// We only need to do a full Fetch if the head has changed. Compare with repo.Head().
		repoHead, err := repo.Head()
		if err != nil {
			return err
		}
		if originHead.Hash() != repoHead.Hash() {

			// repo has changed - this will mean it's saved after the operation
			changed = true

			// Do a full Fetch. Can this be made faster with options?
			if err := repo.Fetch(&git.FetchOptions{
				Force:    true,
				Progress: os.Stdout,
			}); err != nil && err != git.NoErrAlreadyUpToDate {
				return err
			}
		}

		// Get the worktree, and do a hard reset to the HEAD from origin.
		w, err := repo.Worktree()
		if err != nil {
			return err
		}
		if err := w.Reset(&git.ResetOptions{
			Commit: originHead.Hash(),
			Mode:   git.HardReset,
		}); err != nil {
			return err
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
			return err
		}
	}

	if changed {
		fmt.Println("saving")
		if err := sfs.Sync(); err != nil {
			return err
		}
		if err := save(persisted); err != nil {
			return err
		}
	}

	printFs(worktree)

	return nil
}

func save(fs billy.Filesystem) error {

	// open / create the local file for writing
	local, err := os.Create(FNAME)
	if err != nil {
		return err
	}
	defer local.Close()

	// open the persisted git file for reading
	persisted, err := fs.Open(FNAME)
	if err != nil {
		return err
	}
	defer persisted.Close()

	// copy from the persisted git file to the local file
	if _, err := io.Copy(local, persisted); err != nil {
		return err
	}

	return nil
}

func load(fs billy.Filesystem) (found bool, err error) {
	if _, err = os.Stat(FNAME); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// open / create the persisted git file for writing
	persisted, err := fs.Create(FNAME)
	if err != nil {
		return false, err
	}
	defer persisted.Close()

	// open the local file for reading
	local, err := os.Open(FNAME)
	if err != nil {
		return false, err
	}
	defer local.Close()

	// copy from the local file to the persisted git file
	if count, err := io.Copy(persisted, local); err != nil {
		return false, err
	} else {
		fmt.Println(count)
	}

	return true, nil
}

func printFs(fs billy.Filesystem) {
	err := printDir(fs, "/")
	if err != nil {
		fmt.Println(err)
	}
}

func printDir(fs billy.Filesystem, dir string) error {
	fis, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		fpath := filepath.Join(dir, fi.Name())
		fmt.Println(fpath)
		if fi.IsDir() {
			if err := printDir(fs, fpath); err != nil {
				return err
			}
		}
	}
	return nil
}
