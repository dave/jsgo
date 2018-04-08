package getter

import (
	"context"
	"go/build"
	"io"
	"sync"

	"github.com/dave/jsgo/gitcache"
	"github.com/dave/jsgo/session"
	"golang.org/x/sync/singleflight"
)

type Getter struct {
	*session.Session
	gitcache          *gitcache.Request
	gitpackage        *gitcache.HintGroup
	log               io.Writer
	packageCache      map[string]*Package
	buildContext      *build.Context
	foldPath          map[string]string
	downloadCache     map[string]bool
	downloadRootCache map[string]*repoRoot
	repoPackages      map[string]*repoRoot
	fetchGroup        singleflight.Group
	fetchCacheMu      sync.Mutex
	fetchCache        map[string]fetchResult // key is metaImportsForPrefix's importPrefix
}

func New(session *session.Session, log io.Writer, cache *gitcache.Request) *Getter {
	g := &Getter{}
	g.gitcache = cache
	g.Session = session
	g.log = log
	g.packageCache = make(map[string]*Package)
	g.foldPath = make(map[string]string)
	g.downloadCache = make(map[string]bool)
	g.downloadRootCache = make(map[string]*repoRoot) // key is the root dir of the repo
	g.repoPackages = make(map[string]*repoRoot)      // key is the path of the package. NOTE: not all packages are included, but the ones we're interested in should be.
	g.fetchCache = make(map[string]fetchResult)
	g.buildContext = g.BuildContext(false, "")
	return g
}

func (g *Getter) Get(ctx context.Context, path string, update bool, insecure, single bool) error {
	var stk ImportStack
	g.gitpackage = g.gitcache.NewHintGroup(path)
	return g.download(ctx, path, nil, &stk, update, insecure, single)
}

// WithCancel executes the provided function, but returns early with true if the context cancellation
// signal was recieved.
func WithCancel(ctx context.Context, f func()) bool {
	finished := make(chan struct{})
	go func() {
		f()
		close(finished)
	}()
	select {
	case <-finished:
		return false
	case <-ctx.Done():
		return true
	}
}
