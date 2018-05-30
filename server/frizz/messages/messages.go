package messages

import (
	"bytes"
	"encoding/gob"

	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/dave/jsgo/server/servermsg"
	"github.com/dave/services"
	"github.com/dave/services/builder/buildermsg"
	"github.com/dave/services/constor/constormsg"
	"github.com/dave/services/deployer/deployermsg"
	"github.com/dave/services/getter/gettermsg"
	"github.com/gorilla/websocket"
)

type Payload struct {
	Message services.Message
}

func init() {
	// Data messages:
	gob.Register(TypesComplete{})

	// Commands:
	gob.Register(Types{})

	// Initialise types in gotypes
	gotypes.RegisterTypes()

	// Initialise types in deployermsg
	deployermsg.RegisterTypes()

	// Initialise types in servermsg
	servermsg.RegisterTypes()

	// Initialise types in buildermsg
	buildermsg.RegisterTypes()

	// Initialise types in constormsg
	constormsg.RegisterTypes()

	// Initialise types in gettermsg
	gettermsg.RegisterTypes()
}

type Types struct {
	Path string
}
type TypesComplete struct {
	Types []gotypes.Named
}

func Marshal(in services.Message) ([]byte, int, error) {
	p := Payload{in}
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(p); err != nil {
		return nil, 0, err
	}
	return buf.Bytes(), websocket.BinaryMessage, nil
}

func Unmarshal(in []byte) (services.Message, error) {
	var p Payload
	if err := gob.NewDecoder(bytes.NewBuffer(in)).Decode(&p); err != nil {
		return nil, err
	}
	return p.Message, nil
}
