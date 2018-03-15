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
	app *App
}

func (s *CompileStore) Handle(payload *flux.Payload) bool {
	switch payload.Action.(type) {
	case *actions.CompileStart:
		deps := s.app.Archive.Dependencies()
		archive, err := builderjs.BuildPackage(
			map[string]string{"main.go": s.app.Editor.Text()},
			deps,
			false,
		)
		if err != nil {
			s.app.Fail(err)
			return true
		}
		deps = append(deps, archive)

		doc := dom.GetWindow().Document()
		frame := doc.GetElementByID("iframe").(*dom.HTMLIFrameElement)
		frame.SetInnerHTML("")

		script := doc.CreateElement("script")
		script.SetInnerHTML(prelude.Prelude)
		frame.AppendChild(script)

		for _, d := range deps {
			code, _, err := builder.GetPackageCode(context.Background(), d, false, false)
			if err != nil {
				s.app.Fail(err)
				return true
			}

			script := doc.CreateElement("script")
			script.SetInnerHTML(string(code))
			frame.AppendChild(script)
		}

		script1 := doc.CreateElement("script")
		script1.SetInnerHTML(`
			$mainPkg = $packages["main"];
			$synthesizeMethods();
			$packages["runtime"].$init();
			$go($mainPkg.$init, []);
			$flushConsole();
		`)
		frame.AppendChild(script1)

	}
	return true
}
