package messages

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type Message interface{}

var payloads = []interface{}{
	Download{},
	Compile{},
	Store{},
	Complete{},
	Error{},
	Queue{},
	PlaygroundCompile{},
	PlaygroundArchive{},
	PlaygroundIndex{},
}

type Download struct {
	Starting bool
	Message  string
	Done     bool
}

type Compile struct {
	Starting bool
	Message  string
	Done     bool
}

type Store struct {
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

type Queue struct {
	Position int
	Done     bool
}

// PlaygroundCompile is sent by the client to the server asking it to compile the source and return the
// archive files for all dependencies that are not found in the client cache.
type PlaygroundCompile struct {
	Source       map[string]map[string]string // Source packages for this build: map[<package>]map[<filename>]<contents>
	Tags         []string                     // Build tags
	ArchiveCache map[string]string            // Map of path->hash of previously compiled dependencies to use if still in the cache
}

// PlaygroundIndex is an ordered list of dependencies.
type PlaygroundIndex []PlaygroundIndexItem

// PlaygroundIndexItem is an item in PlaygroundIndex. Unchanged is true for any that the client already
// has cached as specified by ArchiveCache in the PlaygroundCompile message. Unchanged dependencies are
// not sent as PlaygroundArchive messages.
type PlaygroundIndexItem struct {
	Path      string
	Hash      string // Hash of the raw file (unzipped)
	Unchanged bool   // Unchanged is true if the package already exists in the client cache.
}

// PlaygroundArchive contains the contents (zipped) of the GopherJS archive file.
type PlaygroundArchive struct {
	Path     string
	Hash     string // Hash of the raw file (unzipped)
	Contents []byte // Contents of the file (zipped)
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

func CompileWriter(send func(Message)) compileWriter {
	return compileWriter{send: send}
}

func DownloadWriter(send func(Message)) downloadWriter {
	return downloadWriter{send: send}
}

type compileWriter struct {
	send func(Message)
}

type downloadWriter struct {
	send func(Message)
}

func (w downloadWriter) Write(b []byte) (n int, err error) {
	w.send(Download{Message: strings.TrimSuffix(string(b), "\n")})
	return len(b), nil
}

func (w compileWriter) Write(b []byte) (n int, err error) {
	w.send(Compile{Message: strings.TrimSuffix(string(b), "\n")})
	return len(b), nil
}
