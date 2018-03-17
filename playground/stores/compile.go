package stores

import (
	"context"

	"github.com/dave/flux"
	"github.com/dave/jsgo/builder"
	"github.com/dave/jsgo/builderjs"
	"github.com/dave/jsgo/playground/actions"
	"github.com/gopherjs/gopherjs/compiler"
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

	// fast compile if possible?
	fast bool

	// imports at last full compile
	imports []string

	// all current imports were imported at last compile
	fresh bool
}

func (s *CompileStore) Fresh() bool {
	return s.fresh
}

func (s *CompileStore) Fast() bool {
	return s.fast
}

func (s *CompileStore) updateFresh(payload *flux.Payload) {
	previous := s.fresh
	fresh := func() bool {
		fromPrevious := map[string]bool{}
		for _, v := range s.imports {
			fromPrevious[v] = true
		}
		for _, p := range s.app.Scanner.Imports() {
			if !fromPrevious[p] {
				return false
			}
		}
		return true
	}()
	if previous != fresh {
		s.fresh = fresh
		payload.Notify()
	}
}

func (s *CompileStore) Handle(payload *flux.Payload) bool {
	switch a := payload.Action.(type) {
	case *actions.UserChangedText:
		payload.Wait(s.app.Scanner)
		s.updateFresh(payload)
	case *actions.FastCompileCheckbox:
		s.fast = a.Value
		payload.Notify()
	case *actions.CompileStart:

		full := true
		if s.fast && s.fresh {
			full = false
		} else {
			s.imports = s.app.Scanner.Imports()
		}
		s.updateFresh(payload)

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
		if full {
			deps = append(deps, archive)
		} else {
			deps = []*compiler.Archive{archive}
		}

		doc := dom.GetWindow().Document()
		var frame *dom.HTMLIFrameElement

		if full {
			holder := doc.GetElementByID("iframe-holder")
			for _, v := range holder.ChildNodes() {
				v.Underlying().Call("remove")
			}
			frame = doc.CreateElement("iframe").(*dom.HTMLIFrameElement)
			frame.SetID("iframe")
			frame.Style().Set("width", "100%")
			frame.Style().Set("height", "100%")
			frame.Style().Set("border", "0")
			holder.AppendChild(frame)
		} else {
			frame = doc.GetElementByID("iframe").(*dom.HTMLIFrameElement)
		}

		content := frame.ContentDocument()
		head := content.GetElementsByTagName("head")[0].(*dom.BasicHTMLElement)

		if full {
			script := doc.CreateElement("script")
			script.SetInnerHTML(prelude.Prelude)
			head.AppendChild(script)
		}

		for _, d := range deps {
			code, _, err := builder.GetPackageCode(context.Background(), d, false, false)
			if err != nil {
				s.app.Fail(err)
				return true
			}

			script := doc.CreateElement("script")
			script.SetInnerHTML(string(code))
			head.AppendChild(script)
		}

		var initCode string
		if full {
			initCode = `
				$mainPkg = $packages["main"];
				$synthesizeMethods();
				$packages["runtime"].$init();
				$go($mainPkg.$init, []);
				$flushConsole();
			`

		} else {
			initCode = `
				$mainPkg = $packages["main"];
				$go($mainPkg.$init, []);
				$flushConsole();
			`
		}
		script1 := doc.CreateElement("script")
		script1.SetInnerHTML(initCode)
		head.AppendChild(script1)

	}
	return true
}
