package messages

import (
	"encoding/json"
	"reflect"

	"github.com/dave/services"
	"github.com/gorilla/websocket"
)

type Compile struct {
	Path string
}

type Complete struct {
	Path    string
	Short   string
	HashMin string
	HashMax string
}

func Marshal(in services.Message) ([]byte, int, error) {
	m := struct {
		Type    string
		Message services.Message
	}{
		Type:    reflect.TypeOf(in).Name(),
		Message: in,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, 0, err
	}
	return b, websocket.TextMessage, nil
}

func Unmarshal(in []byte) (services.Message, error) {
	var m struct {
		Type    string
		Message Compile // the jsgo compile page only ever sends Compile messages
	}
	if err := json.Unmarshal(in, &m); err != nil {
		return nil, err
	}
	return m.Message, nil
}
