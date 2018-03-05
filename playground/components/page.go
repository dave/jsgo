package components

import (
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/dispatcher"
	"github.com/dave/jsgo/playground/splitter"
	"github.com/dave/jsgo/playground/store"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
	"github.com/gopherjs/vecty/event"
	"github.com/gopherjs/vecty/prop"
)

type Page struct {
	vecty.Core

	Sizes         []float64 `vecty:"prop"`
	newItemTitle  string
	left, right   *vecty.HTML
	split         *splitter.Split
	compileButton *vecty.HTML
	optionsButton *vecty.HTML
}

func NewPage() *Page {
	p := &Page{}
	store.Listeners.Add(p, func() {
		p.Sizes = store.SplitSizes
		vecty.Rerender(p)
	})
	return p
}

func (p *Page) Mount() {
	p.split = splitter.New("split")
	p.split.Init(
		js.S{"#left", "#right"},
		js.M{
			"sizes": p.Sizes,
			"onDragEnd": func() {
				dispatcher.Dispatch(&actions.SplitChange{
					Sizes: p.split.GetSizes(),
				})
			},
		},
	)
}

func (p *Page) Unmount() {
	store.Listeners.Remove(p)
}

func (p *Page) onCompile(event *vecty.Event) {
	dispatcher.Dispatch(&actions.Compile{})
}

func (p *Page) Render() vecty.ComponentOrHTML {

	p.left = p.renderLeft()
	p.right = p.renderRight()

	return elem.Body(
		elem.Div(
			vecty.Markup(
				vecty.Class("container-fluid", "p-0", "split", "split-horizontal"),
			),
			p.left,
			p.right,
		),
	)
}

func (p *Page) renderLeft() *vecty.HTML {
	return elem.Div(
		vecty.Markup(
			prop.ID("left"),
			vecty.Class("split"),
		),
		p.renderHeader(),
		NewEditor(),
	)
}

func (p *Page) renderHeader() *vecty.HTML {

	return elem.Navigation(
		vecty.Markup(
			vecty.Class("navbar", "navbar-expand", "navbar-light", "bg-light"),
		),
		elem.Div(
			vecty.Markup(
				prop.ID("navbarSupportedContent"),
				vecty.Class("collapse", "navbar-collapse"),
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
						event.Click(p.onCompile).PreventDefault(),
					),
					vecty.Text("Compile"),
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
							prop.Href("#"),
							event.Click(func(e *vecty.Event) {
								js.Global.Call("alert", "TODO")
							}).StopPropagation(),
						),
						elem.Input(
							vecty.Markup(
								prop.Type(prop.TypeCheckbox),
								vecty.Class("form-check-input", "dropdown-item"),
								prop.ID("dropdownCheckDeps"),
							),
						),
						elem.Label(
							vecty.Markup(
								vecty.Class("form-check-label"),
								prop.For("dropdownCheckDeps"),
							),
							vecty.Text("Update imports"),
						),
					),
					elem.Div(
						vecty.Markup(
							vecty.Class("dropdown-divider"),
						),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href("#"),
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
							prop.Href("#"),
							event.Click(func(e *vecty.Event) {
								js.Global.Call("alert", "TODO")
							}).PreventDefault(),
						),
						vecty.Text("Build tags..."),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href("#"),
							event.Click(func(e *vecty.Event) {
								js.Global.Call("alert", "TODO")
							}).PreventDefault(),
						),
						vecty.Text("Save"),
					),
					elem.Anchor(
						vecty.Markup(
							vecty.Class("dropdown-item"),
							prop.Href("#"),
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

func (p *Page) renderRight() *vecty.HTML {
	return elem.Div(
		vecty.Markup(
			prop.ID("right"),
			vecty.Class("split"),
		),
	)
}
