package splitter

import (
	"github.com/gopherjs/gopherjs/js"
)

func New(name string) *Split {
	return &Split{name: name}
}

func (s *Split) Init(args ...interface{}) {
	s.Object = js.Global.Call("Split", args...)
}

func (s *Split) GetSizes() []float64 {
	raw := s.Call("getSizes").Interface().([]interface{})
	out := make([]float64, len(raw))
	for i, v := range raw {
		out[i] = v.(float64)
	}
	return out[:]
}

type Split struct {
	*js.Object
	name string
}

/*
func (s *Split) SaveSizes() {
	value := stringifyJSON(s.Call("getSizes"))
	if err := locstor.SetItem(s.name, value); err != nil {
		panic(err)
	}
}

func (s *Split) GetSavedSized(init interface{}) (interface{}, error) {
	value, found, err := locstor.GetItem(s.name)
	if err != nil {
		panic(err)
	}
	return parseJSON(value).Interface()
}
*/

func parseJSON(json string) *js.Object {
	return js.Global.Get("JSON").Call("parse", json)
}

func stringifyJSON(o *js.Object) string {
	return js.Global.Get("JSON").Call("stringify", o).String()
}
