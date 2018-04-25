package gcshinter

import (
	"context"

	"cloud.google.com/go/datastore"
)

func New(client *datastore.Client, kind string) *Hinter {
	return &Hinter{
		client: client,
		kind:   kind,
	}
}

type Hinter struct {
	client *datastore.Client
	kind   string
}

func (h *Hinter) Resolve(ctx context.Context, hints []string) (resolved []string, err error) {
	var keys []*datastore.Key
	for _, path := range hints {
		keys = append(keys, h.hintsKey(path))
	}
	urls := map[string]bool{}
	response := make([]Hints, len(keys))
	if err := h.client.GetMulti(ctx, keys, response); err != nil {
		if me, ok := err.(datastore.MultiError); ok {
			for _, merr := range me {
				if merr != datastore.ErrNoSuchEntity {
					return nil, merr
				}
			}
		} else {
			if err != datastore.ErrNoSuchEntity {
				return nil, err
			}
		}
	}
	for _, r := range response {
		for _, url := range r.Hints {
			urls[url] = true
		}
	}
	var out []string
	for url := range urls {
		out = append(out, url)
	}
	return out, nil
}

func (h *Hinter) Save(ctx context.Context, resolved map[string][]string) error {
	var keys []*datastore.Key
	var vals []Hints
	for path, hints := range resolved {
		keys = append(keys, h.hintsKey(path))
		vals = append(vals, Hints{Path: path, Hints: hints})
	}
	if _, err := h.client.PutMulti(ctx, keys, vals); err != nil {
		return err
	}
	return nil
}

func (h *Hinter) hintsKey(path string) *datastore.Key {
	return datastore.NameKey(h.kind, path, nil)
}

type Hints struct {
	Path  string
	Hints []string
}
