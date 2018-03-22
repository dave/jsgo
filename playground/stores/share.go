package stores

import (
	"fmt"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/server/messages"
	"honnef.co/go/js/dom"
)

func NewShareStore(app *App) *ShareStore {
	s := &ShareStore{
		app: app,
	}
	return s
}

type ShareStore struct {
	app *App
}

func (s *ShareStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.ShareStart:
		s.app.Log("sharing")
		s.app.Dispatch(&actions.Dial{
			Url:     defaultUrl(),
			Open:    func() flux.ActionInterface { return &actions.ShareOpen{} },
			Message: func(m interface{}) flux.ActionInterface { return &actions.ShareMessage{Message: m} },
			Close:   func() flux.ActionInterface { return &actions.ShareClose{} },
		})
		payload.Notify()
	case *actions.ShareOpen:
		message := messages.Share{
			Source: map[string]map[string]string{
				"main": s.app.Editor.Files(),
			},
		}
		s.app.Dispatch(&actions.Send{
			Message: message,
		})
	case *actions.ShareMessage:
		switch message := action.Message.(type) {
		case messages.Storing:
			s.app.Log("storing")
		case messages.ShareComplete:
			dom.GetWindow().Document().Underlying().Set("location", fmt.Sprintf("#%s", message.Hash))
			s.app.Log("shared")
		}
	case *actions.ShareClose:
		s.app.Log()
	}
	return true
}
