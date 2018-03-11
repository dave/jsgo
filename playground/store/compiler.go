package store

import (
	"fmt"

	"strings"

	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/dispatcher"
	"github.com/dave/jsgo/server/messages"
	"honnef.co/go/js/dom"
)

var (
	Dependencies = map[string]string{}
)

func init() {
	dispatcher.Register(compilerActions)
}

func compilerActions(action interface{}) {
	switch a := action.(type) {
	case *actions.CompileStart:
		fmt.Println("dialing compile websocket open")

		var url string
		if strings.HasPrefix(dom.GetWindow().Document().DocumentURI(), "https://") {
			url = "wss://compile.jsgo.io/_pg/"
		} else {
			url = "ws://localhost:8081/_pg/"
		}

		err := conn.Dial(
			url,
			func() interface{} { return &actions.CompileOpen{} },
			func(m interface{}) interface{} { return &actions.CompileMessage{Message: m} },
			func() interface{} { return &actions.CompileClose{} },
			func(err error) interface{} { return &actions.Error{Err: err} },
		)
		if err != nil {
			panic(err)
		}
	case *actions.CompileOpen:
		fmt.Println("compile websocket open, sending compile init")
		message := messages.PlaygroundCompile{
			Source: map[string]map[string]string{
				"main": {
					"main.go": EditorText,
				},
			},
			Dependencies: map[string]string{},
		}
		if err := conn.Send(message); err != nil {
			panic(err)
		}
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

	Listeners.Fire()
}
