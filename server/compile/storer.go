package compile

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/services"
)

type Storer struct {
	fileserver services.Fileserver
	queue      chan StorageItem
	wait       sync.WaitGroup
	send       func(messages.Message)
	unchanged  int32
	done       int32
	total      int32
	Err        error
}

func NewStorer(ctx context.Context, fileserver services.Fileserver, send func(messages.Message), workers int) *Storer {
	s := &Storer{
		fileserver: fileserver,
		queue:      make(chan StorageItem, 1000),
		wait:       sync.WaitGroup{},
		send:       send,
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
			saved, err := s.fileserver.Write(ctx, item.Bucket, item.Name, bytes.NewReader(item.Contents), item.Overwrite, item.ContentType, item.CacheControl)
			if err != nil {
				s.Err = err
				return
			}
			if item.Count {
				var unchanged, done int32
				if saved {
					unchanged = atomic.LoadInt32(&s.unchanged)
					done = atomic.AddInt32(&s.done, 1)
				} else {
					unchanged = atomic.AddInt32(&s.unchanged, 1)
					done = atomic.LoadInt32(&s.done)
				}
				remain := atomic.LoadInt32(&s.total) - unchanged - done
				if s.send != nil {
					s.send(messages.Storing{Finished: int(done), Unchanged: int(unchanged), Remain: int(remain)})
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
		cacheHeader = "public,max-age=31536000,immutable"
	}

	s.queue <- StorageItem{
		Message:      message,
		Bucket:       bucket,
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

func (s *Storer) AddHtmlCached(message, name string, contents []byte) {
	s.add(message, name, contents, config.IndexBucket, "text/html", true, true, false)
}

func (s *Storer) AddZip(message, name string, contents []byte) {
	s.add(message, name, contents, config.PkgBucket, "application/zip", false, false, true)
}

type StorageItem struct {
	Message      string
	Bucket       string
	Name         string
	Contents     []byte
	ContentType  string
	CacheControl string
	Overwrite    bool
	Count        bool
}
