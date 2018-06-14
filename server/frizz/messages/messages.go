package messages

import (
	"bytes"
	"encoding/gob"

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
	gob.Register(GetPackages{})

	// Data messages:
	gob.Register(PackageIndex{})
	gob.Register(Source{})
	gob.Register(Objects{})

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

type GetPackages struct {
	Path    string
	Tags    []string
	Source  map[string]string // Map of path->hash of previously downloaded source dependencies to use if still in the cache
	Objects map[string]string // Map of path->hash of previously downloaded objects dependencies to use if still in the cache
}

// PackagesIndex is returned to the client to summarize the Source and Objects messages that were also
// sent, and to confirm which were not sent because the cache was up-to-date.
type PackageIndex struct {
	Path    string
	Tags    []string
	Source  map[string]IndexItem
	Objects map[string]IndexItem
}

// IndexItem is an item in PackagesIndex. Unchanged is true if the client already has cached as
// specified by Cache in the GetPackages message. Unchanged dependencies are not sent as Source / Objects
// messages.
type IndexItem struct {
	Hash      string // Hash of the item
	Unchanged bool   // Unchanged is true if the package already exists in the client cache.
}

// Source contains information about the source pack
type Source struct {
	Path     string
	Hash     string // Hash of the source pack
	Standard bool
}

// Objects contains information about the object pack
type Objects struct {
	Path     string
	Hash     string // Hash of the object pack
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
