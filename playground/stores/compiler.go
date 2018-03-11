package stores

import (
	"fmt"

	"strings"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/server/messages"
	"honnef.co/go/js/dom"
)

type CompilerStore struct {
	app *App

	Dependencies map[string]string
}

func NewCompilerStore(app *App) *CompilerStore {
	s := &CompilerStore{
		app:          app,
		Dependencies: map[string]string{},
	}
	return s
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
					"main.go": s.app.Editor.EditorText,
				},
			},
			Dependencies: map[string]string{},
		}
		s.app.Dispatch(&actions.Send{
			Message: message,
		})
	case *actions.CompileMessage:
		switch message := a.Message.(type) {
		case messages.Complete:
			fmt.Println("compile complete")
		default:
			fmt.Printf("%T\n", message)
		}
	case *actions.CompileClose:
		fmt.Println("compile websocket closed")
	}

	return true
}
