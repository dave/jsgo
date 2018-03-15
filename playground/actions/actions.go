package actions

import (
	"github.com/dave/flux"
	"github.com/dave/jsgo/server/messages"
)

type Send struct {
	Message messages.Message
}

type Dial struct {
	Url     string
	Open    func() flux.ActionInterface
	Message func(interface{}) flux.ActionInterface
	Close   func() flux.ActionInterface
}

type CompileStart struct{}

type UpdateStart struct{}
type UpdateOpen struct{}
type UpdateMessage struct{ Message interface{} }
type UpdateClose struct{}

type Load struct{}

type ChangeSplit struct {
	Sizes []float64
}

type ChangeText struct {
	Text string
}

type UserChangedSplit struct {
	Sizes []float64
}

type UserChangedText struct {
	Text string
}
