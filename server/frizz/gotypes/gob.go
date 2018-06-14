package gotypes

import (
	"encoding/gob"
	"sync"
)

var gobOnce = &sync.Once{}

func RegisterTypesGob() {
	gobOnce.Do(func() {
		gob.Register(Reference{})
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

		gob.Register(PkgName{})
		gob.Register(Const{})
		gob.Register(TypeName{})
		gob.Register(Var{})
		gob.Register(Func{})
		gob.Register(Label{})
		gob.Register(Builtin{})
		gob.Register(Nil{})
	})
}
