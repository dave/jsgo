package stores

import (
	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
)

func NewEditorStore(app *App) *EditorStore {
	s := &EditorStore{
		app: app,
	}
	return s
}

type EditorStore struct {
	app *App

	sizes []float64
	text  string
}

func (s *EditorStore) Sizes() []float64 {
	return s.sizes
}

func (s *EditorStore) Text() string {
	return s.text
}

func (s *EditorStore) Handle(payload *flux.Payload) bool {
	switch a := payload.Action.(type) {

	case *actions.ChangeSplit:
		s.sizes = a.Sizes
		payload.Notify()
	case *actions.ChangeText:
		s.text = a.Text
		payload.Notify()
	case *actions.UserChangedSplit:
		s.sizes = a.Sizes
	case *actions.UserChangedText:
		s.text = a.Text

	}
	return true
}
