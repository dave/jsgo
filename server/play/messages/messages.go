package messages

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/dave/jsgo/server/servermsg"
	"github.com/dave/services"
	"github.com/dave/services/builder/buildermsg"
	"github.com/dave/services/deployer/deployermsg"
	"github.com/dave/services/fileserver/constor/constormsg"
	"github.com/dave/services/getter/gettermsg"
	"github.com/gorilla/websocket"
)

var payloads = []interface{}{

	// Progress messages:
	servermsg.Queueing{},
	gettermsg.Downloading{},

	constormsg.Storing{},
	buildermsg.Building{},

	// Data messages:
	servermsg.Error{},
	ShareComplete{},
	GetComplete{},
	DeployComplete{},

	deployermsg.Archive{},
	deployermsg.Index{},

	// Commands:
	Update{},
	Share{},
	Get{},
	Deploy{},
	Initialise{},
}

type DeployComplete struct {
	Main  string
	Index string
}

// Update is sent by the client to the server asking it to compile the source and return the archive
// files for all dependencies that are not found in the client cache.
type Update struct {
	Source map[string]map[string]string // Source packages for this build: map[<package>]map[<filename>]<contents>
	Tags   []string                     // Build tags
	Cache  map[string]string            // Map of path->hash of previously compiled dependencies to use if still in the cache
	Minify bool
}

// Share is sent by the client to persist the setup on the server. This will be persisted publicly as
// json, so best to use json tags to lower-case the names.
type Share struct {
	Version int                          `json:"version"`
	Source  map[string]map[string]string `json:"source"` // Source packages for this build: map[<package>]map[<filename>]<contents>
	Tags    []string                     `json:"tags"`   // Build tags
}

type Deploy struct {
	Main    string
	Imports []string
	Source  map[string]map[string]string // Source packages for this build: map[<package>]map[<filename>]<contents>
	Tags    []string
}

// Initialise is sent by the client to get the source at Path, and update.
type Initialise struct {
	Path   string
	Minify bool
}

// Get is sent by the client to the server asking it to download a package and return the source.
type Get struct {
	Path string
}

type GetComplete struct {
	Source map[string]map[string]string
}

type ShareComplete struct {
	Hash string
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
		Message json.RawMessage
	}
	if err := json.Unmarshal(in, &m); err != nil {
		return nil, err
	}
	typ, ok := payloadTypes[m.Type]
	if !ok {
		return nil, fmt.Errorf("type not found: %s, %#v", m.Type, m)
	}
	pointer := reflect.New(typ)
	if err := json.Unmarshal(m.Message, pointer.Interface()); err != nil {
		return nil, err
	}
	return pointer.Elem().Interface(), nil
}

func init() {
	for _, i := range payloads {
		t := reflect.TypeOf(i)
		payloadTypes[t.Name()] = t
	}
}

var payloadTypes = make(map[string]reflect.Type)
