package server

import (
	"testing"

	"golang.org/x/net/websocket"
)

func TestSocket(t *testing.T) {
	websocket.Dial("ws://localhost:8081/_pg/", "", "http://localhost:8080")
}
