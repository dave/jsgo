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
	Archive{},
}

type Download struct {
	Starting bool
	Message  string
	Done     bool
}

type Archive struct {
	Path     string
	Hash     string
	Contents []byte
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

type PlaygroundCompile struct {
	// Source packages for this build: map[<package>]map[<filename>]<contents>
	Source map[string]map[string]string

	// Build tags
	Tags []string

	// Hashes of previously compiled dependencies to use if still in the cache
	Dependencies map[string]string
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

func CompileWriter(send chan Message) compileWriter {
	return compileWriter{send: send}
}

func DownloadWriter(send chan Message) downloadWriter {
	return downloadWriter{send: send}
}

type compileWriter struct {
	send chan Message
}

type downloadWriter struct {
	send chan Message
}

func (w downloadWriter) Write(b []byte) (n int, err error) {
	w.send <- Download{Message: strings.TrimSuffix(string(b), "\n")}
	return len(b), nil
}

func (w compileWriter) Write(b []byte) (n int, err error) {
	w.send <- Compile{Message: strings.TrimSuffix(string(b), "\n")}
	return len(b), nil
}
