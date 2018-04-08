package gitcache

import (
	"context"

	"io"

	"fmt"

	"gopkg.in/src-d/go-billy.v4"
)

// Resolver provides the functionality to load and save hints to and from a database (e.g. Google
// Datastore).
type Resolver interface {
	Resolve(ctx context.Context, hints []string) (resolved []string, err error)
	Save(ctx context.Context, resolved map[string][]string) error
}

// Persister provides the functionality to load and save repos to and from a persistence medium (e.g.
// Google Storage) or a local cache.
type Persister interface {
	Save(ctx context.Context, url string, size uint64, reader io.Reader) error
	Load(ctx context.Context, url string, writer io.Writer) (found bool, err error)
}

type Fetcher interface {
	Fetch(ctx context.Context, url string) (billy.Filesystem, error)
}

// New returns a new cache.
func NewCache(resolver Resolver, git Fetcher) *Cache {
	c := &Cache{}
	c.resolver = resolver
	c.git = git
	return c
}

// Cache stores a local cache of marshaled repos (only small repos will be cached because we're limited
// on memory). There should be one Cache per server. All methods should be safe for concurrent execution.
type Cache struct {
	resolver Resolver
	git      Fetcher
}

// Request represents a single request, possibly with several "go get" operations. It is assumed that
// all "git fetch" operations that happen in one request are current for the entire request.
type Request struct {
	cache    *Cache
	fetchers map[*Package]bool
	calls    *CallGroup
}

// New returns a new request. Any packages that we know will be requested during the request can be specified
// with hints, and the request will try pre-fetch in parallel all the repos that we need to fulfill this
// request (using a database of previously encountered package->dependencies). If the dependencies have
// recently changed this will be picked up during the "go get" execution and the correct dependencies
// will be requested (this will ensure correct execution).
func (c *Cache) NewRequest() *Request {
	r := &Request{}
	r.cache = c
	r.calls = new(CallGroup)
	r.fetchers = map[*Package]bool{}
	return r
}

// Hint looks up hints in the database to get a best guess list of repos, then starts to fetch all of
// them in parallel (with a limited pool of workers?)
func (r *Request) Hint(ctx context.Context, hints ...string) error {
	urls, err := r.cache.resolver.Resolve(ctx, hints)
	if err != nil {
		return err
	}
	fmt.Println("******* got hints: ", urls)
	for _, url := range urls {
		url := url
		go r.calls.Do(ctx, url, r.fetch)
	}
	return nil
}

// fetch is called by Request.init and Fetcher.Fetch.
func (r *Request) fetch(ctx context.Context, url string) (billy.Filesystem, error) {
	fs, err := r.cache.git.Fetch(ctx, url)
	if err != nil {
		fmt.Println("error fetching", url, err)
		return nil, err
	}
	return fs, nil
}

// Close should be called once all getters have finished, and saves the hints back to the HintResolver.
func (r *Request) Close(ctx context.Context) error {
	hints := map[string][]string{}
	for g := range r.fetchers {
		var repos []string
		for url := range g.repos {
			repos = append(repos, url)
		}
		hints[g.path] = repos
	}
	return r.cache.resolver.Save(ctx, hints)
}

// NewPackage returns a Package
func (r *Request) NewPackage(path string) *Package {
	f := &Package{}
	f.request = r
	f.path = path
	f.repos = map[string]bool{}
	r.fetchers[f] = true
	return f
}

// Package represents a single "go get" command, and records the repos that were actually requested.
// These are then passed back to the HintResolver to be saved in the database.
type Package struct {
	path    string
	repos   map[string]bool
	request *Request
}

// Fetch does either a git clone or a git fetch to ensure we have the latest version of the repo and
// returns the work tree. If a request for this repo is already in flight (e.g. from the init method),
// we wait for that one to finish instead of starting a new one.
func (p *Package) Fetch(ctx context.Context, url string) (billy.Filesystem, error) {
	fmt.Println("=== adding repo", p.path, url)
	p.repos[url] = true
	return p.request.calls.Do(ctx, url, p.request.fetch)
}
