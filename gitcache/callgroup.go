package gitcache

import (
	"context"
	"sync"

	"gopkg.in/src-d/go-billy.v4"
)

// call is an in-flight or completed Do call
type call struct {
	wg  sync.WaitGroup
	val billy.Filesystem
	err error
}

// Group represents a class of work and forms a namespace in which units of work can be executed with
// duplicate suppression.
type CallGroup struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}

// Do executes and returns the results of the given function, making sure that only one execution is
// in-flight for a given key at a time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *CallGroup) Do(ctx context.Context, url string, fn func(ctx context.Context, url string) (billy.Filesystem, error)) (billy.Filesystem, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[url]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[url] = c
	g.mu.Unlock()

	c.val, c.err = fn(ctx, url)
	c.wg.Done()

	// only delete the item from the cache if it has an error
	if c.err != nil {
		g.mu.Lock()
		delete(g.m, url)
		g.mu.Unlock()
	}

	return c.val, c.err
}
