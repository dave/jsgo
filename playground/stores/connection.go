package stores

import (
	"errors"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/server/messages"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket/websocketjs"
)

type ConnectionStore struct {
	app *App

	open bool
	ws   *websocketjs.WebSocket
}

func NewConnectionStore(app *App) *ConnectionStore {
	s := &ConnectionStore{
		app: app,
	}
	return s
}

func (s *ConnectionStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.Send:
		if !s.open {
			s.app.Fail(errors.New("connection closed"))
			return true
		}
		b, err := messages.Marshal(action.Message)
		if err != nil {
			s.app.Fail(err)
			return true
		}
		if err := s.ws.Send(string(b)); err != nil {
			s.app.Fail(err)
			return true
		}
	case *actions.Dial:
		if s.open {
			s.app.Fail(errors.New("connection already open"))
			return true
		}
		var err error
		if s.ws, err = websocketjs.New(action.Url); err != nil {
			s.app.Fail(err)
			return true
		}
		s.open = true
		s.ws.AddEventListener("open", false, func(ev *js.Object) {
			s.app.Dispatch(action.Open())
		})
		s.ws.AddEventListener("message", false, func(ev *js.Object) {
			m, err := messages.Unmarshal([]byte(ev.Get("data").String()))
			if err != nil {
				s.app.Fail(err)
				return
			}
			s.app.Dispatch(action.Message(m))
		})
		s.ws.AddEventListener("close", false, func(ev *js.Object) {
			s.app.Dispatch(action.Close())
			s.ws.Close()
			s.open = false
		})
		s.ws.AddEventListener("error", false, func(ev *js.Object) {
			s.app.Fail(errors.New("error from server"))
			s.ws.Close()
			s.open = false
		})
	}
	return true
}
