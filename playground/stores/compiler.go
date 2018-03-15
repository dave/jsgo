package stores

import (
	"errors"
	"fmt"

	"strings"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/server/messages"
	"honnef.co/go/js/dom"
)

type CompilerStore struct {
	app *App

	dependencies map[string]string
	cache        map[string]string
}

func NewCompilerStore(app *App) *CompilerStore {
	s := &CompilerStore{
		app:          app,
		dependencies: map[string]string{},
		cache:        map[string]string{},
	}
	return s
}

// Cache takes a snapshot of the cache
func (s *CompilerStore) Cache() map[string]string {
	cache := map[string]string{}
	for k, v := range s.cache {
		cache[k] = v
	}
	return cache
}

func (s *CompilerStore) Handle(payload *flux.Payload) bool {
	switch a := payload.Action.(type) {
	case *actions.CompileStart:
		fmt.Println("dialing compile websocket open")

		var url string
		if strings.HasPrefix(dom.GetWindow().Document().DocumentURI(), "https://") {
			url = "wss://compile.jsgo.io/_pg/"
		} else {
			url = "ws://localhost:8081/_pg/"
		}

		s.app.Dispatch(&actions.Dial{
			Url:     url,
			Open:    func() flux.ActionInterface { return &actions.CompileOpen{} },
			Message: func(m interface{}) flux.ActionInterface { return &actions.CompileMessage{Message: m} },
			Close:   func() flux.ActionInterface { return &actions.CompileClose{} },
		})
	case *actions.CompileOpen:
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
	case *actions.CompileMessage:
		switch message := a.Message.(type) {
		case messages.PlaygroundArchive:
			payload.Wait(s.app.Local)
			s.cache[message.Path] = message.Hash

		case messages.PlaygroundIndex:
			// check that the cache includes all the current versions of the dependencies
			cache := s.Cache()
			for _, v := range message {
				cached, ok := cache[v.Path]
				if !ok {
					s.app.Fail(errors.New("cache missing - maybe messages out of order"))
					return true
				}
				if cached != v.Hash {
					s.app.Fail(errors.New("cache wrong version - maybe messages out of order"))
					return true
				}
			}
			fmt.Println("running js?")

		default:
			fmt.Printf("%T: %#v\n", message, message)
		}
	case *actions.CompileClose:
		fmt.Println("compile websocket closed")
	}

	return true
}
