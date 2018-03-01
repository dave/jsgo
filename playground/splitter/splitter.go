package splitter

import (
	"github.com/go-humble/locstor"
	"github.com/gopherjs/gopherjs/js"
)

func New(name string) *Split {
	return &Split{name: name}
}

func (s *Split) Init(args ...interface{}) {
	s.Object = js.Global.Call("Split", args...)
}

type Split struct {
	*js.Object
	name string
}

func (s *Split) SaveSizes() {
	value := stringifyJSON(s.Call("getSizes"))
	if err := locstor.SetItem(s.name, value); err != nil {
		panic(err)
	}
}

func (s *Split) GetSavedSized(init interface{}) interface{} {
	value, err := locstor.GetItem(s.name)
	if err != nil {
		if _, isNotFound := err.(locstor.ItemNotFoundError); isNotFound {
			return init
		}
		panic(err)
	}
	return parseJSON(value).Interface()
}

func parseJSON(json string) *js.Object {
	return js.Global.Get("JSON").Call("parse", json)
}

func stringifyJSON(o *js.Object) string {
	return js.Global.Get("JSON").Call("stringify", o).String()
}
