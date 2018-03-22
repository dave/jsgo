package stores

import (
	"errors"

	"honnef.co/go/js/dom"

	"fmt"

	"strings"

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

func (s *ConnectionStore) Open() bool {
	return s.open
}

func (s *ConnectionStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.Send:
		s.app.Debug(fmt.Sprintf("Sending %T", action.Message), action.Message)
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
		s.app.Debug("Web socket dialing", action.Url)
		var err error
		if s.ws, err = websocketjs.New(action.Url); err != nil {
			s.app.Fail(err)
			return true
		}
		s.open = true
		s.ws.AddEventListener("open", false, func(ev *js.Object) {
			go func() {
				s.app.Debug("Web socket open")
				s.app.Dispatch(action.Open())
			}()
		})
		s.ws.AddEventListener("message", false, func(ev *js.Object) {
			go func() {
				m, err := messages.Unmarshal([]byte(ev.Get("data").String()))
				if err != nil {
					s.app.Fail(err)
					return
				}
				s.app.Debug(fmt.Sprintf("Received %T", m), m)
				if e, ok := m.(messages.Error); ok {
					if e.Path != "" {
						s.app.Fail(fmt.Errorf("%s: %s", e.Path, e.Message))
					} else {
						s.app.Fail(errors.New(e.Message))
					}
					return
				}
				s.app.Dispatch(action.Message(m))
			}()
		})
		s.ws.AddEventListener("close", false, func(ev *js.Object) {
			go func() {
				s.app.Debug("Web socket closed")
				s.app.Dispatch(action.Close())
				s.ws.Close()
				s.open = false
			}()
		})
		s.ws.AddEventListener("error", false, func(ev *js.Object) {
			go func() {
				s.app.Debug("Web socket error")
				s.app.Fail(errors.New("error from server"))
				s.ws.Close()
				s.open = false
			}()
		})
	}
	return true
}

func defaultUrl() string {
	var url string
	if strings.HasPrefix(dom.GetWindow().Document().DocumentURI(), "https://") {
		url = "wss://compile.jsgo.io/_pg/"
	} else {
		url = "ws://localhost:8081/_pg/"
	}
	return url
}
