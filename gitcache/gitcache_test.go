package gitcache

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"testing"
)

func TestAll(t *testing.T) {
	c := New()
}

type resolver struct {
	hints map[string][]string
}

func (r *resolver) ResolveHints(ctx context.Context, packagePathHints []string) (resolvedRepoUrls []string, err error) {
	if r.hints == nil {
		r.hints = map[string][]string{}
	}
	urls := map[string]bool{}
	for _, path := range packagePathHints {
		hints, found := r.hints[path]
		if !found {
			continue
		}
		for _, url := range hints {
			urls[url] = true
		}
	}
	var resolved []string
	for url := range urls {
		resolved = append(resolved, url)
	}
	return resolved, nil
}

func (r *resolver) SaveHints(ctx context.Context, packageRepoMap map[string][]string) error {
	if r.hints == nil {
		r.hints = map[string][]string{}
	}
	for path, urls := range packageRepoMap {
		r.hints[path] = urls
	}
	return nil
}

type persister struct {
	repos map[string][]byte
}

func (p *persister) SaveRepo(url string, reader io.Reader) error {
	if p.repos == nil {
		p.repos = map[string][]byte{}
	}
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	p.repos[url] = b
	return nil
}

func (p *persister) LoadRepo(url string, writer io.Writer) (found bool, err error) {
	if p.repos == nil {
		p.repos = map[string][]byte{}
	}
	b, ok := p.repos[url]
	if !ok {
		return false, nil
	}
	if _, err := io.Copy(writer, bytes.NewBuffer(b)); err != nil {
		return false, err
	}
	return true, nil
}
