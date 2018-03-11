package main

import (
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/playground/stores"
	"github.com/dave/jsgo/playground/views"
	"github.com/gopherjs/vecty"
	"github.com/vincent-petithory/dataurl"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func main() {
	if document.ReadyState() == "loading" {
		document.AddEventListener("DOMContentLoaded", false, func(dom.Event) {
			go run()
		})
	} else {
		go run()
	}
}

func run() {

	vecty.AddStylesheet("https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/css/bootstrap.min.css")
	vecty.AddStylesheet(dataurl.New([]byte(Styles), "text/css").String())

	app := &stores.App{}
	app.Init()
	p := views.NewPage(app)
	vecty.RenderBody(p)

	app.Dispatch(&actions.Load{})
}

const Styles = `
	html, body {
		height: 100%;
	}
	.editor {
		height: calc(100% - 65px);
		width: 100%;
	}
	.split {
		height: 100%;
		width: 100%;
	}
	.gutter {
		height: 100%;
		background-color: #eee;
		background-repeat: no-repeat;
		background-position: 50%;
	}
	.gutter.gutter-horizontal {
		cursor: col-resize;
		background-image:  url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAeCAYAAADkftS9AAAAIklEQVQoU2M4c+bMfxAGAgYYmwGrIIiDjrELjpo5aiZeMwF+yNnOs5KSvgAAAABJRU5ErkJggg==')
	}
	.split {
		-webkit-box-sizing: border-box;
		-moz-box-sizing: border-box;
		box-sizing: border-box;
	}
	.split, .gutter.gutter-horizontal {
		float: left;
	}
	.preview {
		border: 0;
		height: 100%;
		width: 100%;
	}
`
