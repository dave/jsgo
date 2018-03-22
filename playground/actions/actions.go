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
type UpdateStart struct{ Run bool }
type UpdateOpen struct{}
type UpdateMessage struct{ Message interface{} }
type UpdateClose struct{ Run bool }

// UpdateStart updates the deps from the server and if Run == true, compiles and runs the app
type GetStart struct{ Path string }
type GetOpen struct{ Path string }
type GetMessage struct {
	Path    string
	Message interface{}
}
type GetClose struct{}

type Load struct{}

type ChangeSplit struct {
	Sizes []float64
}

type ChangeText struct {
	Text string
}

type LoadFiles struct {
	Files map[string]string
}

type ChangeFile struct {
	Name string
}

type UserChangedSplit struct {
	Sizes []float64
}

type UserChangedText struct {
	Text string
}

type UserChangedFile struct {
	Name string
}

type AddFile struct {
	Name string
}
type DeleteFile struct {
	Name string
}
type AddFileClick struct{}
type DeleteFileClick struct{}

type ImportsChanged struct{}

type FormatCode struct {
	Then flux.ActionInterface
}

type ShareStart struct{}
type ShareOpen struct{}
type ShareMessage struct{ Message interface{} }
type ShareClose struct{}
