package logger

import (
	"golang.org/x/net/websocket"
)

func New(conn *websocket.Conn) *Logger {
	return &Logger{
		ws: conn,
	}
}

type Logger struct {
	ws *websocket.Conn
}

func (l *Logger) Log(typ LoggerType, payload interface{}) {
	m := Message{Type: string(typ), Payload: payload}
	if err := websocket.JSON.Send(l.ws, m); err != nil {
		// This should never happen
		l.ws.Write([]byte(`{"type":"error","payload":{"path":"error","message":"error marshaling payload"}}`))
		return
	}
	return
}

func (l *Logger) CompileWriter() compileWriter {
	return compileWriter{l: l}
}

func (l *Logger) DownloadWriter() downloadWriter {
	return downloadWriter{l: l}
}

type downloadWriter struct {
	l *Logger
}

func (w downloadWriter) Write(b []byte) (n int, err error) {
	w.l.Log(Download, DownloadingPayload{Path: string(b)})
	return len(b), nil
}

type compileWriter struct {
	l *Logger
}

func (w compileWriter) Write(b []byte) (n int, err error) {
	w.l.Log(Compile, CompilingPayload{Path: string(b)})
	return len(b), nil
}

type LoggerType string

const Download LoggerType = "download"

const Compile LoggerType = "compile"

const Store LoggerType = "store"

const Index LoggerType = "index"

const Complete LoggerType = "complete"

const Error LoggerType = "error"

const Queue LoggerType = "queue"

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type DownloadingPayload struct {
	Path string `json:"path,omitempty"`
	Done bool   `json:"done"`
}

type CompilingPayload struct {
	Path string `json:"path,omitempty"`
	Done bool   `json:"done"`
}

type StoringPayload struct {
	Path string `json:"path,omitempty"`
	Done bool   `json:"done"`
}

type IndexPayload struct {
	Path string `json:"path,omitempty"`
	Done bool   `json:"done"`
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
