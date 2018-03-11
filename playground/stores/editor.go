package stores

import (
	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
)

type EditorStore struct {
	app *App

	SplitSizes []float64
	EditorText string
}

func NewEditorStore(app *App) *EditorStore {
	s := &EditorStore{
		app: app,
	}
	return s
}

func (s *EditorStore) Handle(payload *flux.Payload) bool {
	switch a := payload.Action.(type) {

	case *actions.ChangeSplit:
		s.SplitSizes = a.Sizes
		payload.Notify()
	case *actions.ChangeText:
		s.EditorText = a.Text
		payload.Notify()
	case *actions.UserChangedSplit:
		s.SplitSizes = a.Sizes
	case *actions.UserChangedText:
		s.EditorText = a.Text

	}
	return true
}
