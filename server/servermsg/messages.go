package servermsg

import "encoding/gob"

func RegisterTypes() {
	gob.Register(Queueing{})
	gob.Register(Error{})
}

type Queueing struct {
	Position int
	Done     bool
}

type Error struct {
	Message string
}
