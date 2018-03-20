package views

import (
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/stores"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/prop"
)

type DeleteFileModal struct {
	vecty.Core
	app *stores.App
	sel *vecty.HTML
}

func NewDeleteFileModal(app *stores.App) *DeleteFileModal {
	v := &DeleteFileModal{
		app: app,
	}
	return v
}

func (v *DeleteFileModal) Render() vecty.ComponentOrHTML {
	items := []vecty.MarkupOrChild{
		vecty.Markup(
			vecty.Class("form-control"),
			prop.ID("delete-file-select"),
		),
	}
	for _, name := range v.app.Editor.Filenames() {
		items = append(items,
			elem.Option(
				vecty.Markup(
					prop.Value(name),
					vecty.Property("selected", v.app.Editor.Current() == name),
				),
				vecty.Text(name),
			),
		)
	}
	v.sel = elem.Select(items...)

	return Modal(
		"Delete file...",
		"delete-file-modal",
		v.action,
	).Body(
		elem.Form(
			elem.Div(
				vecty.Markup(
					vecty.Class("form-group"),
				),
				elem.Label(
					vecty.Markup(
						vecty.Property("for", "delete-file-select"),
						vecty.Class("col-form-label"),
					),
					vecty.Text("File"),
				),
				v.sel,
			),
		),
	).Build()
}

func (v *DeleteFileModal) action(*vecty.Event) {
	n := v.sel.Node()
	i := n.Get("selectedIndex").Int()
	value := n.Get("options").Index(i).Get("value").String()
	v.app.Dispatch(&actions.DeleteFile{
		Name: value,
	})
}
