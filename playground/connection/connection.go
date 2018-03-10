package connection

import (
	"errors"

	"github.com/dave/jsgo/server/messages"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket/websocketjs"
)

type Conn struct {
	open    bool
	ws      *websocketjs.WebSocket
	Receive chan interface{}
}

func New() *Conn {
	c := &Conn{}
	c.Receive = make(chan interface{})
	return c
}

func (c *Conn) Send(message messages.Message) error {
	if !c.open {
		return errors.New("connection closed")
	}
	b, err := messages.Marshal(message)
	if err != nil {
		return err
	}
	if err := c.ws.Send(string(b)); err != nil {
		return err
	}
	return nil
}

func (c *Conn) Dial(url string, openAction func() interface{}, messageAction func(interface{}) interface{}, closeAction func() interface{}, errorAction func(error) interface{}) error {
	if c.open {
		return errors.New("connection already open")
	}
	var err error
	if c.ws, err = websocketjs.New(url); err != nil {
		return err
	}
	c.open = true
	c.ws.AddEventListener("open", false, func(ev *js.Object) {
		select {
		case c.Receive <- openAction():
		default:
		}
	})
	c.ws.AddEventListener("message", false, func(ev *js.Object) {
		m, err := messages.Unmarshal([]byte(ev.Get("data").String()))
		if err != nil {
			panic(err)
		}
		select {
		case c.Receive <- messageAction(m):
		default:
		}
	})
	c.ws.AddEventListener("close", false, func(ev *js.Object) {
		select {
		case c.Receive <- closeAction():
		default:
		}
		c.ws.Close()
		c.open = false
	})
	c.ws.AddEventListener("error", false, func(ev *js.Object) {
		select {
		case c.Receive <- errorAction(errors.New("error from websocket")):
		default:
		}
		c.ws.Close()
		c.open = false
	})
	return nil
}
