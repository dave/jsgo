package views

import (
	"fmt"

	"strings"

	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/stores"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

type AddFileModal struct {
	vecty.Core
	app *stores.App
}

func NewAddFileModal(app *stores.App) *AddFileModal {
	v := &AddFileModal{
		app: app,
	}
	return v
}

func (v *AddFileModal) Mount() {
	js.Global.Call("$", "#add-file").Call("on", "hide.bs.modal", func(ev *vecty.Event) {
		v.app.Dispatch(&actions.UsedClosedAddFileModal{})
	})
}

func (v *AddFileModal) Render() vecty.ComponentOrHTML {
	input := elem.Input(
		vecty.Markup(
			prop.Type(prop.TypeText),
			vecty.Class("form-control"),
			prop.ID("file-name"),
		),
	)
	return Modal(
		"add-file",
		func(ev *vecty.Event) {
			value := input.Node().Get("value").String()
			if strings.Contains(value, "/") {
				v.app.Fail(fmt.Errorf("filename %s must not contain a slash", value))
				return
			}
			if !strings.HasSuffix(value, ".go") {
				v.app.Fail(fmt.Errorf("filename %s must end .go", value))
				return
			}
			for name := range v.app.Editor.Files() {
				if name == value {
					v.app.Fail(fmt.Errorf("%s already exists", value))
					return
				}
			}
			v.app.Dispatch(&actions.AddFile{
				Name: value,
			})
			v.app.Dispatch(&actions.CloseAddFileModal{})
		},
	).Body(
		elem.Form(
			elem.Div(
				vecty.Markup(vecty.Class("form-group")),
				elem.Label(
					vecty.Markup(
						vecty.Property("for", "file-name"),
						vecty.Class("col-form-label"),
					),
					vecty.Text("Filename"),
				),
				input,
			),
		),
	).Build()

}
