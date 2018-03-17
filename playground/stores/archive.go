package stores

import (
	"fmt"

	"strings"

	"bytes"

	"encoding/gob"

	"compress/gzip"

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

	// cache (path -> hash) of all the archives cached in local storage
	cache map[string]CacheItem

	// index of the previously received update (path -> hash for all dependent packages)
	index []messages.IndexItem

	// is the cache up to date?
	complete bool

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

func (s *ArchiveStore) Dependencies() []*compiler.Archive {
	var deps []*compiler.Archive
	for _, v := range s.index {
		a, ok := s.cache[v.Path]
		if !ok {
			s.app.Fail(fmt.Errorf("%s not found", v.Path))
			return nil
		}
		deps = append(deps, a.Archive)
	}
	return deps
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
	case *actions.UserChangedText:
		payload.Wait(s.app.Scanner)
		s.updateFresh(payload)
	case *actions.UpdateStart:
		fmt.Println("dialing compile websocket open")
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
			Close:   func() flux.ActionInterface { return &actions.UpdateClose{} },
		})
		payload.Notify()

	case *actions.UpdateOpen:
		fmt.Println("compile websocket open, sending compile init")
		hashes := map[string]string{}
		for path, item := range s.Cache() {
			hashes[path] = item.Hash
		}
		message := messages.Update{
			Source: map[string]map[string]string{
				"main": {
					"main.go": s.app.Editor.Text(),
				},
			},
			Cache: hashes,
		}
		s.app.Dispatch(&actions.Send{
			Message: message,
		})
	case *actions.UpdateMessage:
		switch message := a.Message.(type) {
		case messages.Archive:
			fmt.Printf("%T: %#v\n", message, messages.Archive{Path: message.Path, Hash: message.Hash})
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
			s.updateComplete(payload)
			s.updateFresh(payload)
		case messages.Index:
			s.index = message
			s.updateComplete(payload)
			s.updateFresh(payload)
			fmt.Printf("%T: %#v\n", message, message)
		default:
			fmt.Printf("%T: %#v\n", message, message)
		}
	case *actions.UpdateClose:
		s.updating = false
		s.updateComplete(payload)
		s.updateFresh(payload)
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
		for _, item := range s.index {
			cached, ok := s.cache[item.Path]
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
		fromIndex := map[string]bool{}
		for _, v := range s.index {
			fromIndex[v.Path] = true
		}
		for _, p := range s.app.Scanner.Imports() {
			if !fromIndex[p] {
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
