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

/*
func (s *LocalStore) GetArchive(path string) (*compiler.Archive, bool, error) {
	var b []byte
	found, err := s.local.Find(path, &b)
	if err != nil {
		return nil, false, fmt.Errorf("%s error %s", path, err)
	}
	if !found {
		return nil, false, nil
	}
	var a compiler.Archive
	buf := bytes.NewBuffer(b)
	if err := gob.NewDecoder(buf).Decode(&a); err != nil {
		return nil, false, fmt.Errorf("%s gob error %s", path, err)
	}
	return &a, true, nil
}
*/

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
	}
	return true
}

var (
	defaultSizes = []float64{75, 25}
	defaultText  = `package main

import (
	"honnef.co/go/js/dom"
)

func main() {
	body := dom.GetWindow().Document().GetElementsByTagName("body")[0]
	body.SetInnerHTML("Hello, World!")
}`
)
