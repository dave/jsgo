package stores

import (
	"go/parser"
	"go/token"

	"strconv"

	"sort"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
)

func NewScannerStore(app *App) *ScannerStore {
	s := &ScannerStore{
		app:     app,
		imports: map[string][]string{},
	}
	return s
}

type ScannerStore struct {
	app     *App
	imports map[string][]string
}

// Imports returns all the imports from all files
func (s *ScannerStore) Imports() []string {
	var a []string
	for _, f := range s.imports {
		for _, i := range f {
			a = append(a, i)
		}
	}
	return a
}

func (s *ScannerStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.UserChangedText:
		if s.refresh(s.app.Editor.Current(), action.Text) {
			s.app.Dispatch(&actions.ImportsChanged{})
			payload.Notify()
		}
	case *actions.LoadFiles:
		payload.Wait(s.app.Editor)
		var changed bool
		for name, contents := range s.app.Editor.Files() {
			if s.refresh(name, contents) {
				changed = true
			}
		}
		if changed {
			s.app.Dispatch(&actions.ImportsChanged{})
			payload.Notify()
		}
	}
	return true
}

func (s *ScannerStore) refresh(filename, contents string) bool {
	fset := token.NewFileSet()

	// ignore errors
	f, _ := parser.ParseFile(fset, filename, contents, parser.ImportsOnly)

	var imports []string
	for _, v := range f.Imports {
		// ignore errors
		unquoted, _ := strconv.Unquote(v.Path.Value)
		imports = append(imports, unquoted)
	}
	sort.Strings(imports)

	if s.changed(s.imports[filename], imports) {
		s.imports[filename] = imports
		s.app.Debug("Imports", s.imports)
		return true
	}
	return false
}

func (s *ScannerStore) changed(imports, compare []string) bool {
	if len(compare) != len(imports) {
		return true
	}
	for i := range compare {
		if imports[i] != compare[i] {
			return true
		}
	}
	return false
}
