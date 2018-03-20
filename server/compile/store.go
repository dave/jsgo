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
			if item.OnlyIfNotExist {
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

func (s *Storer) AddSrc(message, name string, contents []byte) {
	s.wait.Add(1)

	unchanged := atomic.LoadInt32(&s.unchanged)
	done := atomic.LoadInt32(&s.done)
	remain := atomic.AddInt32(&s.total, 1) - unchanged - done
	if s.send != nil {
		s.send(messages.Storing{Finished: int(done), Unchanged: int(unchanged), Remain: int(remain)})
	}

	s.queue <- StorageItem{
		Message:        message,
		Bucket:         s.buckets[config.SrcBucket],
		Name:           name,
		Contents:       contents,
		ContentType:    "application/json",
		CacheControl:   "public, max-age=31536000",
		OnlyIfNotExist: true,
		Count:          true,
	}
}

func (s *Storer) AddJs(message, name string, contents []byte) {
	s.wait.Add(1)

	unchanged := atomic.LoadInt32(&s.unchanged)
	done := atomic.LoadInt32(&s.done)
	remain := atomic.AddInt32(&s.total, 1) - unchanged - done
	if s.send != nil {
		s.send(messages.Storing{Finished: int(done), Unchanged: int(unchanged), Remain: int(remain)})
	}

	s.queue <- StorageItem{
		Message:        message,
		Bucket:         s.buckets[config.PkgBucket],
		Name:           name,
		Contents:       contents,
		ContentType:    "application/javascript",
		CacheControl:   "public, max-age=31536000",
		OnlyIfNotExist: true,
		Count:          true,
	}
}

func (s *Storer) AddHtml(message, name string, contents []byte) {
	s.wait.Add(1)
	s.queue <- StorageItem{
		Message:        message,
		Bucket:         s.buckets[config.IndexBucket],
		Name:           name,
		Contents:       contents,
		ContentType:    "text/html",
		CacheControl:   "no-cache",
		OnlyIfNotExist: false,
		Count:          false,
	}
}

type StorageItem struct {
	Message        string
	Bucket         *storage.BucketHandle
	Name           string
	Contents       []byte
	ContentType    string
	CacheControl   string
	OnlyIfNotExist bool
	Count          bool
}
