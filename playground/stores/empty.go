package stores

import (
	"github.com/dave/flux"
)

func NewEmptyStore(app *App) *EmptyStore {
	s := &EmptyStore{
		app: app,
	}
	return s
}

type EmptyStore struct {
	app *App
}

func (s *EmptyStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	default:
		_ = action
	}
	return true
}
