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

// CompileStart compiles the app and injects the js into the iframe
type CompileStart struct{}

// UpdateStart updates the deps from the server and if Run == true, compiles and runs the app
type UpdateStart struct {
	Run bool
}
type UpdateOpen struct{}
type UpdateMessage struct{ Message interface{} }
type UpdateClose struct {
	Run bool
}

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

type ImportsChanged struct{}
