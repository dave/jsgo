package localhinter

import (
	"context"
)

func New() *Hinter {
	return &Hinter{}
}

type Hinter struct {
	hints map[string][]string
}

func (h *Hinter) Resolve(ctx context.Context, hints []string) (resolved []string, err error) {
	if h.hints == nil {
		return nil, nil
	}
	urls := map[string]bool{}
	for _, hint := range hints {
		for _, r := range h.hints[hint] {
			urls[r] = true
		}
	}
	for r := range urls {
		resolved = append(resolved, r)
	}
	return resolved, nil
}

func (h *Hinter) Save(ctx context.Context, resolved map[string][]string) error {
	if h.hints == nil {
		h.hints = map[string][]string{}
	}
	for path, urls := range resolved {
		h.hints[path] = urls
	}
	return nil
}
