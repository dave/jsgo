package gitfetcher

import (
	"context"
	"errors"
	"strconv"

	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/services"
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

func New(cache, fileserver services.Fileserver) *Fetcher {
	return &Fetcher{
		cache:      cache,
		fileserver: fileserver,
	}
}

type Fetcher struct {
	cache, fileserver services.Fileserver
}

func (f *Fetcher) Fetch(ctx context.Context, url string) (billy.Filesystem, error) {

	persisted, sfs, store, worktree, err := initFilesystems()
	if err != nil {
		return nil, err
	}

	exists, err := load(ctx, f.cache, url, persisted)
	if err != nil {
		return nil, err
	}

	if !exists {
		exists, err = load(ctx, f.fileserver, url, persisted)
		if err != nil {
			return nil, err
		}
	}

	var changed bool

	if exists {
		if changed, err = doFetch(ctx, store, worktree); err != nil {
			// If error while fetching, try a full clone before exiting. Make sure we re-initialise
			// the filesystems.
			persisted, sfs, store, worktree, err = initFilesystems()
			if err != nil {
				return nil, err
			}
			if changed, err = doClone(ctx, url, store, worktree); err != nil {
				return nil, err
			}
		}

	} else {
		if changed, err = doClone(ctx, url, store, worktree); err != nil {
			return nil, err
		}
	}

	if err := sfs.Sync(); err != nil {
		return nil, err
	}
	// we don't want the context to be cancelled half way through saving, so let's create a new one:
	gitctx, _ := context.WithTimeout(context.Background(), config.GitSaveTimeout)
	if changed {
		go save(gitctx, f.fileserver, url, persisted)
	}
	go save(gitctx, f.cache, url, persisted)

	return worktree, nil
}

func initFilesystems() (persisted billy.Filesystem, sfs sivafs.SivaFS, store *filesystem.Storage, worktree billy.Filesystem, err error) {

	persisted = memfs.New()

	sfs, err = sivafs.NewFilesystem(persisted, FNAME, memfs.New())
	if err != nil {
		return nil, nil, nil, nil, err
	}

	store, err = filesystem.NewStorage(sfs)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	worktree = memfs.New()

	return persisted, sfs, store, worktree, nil
}

func doFetch(ctx context.Context, store *filesystem.Storage, worktree billy.Filesystem) (changed bool, err error) {

	// Opening git repo
	repo, err := git.Open(store, worktree)
	if err != nil {
		return false, err
	}

	// Get the origin remote (all repos have origin?)
	remote, err := repo.Remote("origin")
	if err != nil {
		return false, err
	}

	// Get a list of references from the remote
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return false, err
	}

	// Find the HEAD reference. If we can't find it, return an error.
	rs := memory.ReferenceStorage{}
	for _, ref := range refs {
		rs[ref.Name()] = ref
	}
	originHead, err := storer.ResolveReference(rs, plumbing.HEAD)
	if err != nil {
		return false, err
	}
	if originHead == nil {
		return false, errors.New("HEAD not found")
	}

	// We only need to do a full Fetch if the head has changed. Compare with repo.Head().
	repoHead, err := repo.Head()
	if err != nil {
		return false, err
	}
	if originHead.Hash() != repoHead.Hash() {

		// repo has changed - this will mean it's saved after the operation
		changed = true

		ctx, cancel := context.WithTimeout(ctx, config.GitCloneTimeout)
		defer cancel()

		pw, errchan := newProgressWatcher()
		var errFromWatcher error
		go func() {
			if err := <-errchan; err != nil {
				errFromWatcher = err
				cancel()
			}
		}()

		if err := repo.FetchContext(ctx, &git.FetchOptions{Force: true, Progress: pw}); err != nil && err != git.NoErrAlreadyUpToDate {
			if errFromWatcher != nil {
				return false, errFromWatcher
			}
			return false, err
		}
	}

	// Get the worktree, and do a hard reset to the HEAD from origin.
	w, err := repo.Worktree()
	if err != nil {
		return false, err
	}
	if err := w.Reset(&git.ResetOptions{
		Commit: originHead.Hash(),
		Mode:   git.HardReset,
	}); err != nil {
		return false, err
	}

	return changed, nil
}

func doClone(ctx context.Context, url string, store *filesystem.Storage, worktree billy.Filesystem) (changed bool, err error) {

	ctx, cancel := context.WithTimeout(ctx, config.GitCloneTimeout)
	defer cancel()

	pw, errchan := newProgressWatcher()
	var errFromWatcher error
	go func() {
		if err := <-errchan; err != nil {
			errFromWatcher = err
			cancel()
		}
	}()

	if _, err := git.CloneContext(ctx, store, worktree, &git.CloneOptions{
		URL:          url,
		Progress:     pw,
		Tags:         git.NoTags,
		SingleBranch: true,
	}); err != nil {
		if errFromWatcher != nil {
			return false, errFromWatcher
		}
		return false, err
	}
	return true, nil
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
		if objects > config.GitMaxObjects {
			errchan <- fmt.Errorf("too many git objects (max %d): %d", config.GitMaxObjects, objects)
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

func save(ctx context.Context, fileserver services.Fileserver, url string, fs billy.Filesystem) error {
	// open the persisted git file for reading
	persisted, err := fs.Open(FNAME)
	if err != nil {
		return err
	}
	defer persisted.Close()
	if _, err := fileserver.Write(ctx, config.GitBucket, url, persisted, true, "application/octet-stream", "no-cache"); err != nil {
		return err
	}
	return nil
}

func load(ctx context.Context, fileserver services.Fileserver, url string, fs billy.Filesystem) (found bool, err error) {
	// open / create the persisted git file for writing
	persisted, err := fs.Create(FNAME)
	if err != nil {
		return false, err
	}
	defer persisted.Close()
	return fileserver.Read(ctx, config.GitBucket, url, persisted)
}
