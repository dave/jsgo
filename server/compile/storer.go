package compile

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"

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

func (s *Storer) Add(item StorageItem) {
	s.wait.Add(1)

	if item.Count {
		unchanged := atomic.LoadInt32(&s.unchanged)
		done := atomic.LoadInt32(&s.done)
		remain := atomic.AddInt32(&s.total, 1) - unchanged - done
		if s.send != nil {
			s.send(messages.Storing{Finished: int(done), Unchanged: int(unchanged), Remain: int(remain)})
		}
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

type StorageItem struct {
	Message   string
	Bucket    string
	Name      string
	Contents  []byte
	Mime      string
	Immutable bool
	Count     bool
	Wait      *sync.WaitGroup
}
