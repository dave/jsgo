package stores

import (
	"archive/zip"
	"bytes"

	"io"

	"fmt"

	"github.com/dave/flux"
	"github.com/dave/jsgo/playground/actions"
	"github.com/dave/jsgo/server/messages"
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
	case *actions.CompileMessage:
		switch message := action.Message.(type) {
		case messages.PlaygroundIndex:
			for _, v := range message {
				fmt.Println(v.Path, v.Hash, v.Unchanged)
			}
		case messages.PlaygroundArchive:
			br := bytes.NewReader(message.Contents)
			zr, err := zip.NewReader(br, int64(br.Len()))
			if err != nil {
				s.app.Fail(err)
				return true
			}
			fr, err := zr.File[0].Open()
			if err != nil {
				s.app.Fail(err)
				return true
			}
			defer fr.Close()
			out := &bytes.Buffer{}
			if _, err := io.Copy(out, fr); err != nil {
				s.app.Fail(err)
				return true
			}
			if err := s.local.Save(message.Path, out.Bytes()); err != nil {
				s.app.Fail(err)
				return true
			}
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
