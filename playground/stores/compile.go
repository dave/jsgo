package stores

import (
	"context"

	"github.com/dave/flux"
	"github.com/dave/jsgo/builder"
	"github.com/dave/jsgo/builderjs"
	"github.com/dave/jsgo/playground/actions"
	"github.com/gopherjs/gopherjs/compiler/prelude"
	"honnef.co/go/js/dom"
)

func NewCompileStore(app *App) *CompileStore {
	s := &CompileStore{
		app: app,
	}
	return s
}

type CompileStore struct {
	app       *App
	compiling bool
}

func (s *CompileStore) Compiling() bool {
	return s.compiling
}

func (s *CompileStore) Handle(payload *flux.Payload) bool {
	switch payload.Action.(type) {
	case *actions.CompileStart:
		s.compiling = true
		s.compile()
		s.compiling = false
	}
	return true
}

func (s *CompileStore) compile() {
	s.app.Log("compiling")

	deps, err := s.app.Archive.Collect(s.app.Scanner.Imports())
	if err != nil {
		s.app.Fail(err)
		return
	}
	archive, err := builderjs.BuildPackage(
		s.app.Editor.Files(),
		deps,
		false,
	)
	if err != nil {
		s.app.Fail(err)
		return
	}
	deps = append(deps, archive)

	s.app.Log("running")

	doc := dom.GetWindow().Document()
	holder := doc.GetElementByID("iframe-holder")
	for _, v := range holder.ChildNodes() {
		v.Underlying().Call("remove")
	}
	frame := doc.CreateElement("iframe").(*dom.HTMLIFrameElement)
	frame.SetID("iframe")
	frame.Style().Set("width", "100%")
	frame.Style().Set("height", "100%")
	frame.Style().Set("border", "0")
	holder.AppendChild(frame)

	content := frame.ContentDocument()
	head := content.GetElementsByTagName("head")[0].(*dom.BasicHTMLElement)

	scriptPrelude := doc.CreateElement("script")
	scriptPrelude.SetInnerHTML(prelude.Prelude)
	head.AppendChild(scriptPrelude)

	for _, d := range deps {
		code, _, err := builder.GetPackageCode(context.Background(), d, false, false)
		if err != nil {
			s.app.Fail(err)
			return
		}
		scriptDep := doc.CreateElement("script")
		scriptDep.SetInnerHTML(string(code))
		head.AppendChild(scriptDep)
	}

	scriptInit := doc.CreateElement("script")
	scriptInit.SetInnerHTML(`
		$mainPkg = $packages["main"];
		$synthesizeMethods();
		$packages["runtime"].$init();
		$go($mainPkg.$init, []);
		$flushConsole();
	`)
	head.AppendChild(scriptInit)

	s.app.Log()
}
