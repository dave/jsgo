package messages

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type Message interface{}

var payloads = []interface{}{

	// Progress messages:
	Queueing{},
	Downloading{},
	Compiling{},
	Storing{},
	Updating{},

	// Data messages:
	Complete{},
	Error{},
	Archive{},
	Index{},

	// Commands:
	Update{},
}

type Queueing struct {
	Position int
	Done     bool
}

type Downloading struct {
	Starting bool
	Message  string
	Done     bool
}

type Compiling struct {
	Starting bool
	Message  string
	Done     bool
}

type Updating struct {
	Starting bool
	Message  string
	Done     bool
}

type Storing struct {
	Starting  bool
	Finished  int
	Unchanged int
	Remain    int
	Done      bool
}

type Complete struct {
	Path    string
	Short   string
	HashMin string
	HashMax string
}

type Error struct {
	Path    string
	Message string
}

// Update is sent by the client to the server asking it to compile the source and return the archive
// files for all dependencies that are not found in the client cache.
type Update struct {
	Source map[string]map[string]string // Source packages for this build: map[<package>]map[<filename>]<contents>
	Tags   []string                     // Build tags
	Cache  map[string]string            // Map of path->hash of previously compiled dependencies to use if still in the cache
}

// Index is an ordered list of dependencies.
type Index map[string]IndexItem

// IndexItem is an item in Index. Unchanged is true if the client already has cached as specified by
// Cache in the Update message. Unchanged dependencies are not sent as Archive messages.
type IndexItem struct {
	Hash      string // Hash of the js file
	Unchanged bool   // Unchanged is true if the package already exists in the client cache.
}

// Archive contains the contents (gzipped) of the GopherJS archive file.
type Archive struct {
	Path     string
	Hash     string // Hash of the resultant js
	Contents []byte // Contents of the file (gzipped)
}

func Marshal(in Message) ([]byte, error) {
	m := struct {
		Type    string
		Message Message
	}{
		Type:    reflect.TypeOf(in).Name(),
		Message: in,
	}
	return json.Marshal(m)
}

func Unmarshal(in []byte) (Message, error) {
	var m struct {
		Type    string
		Message json.RawMessage
	}
	if err := json.Unmarshal(in, &m); err != nil {
		return nil, err
	}
	typ, ok := payloadTypes[m.Type]
	if !ok {
		return nil, fmt.Errorf("type not found: %s", m.Type)
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
