package services

import (
	"context"
	"io"

	billy "gopkg.in/src-d/go-billy.v4"
)

// Fileserver provides the functionality to persist files and host them for web delivery. In production
// we use GCS buckets. In local storage mode we use a temporary directory
type Fileserver interface {
	Write(ctx context.Context, bucket, name string, reader io.Reader, overwrite bool, contentType, cacheControl string) (saved bool, err error)
	Read(ctx context.Context, bucket, name string, writer io.Writer) (found bool, err error)
}

// Database provides the functionality to persist and recall data. In production we use the gcs datastore.
// In local development mode we use a memory data store.
type Database interface{}

// Resolver provides the functionality to resolve package paths to repo URLs. In production mode this
// uses the `go get` codebase. In local mode we scan the gopath.
type Resolver interface {
	Resolve(ctx context.Context, path string) (url string, root string, err error)
}

// Hinter provides the functionality to load and save hints to and from a database (e.g. Google
// Datastore).
type Hinter interface {
	Resolve(ctx context.Context, hints []string) (resolved []string, err error)
	Save(ctx context.Context, resolved map[string][]string) error
}

// Fetcher fetches the worktree for a git repository. In production the repo is cloned or loaded from
// the persistence cache and fetched to ensure it's update. In local development mode the repo is loaded
// from the local GOPATH.
type Fetcher interface {
	Fetch(ctx context.Context, url string) (billy.Filesystem, error)
}
