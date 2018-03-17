package stores

import (
	"context"

	"fmt"

	"github.com/dave/flux"
	"github.com/dave/jsgo/builder"
	"github.com/dave/jsgo/builderjs"
	"github.com/dave/jsgo/playground/actions"
	"github.com/gopherjs/gopherjs/compiler/prelude"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

func NewCompileStore(app *App) *CompileStore {
	s := &CompileStore{
		app: app,
	}
	return s
}

type CompileStore struct {
	app *App
}

func (s *CompileStore) Handle(payload *flux.Payload) bool {
	switch payload.Action.(type) {
	case *actions.CompileStart:
		s.compile()
	}
	return true
}

// requestAnimationFrame calls the native JS function of the same name.
func requestAnimationFrame(callback func(float64)) int {
	return js.Global.Call("requestAnimationFrame", callback).Int()
}

func (s *CompileStore) compile() {

	deps, err := s.app.Archive.Collect(s.app.Scanner.Imports())
	if err != nil {
		s.app.Fail(err)
		return
	}
	archive, err := builderjs.BuildPackage(
		map[string]string{"main.go": s.app.Editor.Text()},
		deps,
		false,
	)
	if err != nil {
		s.app.Fail(err)
		return
	}
	deps = append(deps, archive)

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

	fmt.Println("Injecting prelude")
	scriptPrelude := doc.CreateElement("script")
	scriptPrelude.SetInnerHTML(prelude.Prelude)
	head.AppendChild(scriptPrelude)

	for _, d := range deps {
		fmt.Println("Injecting", d.ImportPath)
		code, _, err := builder.GetPackageCode(context.Background(), d, false, false)
		if err != nil {
			s.app.Fail(err)
			return
		}
		scriptDep := doc.CreateElement("script")
		scriptDep.SetInnerHTML(string(code))
		head.AppendChild(scriptDep)
	}

	fmt.Println("Injecting initializer")
	scriptInit := doc.CreateElement("script")
	scriptInit.SetInnerHTML(`
		$mainPkg = $packages["main"];
		$synthesizeMethods();
		$packages["runtime"].$init();
		$go($mainPkg.$init, []);
		$flushConsole();
	`)
	head.AppendChild(scriptInit)
}
