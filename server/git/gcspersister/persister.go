package gcspersister

import (
	"io"

	"context"

	"net/url"

	"cloud.google.com/go/storage"
)

type Persister struct {
	Client *storage.Client
	Bucket *storage.BucketHandle
}

func (p *Persister) Save(ctx context.Context, repo string, size uint64, reader io.Reader) error {
	ob := p.Bucket.Object(url.PathEscape(repo))
	wc := ob.NewWriter(ctx)
	defer wc.Close()
	wc.ContentType = "application/octet-stream"
	wc.CacheControl = "no-cache"
	if _, err := io.Copy(wc, reader); err != nil {
		return err
	}
	return nil
}

func (p *Persister) Load(ctx context.Context, repo string, writer io.Writer) (found bool, err error) {
	ob := p.Bucket.Object(url.PathEscape(repo))
	r, err := ob.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, err
	}
	if _, err := io.Copy(writer, r); err != nil {
		return false, err
	}
	return true, nil
}
