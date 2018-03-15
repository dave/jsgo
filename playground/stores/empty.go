package stores

import (
	"fmt"

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
		fmt.Println(action)
	}
	return true
}
