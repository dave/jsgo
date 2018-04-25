package gcsfileserver

import (
	"context"

	"github.com/dave/jsgo/config"

	"io"

	"cloud.google.com/go/storage"
)

func New(client *storage.Client) *Fileserver {
	f := new(Fileserver)
	f.client = client
	f.buckets = map[string]*storage.BucketHandle{
		config.SrcBucket:   client.Bucket(config.SrcBucket),
		config.PkgBucket:   client.Bucket(config.PkgBucket),
		config.IndexBucket: client.Bucket(config.IndexBucket),
		config.GitBucket:   client.Bucket(config.GitBucket),
	}
	return f
}

type Fileserver struct {
	client  *storage.Client
	buckets map[string]*storage.BucketHandle
}

func (f *Fileserver) Write(ctx context.Context, bucket, name string, reader io.Reader, overwrite bool, contentType, cacheControl string) (saved bool, err error) {
	ob := f.buckets[bucket].Object(name)
	if !overwrite {
		// err == nil => file exists, return with saved == false
		// err == storage.ErrObjectNotExist => file doesn't exist, so continue and write file
		// err != storage.ErrObjectNotExist => any other error, so return the error
		if _, err := ob.Attrs(ctx); err == nil {
			return false, nil
		} else if err != storage.ErrObjectNotExist {
			return false, err
		}
	}
	wc := ob.NewWriter(ctx)
	defer wc.Close()
	wc.ContentType = contentType
	wc.CacheControl = cacheControl
	if _, err := io.Copy(wc, reader); err != nil {
		return false, err
	}
	return true, nil
}

func (f *Fileserver) Read(ctx context.Context, bucket, name string, writer io.Writer) (found bool, err error) {
	ob := f.buckets[bucket].Object(name)
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
