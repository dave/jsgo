package gotypes

import "encoding/gob"

func RegisterTypes() {
	gob.Register(Circular{})
	gob.Register(Basic{})
	gob.Register(Array{})
	gob.Register(Slice{})
	gob.Register(Struct{})
	gob.Register(Pointer{})
	gob.Register(Tuple{})
	gob.Register(Signature{})
	gob.Register(Interface{})
	gob.Register(Map{})
	gob.Register(Chan{})
	gob.Register(Named{})
}
