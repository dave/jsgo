package stores

import (
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
	adding  bool
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

func (s *EditorStore) Adding() bool {
	return s.adding
}

func (s *EditorStore) Files() map[string]string {
	f := map[string]string{}
	for k, v := range s.files {
		f[k] = v
	}
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
	case *actions.AddFileClick:
		js.Global.Call("$", "#add-file").Call("modal", "show")
		s.adding = true
		payload.Notify()
	case *actions.UsedClosedAddFileModal:
		s.adding = false
		payload.Notify()
	case *actions.CloseAddFileModal:
		js.Global.Call("$", "#add-file").Call("modal", "hide")
		s.adding = false
		payload.Notify()
	}
	return true
}
