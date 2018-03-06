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

func (s *Split) SetSizes(sizes []float64) {
	out := make([]interface{}, len(sizes))
	for i, v := range sizes {
		out[i] = v
	}
	s.Object.Call("setSizes", out)
}

func (s *Split) SetSizesIfChanged(sizes []float64) {
	current := s.GetSizes()
	if len(current) != len(sizes) {
		s.SetSizes(sizes)
		return
	}
	for i, v := range sizes {
		if current[i] != v {
			s.SetSizes(sizes)
			return
		}
	}
}

func (s *Split) GetSizes() []float64 {
	raw := s.Call("getSizes").Interface().([]interface{})
	var out []float64
	for _, v := range raw {
		if f, ok := v.(float64); ok {
			out = append(out, f)
		}
	}
	return out
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
