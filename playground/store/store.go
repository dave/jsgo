package store

import (
	"fmt"

	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/connection"
	"github.com/dave/jsgo/playground/dispatcher"
	"github.com/dave/jsgo/playground/store/storeutil"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/locstor"
)

var (
	SplitSizes []float64
	EditorText string

	// Listeners is the listeners that will be invoked when the store changes.
	Listeners = storeutil.NewListenerRegistry()

	stor = locstor.NewDataStore(locstor.JSONEncoding)
	conn = connection.New()
)

func init() {
	dispatcher.Register(onAction)
	go func() {
		for message := range conn.Receive {
			dispatcher.Dispatch(message)
		}
	}()
}

func onAction(action interface{}) {
	switch a := action.(type) {
	case *actions.Compile:
		fmt.Println("dialing compile websocket open")
		err := conn.Dial(
			"ws://localhost:8081/_pg/foo/",
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
	case *actions.Load:

		found, err := stor.Find("split-sizes", &SplitSizes)
		if err != nil {
			panic(err)
		}
		if !found {
			SplitSizes = defaultSplitSizes
		}

		found, err = stor.Find("editor-text", &EditorText)
		if err != nil {
			panic(err)
		}
		if !found {
			EditorText = defaultEditorText
		}

	case *actions.SplitChange:
		SplitSizes = a.Sizes
		if err := stor.Save("split-sizes", SplitSizes); err != nil {
			panic(err)
		}
	case *actions.EditorTextChangedDebounced:
		EditorText = a.Text
		if err := stor.Save("editor-text", EditorText); err != nil {
			panic(err)
		}
	}

	Listeners.Fire()
}

var (
	defaultSplitSizes = []float64{75, 25}
	defaultEditorText = `package main

import (
	"honnef.co/go/js/dom"
)

func main() {
	body := dom.GetWindow().Document().GetElementsByTagName("body")[0]
	body.SetInnerHTML("Hello, World!")
}`
)
