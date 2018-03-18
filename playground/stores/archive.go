package stores

import (
	"fmt"

	"strings"

	"bytes"

	"encoding/gob"

	"compress/gzip"

	"errors"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/server/messages"
	"github.com/gopherjs/gopherjs/compiler"
	"honnef.co/go/js/dom"
)

type ArchiveStore struct {
	app *App

	// is the update in progress?
	updating bool

	// cache (path -> item) of archives
	cache map[string]CacheItem

	// index (path -> item) of the previously received update
	index messages.Index

	// has the cache finished downloading?
	complete bool

	// are the imports from the editor all in the previous update index?
	fresh bool
}

type CacheItem struct {
	Hash    string
	Archive *compiler.Archive
}

func NewArchiveStore(app *App) *ArchiveStore {
	s := &ArchiveStore{
		app:   app,
		cache: map[string]CacheItem{},
	}
	return s
}

func (s *ArchiveStore) Collect(imports []string) ([]*compiler.Archive, error) {
	var deps []*compiler.Archive
	paths := make(map[string]bool)
	var collectDependencies func(path string) error
	collectDependencies = func(path string) error {
		if paths[path] {
			return nil
		}
		item, ok := s.cache[path]
		if !ok {
			return fmt.Errorf("%s not found", path)
		}
		for _, imp := range item.Archive.Imports {
			if err := collectDependencies(imp); err != nil {
				return err
			}
		}
		deps = append(deps, item.Archive)
		paths[item.Archive.ImportPath] = true
		return nil
	}
	if err := collectDependencies("runtime"); err != nil {
		return nil, err
	}
	for _, imp := range imports {
		if err := collectDependencies(imp); err != nil {
			return nil, err
		}
	}
	return deps, nil
}

// Updating is true if the update is in progress
func (s *ArchiveStore) Updating() bool {
	return s.updating
}

// Fresh is true if current cache matches the previously downloaded archives
func (s *ArchiveStore) Fresh() bool {
	return s.fresh
}

// Cache takes a snapshot of the cache (path -> hash)
func (s *ArchiveStore) Cache() map[string]CacheItem {
	cache := map[string]CacheItem{}
	for k, v := range s.cache {
		cache[k] = v
	}
	return cache
}

func (s *ArchiveStore) Handle(payload *flux.Payload) bool {
	switch a := payload.Action.(type) {
	case *actions.ImportsChanged:
		s.updateFresh(payload)
	case *actions.UpdateStart:
		s.app.Log("updating")
		s.updating = true
		s.index = nil
		s.complete = false

		var url string
		if strings.HasPrefix(dom.GetWindow().Document().DocumentURI(), "https://") {
			url = "wss://compile.jsgo.io/_pg/"
		} else {
			url = "ws://localhost:8081/_pg/"
		}

		s.app.Dispatch(&actions.Dial{
			Url:     url,
			Open:    func() flux.ActionInterface { return &actions.UpdateOpen{} },
			Message: func(m interface{}) flux.ActionInterface { return &actions.UpdateMessage{Message: m} },
			Close:   func() flux.ActionInterface { return &actions.UpdateClose{Run: a.Run} },
		})
		payload.Notify()

	case *actions.UpdateOpen:
		hashes := map[string]string{}
		for path, item := range s.Cache() {
			hashes[path] = item.Hash
		}
		message := messages.Update{
			Source: map[string]map[string]string{
				"main": s.app.Editor.Files(),
			},
			Cache: hashes,
		}
		s.app.Dispatch(&actions.Send{
			Message: message,
		})
	case *actions.UpdateMessage:
		switch message := a.Message.(type) {
		case messages.Queueing:
			if message.Position > 1 {
				s.app.Logf("queued position %d", message.Position)
			}
		case messages.Downloading:
			if message.Message != "" {
				s.app.Logf(message.Message)
			} else if message.Done {
				s.app.Logf("building")
			}
		case messages.Archive:
			r, err := gzip.NewReader(bytes.NewBuffer(message.Contents))
			if err != nil {
				s.app.Fail(err)
				return true
			}
			var a compiler.Archive
			if err := gob.NewDecoder(r).Decode(&a); err != nil {
				s.app.Fail(err)
				return true
			}
			s.cache[message.Path] = CacheItem{
				Hash:    message.Hash,
				Archive: &a,
			}
			s.app.Logf("caching %s", a.Name)
		case messages.Index:
			s.index = message
		}
	case *actions.UpdateClose:
		s.updating = false
		s.updateComplete(payload)
		s.updateFresh(payload)
		if !s.complete {
			s.app.Fail(errors.New("websocket closed but update incomplete"))
			return true
		}
		if !s.fresh {
			s.app.Fail(errors.New("websocket closed but archives not fresh"))
			return true
		}
		if a.Run {
			s.app.Dispatch(&actions.CompileStart{})
		} else {
			var downloaded, unchanged int
			for _, v := range s.index {
				if v.Unchanged {
					unchanged++
				} else {
					downloaded++
				}
			}
			if downloaded == 0 && unchanged == 0 {
				s.app.Log()
			} else if downloaded > 0 && unchanged > 0 {
				s.app.Logf("%d downloaded, %d unchanged", downloaded, unchanged)
			} else if downloaded > 0 {
				s.app.Logf("%d downloaded", downloaded)
			} else if unchanged > 0 {
				s.app.Logf("%d unchanged", unchanged)
			}
		}
		payload.Notify()
	}

	return true
}

// updateComplete sets s.complete to true if the index has been received and all the cached archives
// match the index. We run this every time an archive is received, when the index is received and also
// when the update websocket is closed.
func (s *ArchiveStore) updateComplete(payload *flux.Payload) {
	previous := s.complete
	complete := func() bool {
		if s.index == nil {
			return false
		}
		for path, item := range s.index {
			cached, ok := s.cache[path]
			if !ok {
				return false
			}
			if cached.Hash != item.Hash {
				return false
			}
		}
		return true
	}()
	if previous != complete {
		s.complete = complete
		payload.Notify()
	}
}

// updateFresh sets s.fresh to true if the imports from the editor are all in the current index
func (s *ArchiveStore) updateFresh(payload *flux.Payload) {
	previous := s.fresh
	fresh := func() bool {
		if !s.complete {
			return false
		}
		for _, p := range s.app.Scanner.Imports() {
			if _, ok := s.index[p]; !ok {
				return false
			}
		}
		return true
	}()
	if previous != fresh {
		s.fresh = fresh
		payload.Notify()
	}
}
