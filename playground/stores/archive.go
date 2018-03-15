package stores

import (
	"fmt"

	"strings"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/server/messages"
	"honnef.co/go/js/dom"
)

type ArchiveStore struct {
	app *App

	// is the update in progress?
	updating bool

	// cache (path -> hash) of all the archives cached in local storage
	cache map[string]string

	// state of the imports at the last compile
	imports []string

	// index of the previously received update (path -> hash for all dependent packages)
	index []messages.PlaygroundIndexItem

	// is the cache up to date?
	complete bool
}

func NewArchiveStore(app *App) *ArchiveStore {
	s := &ArchiveStore{
		app:   app,
		cache: map[string]string{},
	}
	return s
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
func (s *ArchiveStore) Cache() map[string]string {
	cache := map[string]string{}
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
		message := messages.PlaygroundCompile{
			Source: map[string]map[string]string{
				"main": {
					"main.go": s.app.Editor.Text(),
				},
			},
			ArchiveCache: s.Cache(),
		}
		s.app.Dispatch(&actions.Send{
			Message: message,
		})
	case *actions.UpdateMessage:
		switch message := a.Message.(type) {
		case messages.PlaygroundArchive, messages.PlaygroundIndex:
			switch message := message.(type) {
			case messages.PlaygroundArchive:
				payload.Wait(s.app.Local)
				s.cache[message.Path] = message.Hash
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
					if cached != item.Hash {
						fresh = false
						break
					}
				}
				if fresh {
					s.complete = true
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
