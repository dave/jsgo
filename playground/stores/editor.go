package stores

import (
	"sort"

	"errors"

	"go/format"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
	"github.com/gopherjs/gopherjs/js"
)

func NewEditorStore(app *App) *EditorStore {
	s := &EditorStore{
		app:   app,
		files: map[string]string{},
	}
	return s
}

type EditorStore struct {
	app *App

	sizes   []float64
	files   map[string]string
	current string
}

func (s *EditorStore) Sizes() []float64 {
	return s.sizes
}

func (s *EditorStore) Text() string {
	return s.files[s.current]
}

func (s *EditorStore) Current() string {
	return s.current
}

func (s *EditorStore) Files() map[string]string {
	f := map[string]string{}
	for k, v := range s.files {
		f[k] = v
	}
	return f
}

func (s *EditorStore) Filenames() []string {
	var f []string
	for k := range s.files {
		f = append(f, k)
	}
	sort.Strings(f)
	return f
}

func (s *EditorStore) Handle(payload *flux.Payload) bool {
	switch a := payload.Action.(type) {

	case *actions.ChangeSplit:
		s.sizes = a.Sizes
		payload.Notify()
	case *actions.ChangeText:
		s.files[s.current] = a.Text
		payload.Notify()
	case *actions.ChangeFile:
		s.current = a.Name
		payload.Notify()
	case *actions.UserChangedSplit:
		s.sizes = a.Sizes
	case *actions.UserChangedText:
		s.files[s.current] = a.Text
	case *actions.UserChangedFile:
		s.current = a.Name
		payload.Notify()
	case *actions.AddFileClick:
		js.Global.Call("$", "#add-file-modal").Call("modal", "show")
		js.Global.Call("$", "#add-file-input").Call("focus")
		payload.Notify()
	case *actions.DeleteFileClick:
		js.Global.Call("$", "#delete-file-modal").Call("modal", "show")
		js.Global.Call("$", "#delete-file-input").Call("focus")
		payload.Notify()
	case *actions.AddFile:
		js.Global.Call("$", "#add-file-modal").Call("modal", "hide")
		s.files[a.Name] = ""
		s.current = a.Name
		payload.Notify()
	case *actions.DeleteFile:
		js.Global.Call("$", "#delete-file-modal").Call("modal", "hide")
		if len(s.files) == 1 {
			s.app.Fail(errors.New("can't delete last file"))
			return true
		}
		delete(s.files, a.Name)
		if s.current == a.Name {
			s.current = s.Filenames()[0]
		}
		payload.Notify()
	case *actions.LoadFiles:
		s.files = a.Files
		s.app.Dispatch(&actions.ChangeText{
			Text: s.files[s.current],
		})
		payload.Notify()
	case *actions.FormatCode:
		b, err := format.Source([]byte(s.files[s.current]))
		if err != nil {
			s.app.Fail(err)
			return true
		}
		s.files[s.current] = string(b)
		payload.Notify()
		if a.Then != nil {
			s.app.Dispatch(a.Then)
		}
	}
	return true
}
