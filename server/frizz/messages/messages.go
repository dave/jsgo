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

	// Commands:
	gob.Register(GetSource{})

	// Data messages:
	gob.Register(SourceIndex{})
	gob.Register(Source{})

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

type GetSource struct {
	Path  string
	Cache map[string]string // Map of path->hash of previously downloaded dependencies to use if still in the cache
	Tags  []string
}

// SourceIndex is a list of the source of dependencies.
type SourceIndex map[string]SourceIndexItem

// SourceIndexItem is an item in SourceIndex. Unchanged is true if the client already has cached as
// specified by Cache in the Source message. Unchanged dependencies are not sent as Source messages.
type SourceIndexItem struct {
	Hash      string // Hash of the source pack
	Unchanged bool   // Unchanged is true if the package already exists in the client cache.
}

// Source contains information about the source pack
type Source struct {
	Path     string
	Hash     string // Hash of the source pack
	Standard bool
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
