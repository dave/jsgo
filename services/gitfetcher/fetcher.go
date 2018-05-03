package gitfetcher

import (
	"context"
	"errors"
	"strconv"

	"bufio"
	"fmt"
	"io"
	"regexp"

	"net/url"

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

/*
func init() {
	go func() {
		for {
			<-time.After(time.Second)
			fmt.Println(runtime.NumGoroutine())
		}
	}()
	go func() {
		for {
			<-time.After(time.Second * 10)
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
		}
	}()
}
*/

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
		if changed, err = doFetch(ctx, url, store, worktree); err != nil {
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

func doFetch(ctx context.Context, url string, store *filesystem.Storage, worktree billy.Filesystem) (changed bool, err error) {

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
		defer pw.stop()
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
	defer pw.stop()
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

var progressRegex = []*regexp.Regexp{
	regexp.MustCompile(`Counting objects: (\d+), done\.?`),
	regexp.MustCompile(`Finding sources: +\d+% \(\d+/(\d+)\)`),
}

func newProgressWatcher() (*progressWatcher, chan error) {
	r, w := io.Pipe()
	p := &progressWatcher{
		w: w,
	}
	scanner := bufio.NewScanner(r)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		i := strings.IndexAny(string(data), "\r\n")
		if i >= 0 {
			return i + 1, data[:i], nil
		}
		if atEOF {
			return 0, nil, io.EOF
		}
		return 0, nil, nil
	})
	errchan := make(chan error)
	go func() {
		defer close(errchan)
		for {
			ok := scanner.Scan()
			if !ok {
				return
			}
			if matched, objects := matchProgress(scanner.Text()); matched && objects > config.GitMaxObjects {
				errchan <- fmt.Errorf("too many git objects (max %d): %d", config.GitMaxObjects, objects)
			}
		}
	}()
	return p, errchan
}

type progressWatcher struct {
	w *io.PipeWriter
}

func (p *progressWatcher) stop() {
	p.w.Close()
}

func (p *progressWatcher) Write(b []byte) (n int, err error) {
	return p.w.Write(b)
}

func matchProgress(s string) (matched bool, objects int) {
	for _, r := range progressRegex {
		matches := r.FindStringSubmatch(s)
		if len(matches) != 2 {
			continue
		}
		objects, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		return true, objects
	}
	return false, 0
}

func save(ctx context.Context, fileserver services.Fileserver, repoUrl string, fs billy.Filesystem) error {
	// open the persisted git file for reading
	persisted, err := fs.Open(FNAME)
	if err != nil {
		return err
	}
	defer persisted.Close()
	if _, err := fileserver.Write(ctx, config.GitBucket, url.PathEscape(repoUrl), persisted, true, "application/octet-stream", "no-cache"); err != nil {
		return err
	}
	return nil
}

func load(ctx context.Context, fileserver services.Fileserver, repoUrl string, fs billy.Filesystem) (found bool, err error) {
	// open / create the persisted git file for writing
	persisted, err := fs.Create(FNAME)
	if err != nil {
		return false, err
	}
	defer persisted.Close()
	return fileserver.Read(ctx, config.GitBucket, url.PathEscape(repoUrl), persisted)
}
