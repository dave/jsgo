// package cstorer (concurrent storer) for storing items into a services.Fileserver concurrently
package cstorer

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"

	"github.com/dave/jsgo/services"
)

type Storer struct {
	fileserver services.Fileserver
	queue      chan Item
	wait       sync.WaitGroup
	unchanged  int32
	done       int32
	total      int32
	Err        error
}

func New(ctx context.Context, fileserver services.Fileserver, workers int) *Storer {
	s := &Storer{
		fileserver: fileserver,
		queue:      make(chan Item, 1000),
		wait:       sync.WaitGroup{},
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
			if item.Wait != nil {
				defer item.Wait.Done()
			}
			overwrite := true
			cacheControl := "no-cache"
			if item.Immutable {
				overwrite = false
				cacheControl = "public,max-age=31536000,immutable"
			}
			saved, err := s.fileserver.Write(ctx, item.Bucket, item.Name, bytes.NewReader(item.Contents), overwrite, item.Mime, cacheControl)
			if err != nil {
				s.Err = err
				return
			}
			if item.Count {
				if saved {
					atomic.AddInt32(&s.done, 1)
				} else {
					atomic.AddInt32(&s.unchanged, 1)
				}
			}
			if item.Changed != nil {
				item.Changed(true)
			}
		}()
	}
}

func (s *Storer) Stats() (total, done, unchanged int) {
	total = int(atomic.LoadInt32(&s.total))
	done = int(atomic.LoadInt32(&s.done))
	unchanged = int(atomic.LoadInt32(&s.unchanged))
	return total, done, unchanged
}

func (s *Storer) Add(item Item) {
	s.wait.Add(1)

	if item.Count {
		atomic.AddInt32(&s.total, 1)
	}
	if item.Changed != nil {
		item.Changed(false)
	}

	s.queue <- item

}

const (
	MimeJson = "application/json"
	MimeJs   = "application/javascript"
	MimeBin  = "application/octet-stream"
	MimeHtml = "text/html"
	MimeZip  = "application/zip"
)

type Item struct {
	Message   string
	Bucket    string
	Name      string
	Contents  []byte
	Mime      string
	Immutable bool
	Count     bool
	Wait      *sync.WaitGroup
	Changed   func(done bool)
}
