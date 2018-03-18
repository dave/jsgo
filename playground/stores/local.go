package stores

import (
	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/locstor"
)

type LocalStore struct {
	app *App

	local *locstor.DataStore
}

func NewLocalStore(app *App) *LocalStore {
	s := &LocalStore{
		app:   app,
		local: locstor.NewDataStore(locstor.JSONEncoding),
	}
	return s
}

func (s *LocalStore) Handle(payload *flux.Payload) bool {
	switch action := payload.Action.(type) {
	case *actions.Load:
		var sizes []float64
		found, err := s.local.Find("split-sizes", &sizes)
		if err != nil {
			s.app.Fail(err)
			return true
		}
		if !found {
			sizes = defaultSizes
		}
		s.app.Dispatch(&actions.ChangeSplit{Sizes: sizes})

		var file string
		found, err = s.local.Find("current-file", &file)
		if err != nil {
			s.app.Fail(err)
			return true
		}
		if !found {
			file = defaultFile
		}
		s.app.Dispatch(&actions.ChangeFile{Name: file})

		var text string
		found, err = s.local.Find("editor-text", &text)
		if err != nil {
			s.app.Fail(err)
			return true
		}
		if !found {
			text = defaultText
		}
		s.app.Dispatch(&actions.ChangeText{Text: text})

	case *actions.UserChangedSplit:
		if err := s.local.Save("split-sizes", action.Sizes); err != nil {
			s.app.Fail(err)
			return true
		}
	case *actions.UserChangedText:
		if err := s.local.Save("editor-text", action.Text); err != nil {
			s.app.Fail(err)
			return true
		}
	case *actions.UserChangedFile:
		if err := s.local.Save("current-file", action.Name); err != nil {
			s.app.Fail(err)
			return true
		}
	}
	return true
}

var (
	defaultSizes = []float64{50, 50}
	defaultFile  = "main.go"
	defaultText  = `package main

import (
	"honnef.co/go/js/dom"
	"math/rand"
	"time"
	"fmt"
)

func main() {
    r := rand.Intn(10000)
	body := dom.GetWindow().Document().GetElementsByTagName("body")[0]
	body.SetInnerHTML("Hello, World! " + fmt.Sprint(r))
}

func init() {
    rand.Seed(time.Now().UTC().UnixNano())
}`
)
