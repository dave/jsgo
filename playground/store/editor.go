package store

import (
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/dispatcher"
)

var (
	SplitSizes []float64
	EditorText string
)

func init() {
	dispatcher.Register(editorActions)
}

func editorActions(action interface{}) {
	switch a := action.(type) {
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
