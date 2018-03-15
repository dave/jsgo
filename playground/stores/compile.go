package stores

import (
	"fmt"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
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
	switch action := payload.Action.(type) {
	case *actions.CompileStart:
		fmt.Println(action)
	}
	return true
}
