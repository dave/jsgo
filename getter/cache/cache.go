package cache

import (
	"context"

	"github.com/dave/services"
	"gopkg.in/src-d/go-billy.v4"
)

// New returns a new cache.
func New(database services.Database, fetcher services.Fetcher, resolver services.Resolver, configHintsKind string) *Cache {
	c := &Cache{}
	c.database = database
	c.fetcher = fetcher
	c.resolver = resolver
	c.configHintsKind = configHintsKind
	return c
}

// Cache stores a local cache of marshaled repos (only small repos will be cached because we're limited
// on memory). There should be one Cache per server. All methods should be safe for concurrent execution.
type Cache struct {
	database        services.Database
	fetcher         services.Fetcher
	resolver        services.Resolver
	configHintsKind string
}

// Request represents a single request, possibly with several "go get" operations. It is assumed that
// all "git fetch" operations that happen in one request are current for the entire request.
type Request struct {
	cache *Cache
	hints map[string][]string
	calls *CallGroup
	save  bool // should we save hints?
}

// New returns a new request. Any packages that we know will be requested during the request can be specified
// with hints, and the request will try pre-fetch in parallel all the repos that we need to fulfill this
// request (using a database of previously encountered package->dependencies). If the dependencies have
// recently changed this will be picked up during the "go get" execution and the correct dependencies
// will be requested (this will ensure correct execution).
func (c *Cache) NewRequest(save bool) *Request {
	r := &Request{}
	r.cache = c
	r.calls = new(CallGroup)
	r.hints = map[string][]string{}
	r.save = save
	return r
}

// Hint looks up hints in the database to get a best guess list of repos, then starts to fetch all of
// them in parallel
// TODO: use a worker pool
func (r *Request) InitialiseFromHints(ctx context.Context, paths ...string) error {
	urls, err := r.cache.ResolveHints(ctx, paths)
	if err != nil {
		return err
	}
	for _, url := range urls {
		url := url
		go r.calls.Do(ctx, url, r.fetch)
	}
	return nil
}

// fetch is called by Request.init and HintGroup.Fetch.
func (r *Request) fetch(ctx context.Context, url string) (billy.Filesystem, error) {
	fs, err := r.cache.fetcher.Fetch(ctx, url)
	if err != nil {
		return nil, err
	}
	return fs, nil
}

// Fetch does either a git clone or a git fetch to ensure we have the latest version of the repo and
// returns the work tree. If a request for this repo is already in flight (e.g. from the init method),
// we wait for that one to finish instead of starting a new one.
func (r *Request) Fetch(ctx context.Context, url string) (billy.Filesystem, error) {
	return r.calls.Do(ctx, url, r.fetch)
}

// Stores hints
func (r *Request) SetHints(hints map[string][]string) {
	for path, urls := range hints {
		r.hints[path] = urls
	}
}

// Close should be called once all getters have finished, and saves the hints back to the HintResolver.
func (r *Request) Close(ctx context.Context) error {
	if !r.save {
		return nil
	}
	return r.cache.SaveHints(ctx, r.hints)
}

func (r *Request) Resolver() services.Resolver {
	return r.cache.resolver
}
