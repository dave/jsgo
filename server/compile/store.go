package compile

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"cloud.google.com/go/storage"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/messages"
)

type Storer struct {
	client    *storage.Client
	queue     chan StorageItem
	buckets   map[string]*storage.BucketHandle
	wait      sync.WaitGroup
	send      func(messages.Message)
	unchanged int32
	done      int32
	total     int32
	Err       error
}

func NewStorer(ctx context.Context, client *storage.Client, send func(messages.Message), workers int) *Storer {
	s := &Storer{
		client: client,
		buckets: map[string]*storage.BucketHandle{
			config.SrcBucket:   client.Bucket(config.SrcBucket),
			config.PkgBucket:   client.Bucket(config.PkgBucket),
			config.IndexBucket: client.Bucket(config.IndexBucket),
		},
		queue: make(chan StorageItem, 1000),
		wait:  sync.WaitGroup{},
		send:  send,
	}
	for i := 0; i < workers; i++ {
		go s.Worker(ctx)
	}
	return s
}

func (s *Storer) Close() {
	close(s.queue)
}

func (s *Storer) Wait() {
	s.wait.Wait()
}

func (s *Storer) Worker(ctx context.Context) {
	for item := range s.queue {
		func() {
			defer s.wait.Done()
			ob := item.Bucket.Object(item.Name)
			if !item.Overwrite {
				_, err := ob.Attrs(ctx)
				if err == nil || err != storage.ErrObjectNotExist {
					if item.Count {
						unchanged := atomic.AddInt32(&s.unchanged, 1)
						done := atomic.LoadInt32(&s.done)
						remain := atomic.LoadInt32(&s.total) - unchanged - done
						if s.send != nil {
							s.send(messages.Storing{Finished: int(done), Unchanged: int(unchanged), Remain: int(remain)})
						} else {
							fmt.Printf("Unchanged: %s\n", item.Message)
						}
					}
					return
				}
			}
			wc := ob.NewWriter(ctx)
			defer wc.Close()
			wc.ContentType = item.ContentType
			wc.CacheControl = item.CacheControl
			if _, err := io.Copy(wc, bytes.NewBuffer(item.Contents)); err != nil {
				s.Err = err
				return
			}
			if item.Count {
				unchanged := atomic.LoadInt32(&s.unchanged)
				done := atomic.AddInt32(&s.done, 1)
				remain := atomic.LoadInt32(&s.total) - unchanged - done
				if s.send != nil {
					s.send(messages.Storing{Finished: int(done), Unchanged: int(unchanged), Remain: int(remain)})
				} else {
					fmt.Printf("Finished: %s\n", item.Message)
				}
			}
		}()
	}
}

func (s *Storer) add(message, name string, contents []byte, bucket, mime string, count, cache, overwrite bool) {
	s.wait.Add(1)

	if count {
		unchanged := atomic.LoadInt32(&s.unchanged)
		done := atomic.LoadInt32(&s.done)
		remain := atomic.AddInt32(&s.total, 1) - unchanged - done
		if s.send != nil {
			s.send(messages.Storing{Finished: int(done), Unchanged: int(unchanged), Remain: int(remain)})
		}
	}

	cacheHeader := "no-cache"
	if cache {
		cacheHeader = "public, max-age=31536000"
	}

	s.queue <- StorageItem{
		Message:      message,
		Bucket:       s.buckets[bucket],
		Name:         name,
		Contents:     contents,
		ContentType:  mime,
		CacheControl: cacheHeader,
		Overwrite:    overwrite,
		Count:        count,
	}

}

func (s *Storer) AddSrc(message, name string, contents []byte) {
	s.add(message, name, contents, config.SrcBucket, "application/json", true, true, false)
}

func (s *Storer) AddJs(message, name string, contents []byte) {
	s.add(message, name, contents, config.PkgBucket, "application/javascript", true, true, false)
}

func (s *Storer) AddArchive(message, name string, contents []byte) {
	s.add(message, name, contents, config.PkgBucket, "application/octet-stream", true, true, false)
}

func (s *Storer) AddHtml(message, name string, contents []byte) {
	s.add(message, name, contents, config.IndexBucket, "text/html", false, false, true)
}

func (s *Storer) AddZip(message, name string, contents []byte) {
	s.add(message, name, contents, config.PkgBucket, "application/zip", false, false, true)
}

type StorageItem struct {
	Message      string
	Bucket       *storage.BucketHandle
	Name         string
	Contents     []byte
	ContentType  string
	CacheControl string
	Overwrite    bool
	Count        bool
}
