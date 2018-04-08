package gitcache

import (
	"context"

	"io"

	"gopkg.in/src-d/go-billy.v4"
)

// HintResolver provides the functionality to load and save hints to and from a database (e.g. Google
// Datastore).
type HintResolver interface {
	ResolveHints(ctx context.Context, packagePathHints []string) (resolvedRepoUrls []string, err error)
	SaveHints(ctx context.Context, packageRepoMap map[string][]string) error
}

// RepoLoadSaver provides the functionality to load and save repos to and from a persistence medium (e.g.
// Google Storage) or a local cache.
type RepoLoadSaver interface {
	SaveRepo(url string, reader io.Reader) error
	LoadRepo(url string, writer io.Writer) (found bool, err error)
}

// New returns a new cache.
func New(resolver HintResolver, remote, local RepoLoadSaver) *Cache {
	c := &Cache{}
	c.resolver = resolver
	c.remote = remote
	c.local = local
	return c
}

// Cache stores a local cache of marshaled repos (only small repos will be cached because we're limited
// on memory). There should be one Cache per server. All methods should be safe for concurrent execution.
type Cache struct {
	resolver      HintResolver
	remote, local RepoLoadSaver
}

// Request represents a single request, possibly with several "go get" operations. It is assumed that
// all "git fetch" operations that happen in one request are current for the entire request.
type Request struct {
	cache   *Cache
	hints   []string
	getters map[*Getter]bool
}

// New returns a new request. Any packages that we know will be requested during the request can be specified
// with hints, and the request will try pre-fetch in parallel all the repos that we need to fulfill this
// request (using a database of previously encountered package->dependencies). If the dependencies have
// recently changed this will be picked up during the "go get" execution and the correct dependencies
// will be requested (this will ensure correct execution).
func (c *Cache) New(hints []string) *Request {
	r := &Request{}
	r.cache = c
	r.hints = hints
	r.getters = map[*Getter]bool{}
	r.init()
	return r
}

// init looks up hints in the database to get a best guess list of repos, then starts to fetch all of
// them in parallel (with a limited pool of workers?)
func (r *Request) init() {
}

// Close should be called once all getters have finished, and saves the hints back to the HintResolver.
func (r *Request) Close(ctx context.Context) error {
	hints := map[string][]string{}
	for g := range r.getters {
		var repos []string
		for url := range g.repos {
			repos = append(repos, url)
		}
		hints[g.path] = repos
	}
	return r.cache.resolver.SaveHints(ctx, hints)
}

// Getter returns a Getter
func (r *Request) Getter(path string) {
	g := &Getter{}
	g.request = r
	g.path = path
	g.repos = map[string]bool{}
}

// Getter represents a single "go get", and records the repos that were actually requested. These are
// then passed back to the HintResolver to be saved in the database.
type Getter struct {
	path    string
	repos   map[string]bool
	request *Request
}

// Get does either a git clone or a git fetch to ensure we have the latest version of the repo and returns
// the work tree. If a request for this repo is already in flight (e.g. from the init method), we wait
// for that one to finish instead of starting a new one.
func (g *Getter) Get(ctx context.Context, url string) (billy.Filesystem, error) {
	g.repos[url] = true
	return nil, nil
}
