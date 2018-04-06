package main

import (
	"path/filepath"

	"os"

	"io"

	"fmt"

	"gopkg.in/src-d/go-billy-siva.v4"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

const REPO = "https://github.com/dave/jstest.git"
const FNAME = "repo.bin"

// Output:
/*
	First run - should do git clone:
	creating siva fs
	creating storage from siva fs
	loading local file into persisted fs

	Before git operation...
	files in persisted fs:
	files in siva fs:

	Local file doesn't exist... (git clone)
	* git.Clone
	Counting objects: 11, done.
	Total 11 (delta 0), reused 0 (delta 0), pack-reused 11

	After git operation...
	files in persisted fs:
	/repo.bin
	files in siva fs:
	/objects
	/objects/pack
	/objects/pack/pack-900cc2c139ecccc51a6c0f2bd65ea4af44861bae.idx
	/objects/pack/pack-900cc2c139ecccc51a6c0f2bd65ea4af44861bae.pack
	/refs
	/refs/remotes
	/refs/remotes/origin
	/refs/remotes/origin/master
	/refs/heads
	/refs/heads/master
	/HEAD
	/config
	/index
	files in worktree fs:
	/.git
	/main.go

	Saving persisted file to local disk...
	saved 3745 bytes

	Second run - should do git pull:
	creating siva fs
	creating storage from siva fs
	loading local file into persisted fs
	loaded 3745 bytes

	Before git operation...
	files in persisted fs:
	/repo.bin
	files in siva fs:
	panic: runtime error: slice bounds out of range
*/

func main() {
	defer func() {
		os.Remove(FNAME)
	}()
	fmt.Println("")
	fmt.Println("First run - should do git clone:")
	if err := Get(REPO); err != nil {
		fmt.Println(err)
	}
	fmt.Println("")
	fmt.Println("Second run - should do git pull:")
	if err := Get(REPO); err != nil {
		fmt.Println(err)
	}
}

func Get(url string) error {
	persisted := memfs.New()

	fmt.Println("creating siva fs")
	sfs, err := sivafs.NewFilesystem(persisted, FNAME, memfs.New())
	if err != nil {
		fmt.Println("error creating siva fs")
		return err
	}

	fmt.Println("creating storage from siva fs")
	store, err := filesystem.NewStorage(sfs)
	if err != nil {
		fmt.Println("error creating storage")
		return err
	}
	worktree := memfs.New()

	fmt.Println("loading local file into persisted fs")
	exists, err := load(persisted)
	if err != nil {
		fmt.Println("error loading local file")
		return err
	}

	fmt.Println("")
	fmt.Println("Before git operation...")

	fmt.Println("files in persisted fs:")
	printFs(persisted)

	fmt.Println("files in siva fs:")
	printFs(sfs)

	if exists {
		fmt.Println("")
		fmt.Println("Local file exists... (git pull)")
		fmt.Println("* git.Open")
		repo, err := git.Open(store, worktree)
		if err != nil {
			fmt.Println("error during git.Open")
			return err
		}
		fmt.Println("* repo.Worktree")
		w, err := repo.Worktree()
		if err != nil {
			fmt.Println("error during repo.Worktree")
			return err
		}
		fmt.Println("* w.Pull")
		if err := w.Pull(&git.PullOptions{Force: true, Progress: os.Stdout}); err != nil {
			fmt.Println("error during w.Pull")
			return err
		}
	} else {
		fmt.Println("")
		fmt.Println("Local file doesn't exist... (git clone)")
		fmt.Println("* git.Clone")
		if _, err := git.Clone(store, worktree, &git.CloneOptions{URL: url, Progress: os.Stdout}); err != nil {
			fmt.Println("error during git.Clone")
			return err
		}
	}

	fmt.Println("")
	fmt.Println("After git operation...")

	fmt.Println("files in persisted fs:")
	printFs(persisted)

	fmt.Println("files in siva fs:")
	printFs(sfs)

	fmt.Println("files in worktree fs:")
	printFs(worktree)

	fmt.Println("")
	fmt.Println("Saving persisted file to local disk...")
	if err := save(persisted); err != nil {
		fmt.Println("error saving persisted file")
		return err
	}

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
	count, err := io.Copy(local, persisted)
	if err != nil {
		return err
	}

	fmt.Printf("saved %d bytes\n", count)

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
	count, err := io.Copy(persisted, local)
	if err != nil {
		return false, err
	}

	fmt.Printf("loaded %d bytes\n", count)

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
