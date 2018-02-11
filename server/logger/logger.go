package logger

import (
	"encoding/json"

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

func (l *Logger) Log(typ LoggerType, payload interface{}) error {
	m := Message{Type: string(typ), Payload: payload}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := l.ws.Write(b); err != nil {
		return err
	}
	return nil
}

type LoggerType string

const Download LoggerType = "download"

const Compile LoggerType = "compile"

const Store LoggerType = "store"

const Index LoggerType = "index"

const Complete LoggerType = "complete"

const Error LoggerType = "error"

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
	Done bool `json:"done"`
}

type CompletePayload struct {
	Path    string `json:"path"`
	HashMin string `json:"hashmin"`
	HashMax string `json:"hashmax"`
}

type ErrorPayload struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}
