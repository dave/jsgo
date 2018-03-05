package messages

import (
	"encoding/json"
	"fmt"
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
	w.send <- Message{Type: Download, Payload: DownloadPayload{Message: strings.TrimSuffix(string(b), "\n")}}
	return len(b), nil
}

func (w compileWriter) Write(b []byte) (n int, err error) {
	w.send <- Message{Type: Compile, Payload: CompilePayload{Message: strings.TrimSuffix(string(b), "\n")}}
	return len(b), nil
}

type Type string

const Download Type = "download"

const Compile Type = "compile"

const Store Type = "store"

const Complete Type = "complete"

const Error Type = "error"

const Queue Type = "queue"

const PlaygroundCompile Type = "playground-compile"

type Message struct {
	Type    Type        `json:"type"`
	Payload interface{} `json:"payload"`
}

type DownloadPayload struct {
	Starting bool   `json:"starting"`
	Message  string `json:"message,omitempty"`
	Done     bool   `json:"done"`
}

type CompilePayload struct {
	Starting bool   `json:"starting"`
	Message  string `json:"message,omitempty"`
	Done     bool   `json:"done"`
}

type StorePayload struct {
	Starting  bool `json:"starting"`
	Finished  int  `json:"finished"`
	Unchanged int  `json:"unchanged"`
	Remain    int  `json:"remain"`
	Done      bool `json:"done"`
}

type CompletePayload struct {
	Path    string `json:"path"`
	Short   string `json:"short"`
	HashMin string `json:"hashmin"`
	HashMax string `json:"hashmax"`
}

type ErrorPayload struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type QueuePayload struct {
	Position int  `json:"position"`
	Done     bool `json:"done"`
}

type PlaygroundCompilePayload struct {
	// Source packages for this build: map[<package>]map[<filename>]<contents>
	Source map[string]map[string]string `json:"source"`

	// Build tags
	Tags []string `json:"tags"`

	// Hashes of previously compiled dependencies to use if still in the cache
	Dependencies map[string]string `json:"dependencies"`
}

func Parse(in []byte) (Message, error) {
	var m struct {
		Type    Type            `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(in, &m); err != nil {
		return Message{}, err
	}
	switch m.Type {
	case Compile:
		var payload CompilePayload
		if err := json.Unmarshal(m.Payload, &payload); err != nil {
			return Message{}, err
		}
		return Message{Type: m.Type, Payload: payload}, nil
	case Download:
		var payload DownloadPayload
		if err := json.Unmarshal(m.Payload, &payload); err != nil {
			return Message{}, err
		}
		return Message{Type: m.Type, Payload: payload}, nil
	case Store:
		var payload StorePayload
		if err := json.Unmarshal(m.Payload, &payload); err != nil {
			return Message{}, err
		}
		return Message{Type: m.Type, Payload: payload}, nil
	case Complete:
		var payload CompletePayload
		if err := json.Unmarshal(m.Payload, &payload); err != nil {
			return Message{}, err
		}
		return Message{Type: m.Type, Payload: payload}, nil
	case Error:
		var payload ErrorPayload
		if err := json.Unmarshal(m.Payload, &payload); err != nil {
			return Message{}, err
		}
		return Message{Type: m.Type, Payload: payload}, nil
	case Queue:
		var payload QueuePayload
		if err := json.Unmarshal(m.Payload, &payload); err != nil {
			return Message{}, err
		}
		return Message{Type: m.Type, Payload: payload}, nil
	}
	return Message{}, fmt.Errorf("invalid type %s", m.Type)
}
