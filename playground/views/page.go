package views

import (
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/stores"
	"github.com/dave/splitter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
)

type Page struct {
	vecty.Core
	app *stores.App

	Sizes []float64 `vecty:"prop"`

	newItemTitle  string
	left, right   *vecty.HTML
	split         *splitter.Split
	compileButton *vecty.HTML
	optionsButton *vecty.HTML
}

func NewPage(app *stores.App) *Page {
	v := &Page{
		app: app,
	}
	return v
}

func (v *Page) Mount() {
	v.app.Watch(v, func(done chan struct{}) {
		defer close(done)
		v.Sizes = v.app.Editor.Sizes()
		v.split.SetSizesIfChanged(v.Sizes)

		// Only top-level page should fire vecty.Rerender
		vecty.Rerender(v)
	})

	v.split = splitter.New("split")
	v.split.Init(
		js.S{"#left", "#right"},
		js.M{
			"sizes": v.Sizes,
			"onDragEnd": func() {
				v.app.Dispatch(&actions.UserChangedSplit{
					Sizes: v.split.GetSizes(),
				})
			},
		},
	)
}

func (v *Page) Unmount() {
	v.app.Delete(v)
}

func (v *Page) Render() vecty.ComponentOrHTML {

	v.left = v.renderLeft()
	v.right = v.renderRight()

	return elem.Body(
		elem.Div(
			vecty.Markup(
				vecty.Class("container-fluid", "p-0", "split", "split-horizontal"),
			),
			v.left,
			v.right,
		),
	)
}

func (v *Page) renderLeft() *vecty.HTML {
	return elem.Div(
		vecty.Markup(
			prop.ID("left"),
			vecty.Class("split"),
		),
		v.renderHeader(),
		NewEditor(v.app),
	)
}

func (v *Page) renderHeader() *vecty.HTML {

	return elem.Navigation(
		vecty.Markup(
			vecty.Class("navbar", "navbar-expand", "navbar-light", "bg-light"),
		),
		elem.UnorderedList(
			vecty.Markup(
				vecty.Class("navbar-nav", "mr-auto"),
			),
			elem.ListItem(
				vecty.Markup(
					vecty.Class("nav-item", "dropdown"),
				),
				elem.Anchor(
					vecty.Markup(
						prop.ID("packageDropdown"),
						prop.Href(""),
						vecty.Class("nav-link", "dropdown-toggle"),
						vecty.Property("role", "button"),
						vecty.Data("toggle", "dropdown"),
						vecty.Property("aria-haspopup", "true"),
						vecty.Property("aria-expanded", "false"),
						event.Click(func(ev *vecty.Event) {}).PreventDefault(),
					),
					vecty.Text("main"),
				),
				elem.Div(
					vecty.Markup(
						vecty.Class("dropdown-menu"),
						vecty.Property("aria-labelledby", "packageDropdown"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {}).PreventDefault(),
						),
						vecty.Text("Package 1"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {}).PreventDefault(),
						),
						vecty.Text("Package 2"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {}).PreventDefault(),
						),
						vecty.Text("Package 3"),
					),
					elem.Div(
						vecty.Markup(
							vecty.Class("dropdown-divider"),
						),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {}).PreventDefault(),
						),
						vecty.Text("Add package..."),
					),
				),
			),
			elem.ListItem(
				vecty.Markup(
					vecty.Class("nav-item", "dropdown"),
				),
				elem.Anchor(
					vecty.Markup(
						prop.ID("fileDropdown"),
						prop.Href(""),
						vecty.Class("nav-link", "dropdown-toggle"),
						vecty.Property("role", "button"),
						vecty.Data("toggle", "dropdown"),
						vecty.Property("aria-haspopup", "true"),
						vecty.Property("aria-expanded", "false"),
						event.Click(func(ev *vecty.Event) {}).PreventDefault(),
					),
					vecty.Text("main.go"),
				),
				elem.Div(
					vecty.Markup(
						vecty.Class("dropdown-menu"),
						vecty.Property("aria-labelledby", "fileDropdown"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {}).PreventDefault(),
						),
						vecty.Text("File 1"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {}).PreventDefault(),
						),
						vecty.Text("File 2"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {}).PreventDefault(),
						),
						vecty.Text("File 3"),
					),
					elem.Div(
						vecty.Markup(
							vecty.Class("dropdown-divider"),
						),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {}).PreventDefault(),
						),
						vecty.Text("Add file..."),
					),
				),
			),
		),
		elem.UnorderedList(
			vecty.Markup(
				vecty.Class("navbar-nav", "mx-auto"),
			),
			elem.Span(
				vecty.Markup(
					vecty.Class("navbar-text"),
					prop.ID("message"),
				),
				vecty.Text(""),
			),
		),
		elem.UnorderedList(
			vecty.Markup(
				vecty.Class("navbar-nav", "ml-auto"),
			),
			elem.ListItem(
				vecty.Markup(
					vecty.Class("nav-item", "btn-group"),
				),
				elem.Button(
					vecty.Markup(
						vecty.Property("type", "button"),
						vecty.Class("btn", "btn-primary"),
						event.Click(func(e *vecty.Event) {
							if v.app.Archive.Updating() {
								return
							} else if v.app.Archive.Fresh() {
								v.app.Dispatch(&actions.CompileStart{})
							} else {
								v.app.Dispatch(&actions.UpdateStart{
									Run: true,
								})
							}
						}).PreventDefault(),
						vecty.Property("disabled", v.app.Archive.Updating()),
					),
					vecty.Text("Run"),
				),
				elem.Button(
					vecty.Markup(
						vecty.Property("type", "button"),
						vecty.Data("toggle", "dropdown"),
						vecty.Property("aria-haspopup", "true"),
						vecty.Property("aria-expanded", "false"),
						vecty.Class("btn", "btn-primary", "dropdown-toggle", "dropdown-toggle-split"),
					),
					elem.Span(vecty.Markup(vecty.Class("sr-only")), vecty.Text("Options")),
				),
				elem.Div(
					vecty.Markup(
						vecty.Class("dropdown-menu", "dropdown-menu-right"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {
								v.app.Dispatch(&actions.UpdateStart{})
							}).PreventDefault(),
						),
						vecty.Text("Update"),
					),

					elem.Div(
						vecty.Markup(
							vecty.Class("dropdown-divider"),
						),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {
								js.Global.Call("alert", "TODO")
							}).PreventDefault(),
						),
						vecty.Text("Format code"),
					),
					elem.Div(
						vecty.Markup(
							vecty.Class("dropdown-divider"),
						),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {
								js.Global.Call("alert", "TODO")
							}).PreventDefault(),
						),
						vecty.Text("Build tags..."),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {
								js.Global.Call("alert", "TODO")
							}).PreventDefault(),
						),
						vecty.Text("Save"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href(""),
							event.Click(func(e *vecty.Event) {
								js.Global.Call("alert", "TODO")
							}).PreventDefault(),
						),
						vecty.Text("Deploy"),
					),
				),
			),
		),
	)
}

func (v *Page) renderRight() *vecty.HTML {
	return elem.Div(
		vecty.Markup(
			prop.ID("right"),
			vecty.Class("split"),
		),
		elem.Div(
			vecty.Markup(
				prop.ID("iframe-holder"),
			),
		),
	)
}
