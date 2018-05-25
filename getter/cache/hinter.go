package cache

import (
	"context"

	"cloud.google.com/go/datastore"
)

func (c *Cache) ResolveHints(ctx context.Context, hints []string) (resolved []string, err error) {
	var keys []*datastore.Key
	for _, path := range hints {
		keys = append(keys, c.hintsKey(path))
	}
	urls := map[string]bool{}
	response := make([]Hints, len(keys))
	if err := c.database.GetMulti(ctx, keys, response); err != nil {
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

func (c *Cache) SaveHints(ctx context.Context, resolved map[string][]string) error {
	var keys []*datastore.Key
	var vals []Hints
	for path, hints := range resolved {
		keys = append(keys, c.hintsKey(path))
		vals = append(vals, Hints{Path: path, Hints: hints})
	}
	if _, err := c.database.PutMulti(ctx, keys, vals); err != nil {
		return err
	}
	return nil
}

func (c *Cache) hintsKey(path string) *datastore.Key {
	return datastore.NameKey(c.configHintsKind, path, nil)
}

type Hints struct {
	Path  string
	Hints []string
}
