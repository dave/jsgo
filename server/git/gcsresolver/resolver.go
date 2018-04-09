package gcsresolver

import (
	"context"

	"cloud.google.com/go/datastore"
)

type Resolver struct {
	Client *datastore.Client
	Kind   string
}

func (r *Resolver) Resolve(ctx context.Context, hints []string) (resolved []string, err error) {
	var keys []*datastore.Key
	for _, path := range hints {
		keys = append(keys, r.hintsKey(path))
	}
	urls := map[string]bool{}
	response := make([]Hints, len(keys))
	if err := r.Client.GetMulti(ctx, keys, response); err != nil {
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

func (r *Resolver) Save(ctx context.Context, resolved map[string][]string) error {
	var keys []*datastore.Key
	var vals []Hints
	for path, hints := range resolved {
		keys = append(keys, r.hintsKey(path))
		vals = append(vals, Hints{Path: path, Hints: hints})
	}
	if _, err := r.Client.PutMulti(ctx, keys, vals); err != nil {
		return err
	}
	return nil
}

func (r *Resolver) hintsKey(path string) *datastore.Key {
	return datastore.NameKey(r.Kind, path, nil)
}

type Hints struct {
	Path  string
	Hints []string
}
