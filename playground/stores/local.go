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

		var current string
		found, err = s.local.Find("current-file", &current)
		if err != nil {
			s.app.Fail(err)
			return true
		}
		if !found {
			current = defaultFile
		}
		s.app.Dispatch(&actions.ChangeFile{Name: current})

		var files map[string]string
		found, err = s.local.Find("files", &files)
		if err != nil {
			s.app.Fail(err)
			return true
		}
		if !found {
			files = defaultFiles
		}
		s.app.Dispatch(&actions.LoadFiles{Files: files})

	case *actions.UserChangedSplit:
		if err := s.local.Save("split-sizes", action.Sizes); err != nil {
			s.app.Fail(err)
			return true
		}
	case *actions.UserChangedText, *actions.AddFile:
		payload.Wait(s.app.Editor)
		if err := s.local.Save("files", s.app.Editor.Files()); err != nil {
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
	defaultFiles = map[string]string{
		"main.go": `package main

import (
	"honnef.co/go/js/dom"
)

func main() {
    body := dom.GetWindow().Document().GetElementsByTagName("body")[0]
	body.SetInnerHTML("Hello, World! " + randnum())
}`,
		"rand.go": `package main

import (
    "fmt"
    "time"
    "math/rand"
)

func randnum() string {
    r := rand.Intn(10000)
    return fmt.Sprint(r)
}

func init() {
    rand.Seed(time.Now().UTC().UnixNano())
}`}
)
