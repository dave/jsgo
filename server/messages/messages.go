package messages

import "strings"

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
	w.send <- Message{Type: Download, Payload: Payload{Message: strings.TrimSuffix(string(b), "\n")}}
	return len(b), nil
}

func (w compileWriter) Write(b []byte) (n int, err error) {
	w.send <- Message{Type: Compile, Payload: Payload{Message: strings.TrimSuffix(string(b), "\n")}}
	return len(b), nil
}

type Type string

const Download Type = "download"

const Compile Type = "compile"

const Store Type = "store"

const Complete Type = "complete"

const Error Type = "error"

const Queue Type = "queue"

type Message struct {
	Type    Type        `json:"type"`
	Payload interface{} `json:"payload"`
}

type Payload struct {
	Message string `json:"message,omitempty"`
	Done    bool   `json:"done"`
}

type StorePayload struct {
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
