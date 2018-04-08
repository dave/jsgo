package gcspersister

import (
	"io"

	"fmt"

	"crypto/sha1"

	"context"

	"cloud.google.com/go/storage"
)

type Persister struct {
	Client *storage.Client
	Bucket *storage.BucketHandle
}

func (p *Persister) Save(ctx context.Context, url string, size uint64, reader io.Reader) error {
	name := filename(url)
	ob := p.Bucket.Object(name)
	wc := ob.NewWriter(ctx)
	defer wc.Close()
	wc.ContentType = "application/octet-stream"
	wc.CacheControl = "no-cache"
	if _, err := io.Copy(wc, reader); err != nil {
		return err
	}
	return nil
}

func (p *Persister) Load(ctx context.Context, url string, writer io.Writer) (found bool, err error) {
	name := filename(url)
	ob := p.Bucket.Object(name)
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

func filename(url string) string {
	sha := sha1.New()
	sha.Write([]byte(url))
	return fmt.Sprintf("%x", sha.Sum(nil))
}
