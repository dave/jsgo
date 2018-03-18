package views

import (
	"fmt"

	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/stores"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
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
	return Modal(
		"add-file",
		func(ev *vecty.Event) {
			fmt.Println("Save")
			v.app.Dispatch(&actions.CloseAddFileModal{})
		},
	).Body(
		elem.Paragraph(vecty.Text("Foo")),
	).Build()

}
