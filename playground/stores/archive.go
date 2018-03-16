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

	// state of the imports at the last compile
	imports []string

	// index of the previously received update (path -> hash for all dependent packages)
	index []messages.PlaygroundIndexItem

	dependencies []*compiler.Archive

	// is the cache up to date?
	complete bool
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
	for _, d := range s.dependencies {
		deps = append(deps, d)
	}
	return deps
}

func (s *ArchiveStore) Index() []messages.PlaygroundIndexItem {
	var index []messages.PlaygroundIndexItem
	for _, item := range s.index {
		index = append(index, item)
	}
	return index
}

// Updating is true if the update is in progress
func (s *ArchiveStore) Updating() bool {
	return s.updating
}

// Fresh is true if current cache matches the previously downloaded archives
func (s *ArchiveStore) Fresh(imports []string) bool {
	if !s.complete {
		return false
	}
	fromIndex := map[string]bool{}
	for _, v := range s.index {
		fromIndex[v.Path] = true
	}
	for _, p := range imports {
		if !fromIndex[p] {
			return false
		}
	}
	return true
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
	case *actions.UpdateStart:
		fmt.Println("dialing compile websocket open")
		s.updating = true
		s.imports = s.app.Scanner.Imports()
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
		message := messages.PlaygroundCompile{
			Source: map[string]map[string]string{
				"main": {
					"main.go": s.app.Editor.Text(),
				},
			},
			ArchiveCache: hashes,
		}
		s.app.Dispatch(&actions.Send{
			Message: message,
		})
	case *actions.UpdateMessage:
		switch message := a.Message.(type) {
		case messages.PlaygroundArchive, messages.PlaygroundIndex:
			switch message := message.(type) {
			case messages.PlaygroundArchive:
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
			case messages.PlaygroundIndex:
				s.index = message
			}
			if s.index != nil {
				fresh := true
				for _, item := range s.index {
					cached, ok := s.cache[item.Path]
					if !ok {
						fresh = false
						break
					}
					if cached.Hash != item.Hash {
						fresh = false
						break
					}
				}
				if fresh {
					s.complete = true
					var deps []*compiler.Archive
					for _, v := range s.index {
						a, ok := s.cache[v.Path]
						if !ok {
							s.app.Fail(fmt.Errorf("%s not found", v.Path))
							return true
						}
						deps = append(deps, a.Archive)
					}
					s.dependencies = deps
					payload.Notify()
				}
			}

		default:
			fmt.Printf("%T: %#v\n", message, message)
		}
	case *actions.UpdateClose:
		s.updating = false
		fmt.Println("compile websocket closed")
		payload.Notify()
	}

	return true
}
