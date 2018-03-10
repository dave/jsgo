package components

import (
	"time"

	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/dispatcher"
	"github.com/dave/jsgo/playground/store"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
	"github.com/tulir/gopher-ace"
)

type Editor struct {
	vecty.Core

	Text   string `vecty:"prop"`
	editor ace.Editor
}

func NewEditor() *Editor {
	e := &Editor{}
	return e
}

func (e *Editor) Mount() {
	store.Listeners.Add(e, func() {
		e.Text = store.EditorText
		if e.Text != e.editor.GetValue() {
			// only update the editor if the text is changed
			e.editor.SetValue(e.Text)
			e.editor.ClearSelection()
			e.editor.MoveCursorTo(0, 0)
		}
	})
	e.editor = ace.Edit("editor")
	e.editor.SetOptions(map[string]interface{}{
		"mode": "ace/mode/golang",
	})
	if e.Text != "" {
		e.editor.SetValue(e.Text)
		e.editor.ClearSelection()
		e.editor.MoveCursorTo(0, 0)
	}
	var changes int
	e.editor.OnChange(func(ev *js.Object) {
		changes++
		before := changes
		go func() {
			<-time.After(time.Millisecond * 250)
			if before == changes {
				dispatcher.Dispatch(&actions.EditorTextChangedDebounced{
					Text: e.editor.GetValue(),
				})
			}
		}()
	})
}

func (e *Editor) Unmount() {
	store.Listeners.Remove(e)
}

func (e *Editor) Render() vecty.ComponentOrHTML {
	return elem.Div(
		vecty.Markup(
			prop.ID("editor"),
			vecty.Class("editor"),
		),
	)
}
