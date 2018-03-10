package messages

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

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

type Message interface{}

type Download struct {
	Starting bool   `json:"starting"`
	Message  string `json:"message,omitempty"`
	Done     bool   `json:"done"`
}

type Compile struct {
	Starting bool   `json:"starting"`
	Message  string `json:"message,omitempty"`
	Done     bool   `json:"done"`
}

type Store struct {
	Starting  bool `json:"starting"`
	Finished  int  `json:"finished"`
	Unchanged int  `json:"unchanged"`
	Remain    int  `json:"remain"`
	Done      bool `json:"done"`
}

type Complete struct {
	Path    string `json:"path"`
	Short   string `json:"short"`
	HashMin string `json:"hashmin"`
	HashMax string `json:"hashmax"`
}

type Error struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type Queue struct {
	Position int  `json:"position"`
	Done     bool `json:"done"`
}

type PlaygroundCompile struct {
	// Source packages for this build: map[<package>]map[<filename>]<contents>
	Source map[string]map[string]string `json:"source"`

	// Build tags
	Tags []string `json:"tags"`

	// Hashes of previously compiled dependencies to use if still in the cache
	Dependencies map[string]string `json:"dependencies"`
}

var payloads = []interface{}{
	Download{},
	Compile{},
	Store{},
	Complete{},
	Error{},
	Queue{},
	PlaygroundCompile{},
}

func Marshal(in Message) ([]byte, error) {
	m := struct {
		Type    string  `json:"type"`
		Message Message `json:"message"`
	}{
		Type:    reflect.TypeOf(in).Name(),
		Message: in,
	}
	return json.Marshal(m)
}

func Unmarshal(in []byte) (Message, error) {
	var m struct {
		Type    string          `json:"type"`
		Message json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal(in, &m); err != nil {
		return nil, err
	}
	typ, ok := payloadTypes[m.Type]
	if !ok {
		return nil, fmt.Errorf("type not found: %s", m.Type)
	}
	payload := reflect.New(typ).Elem().Interface()
	if err := json.Unmarshal(m.Message, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func init() {
	for _, i := range payloads {
		t := reflect.TypeOf(i)
		payloadTypes[t.Name()] = t
	}
}

var payloadTypes = make(map[string]reflect.Type)
