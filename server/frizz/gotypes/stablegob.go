package gotypes

import (
	"sync"

	"github.com/dave/stablegob"
)

var stableGobOnce = &sync.Once{}

func RegisterTypesStablegob() {
	stableGobOnce.Do(func() {
		stablegob.Register(&Reference{})
		stablegob.Register(&Basic{})
		stablegob.Register(&Array{})
		stablegob.Register(&Slice{})
		stablegob.Register(&Struct{})
		stablegob.Register(&Pointer{})
		stablegob.Register(&Tuple{})
		stablegob.Register(&Signature{})
		stablegob.Register(&Interface{})
		stablegob.Register(&Map{})
		stablegob.Register(&Chan{})
		stablegob.Register(&Named{})

		stablegob.Register(&PkgName{})
		stablegob.Register(&Const{})
		stablegob.Register(&TypeName{})
		stablegob.Register(&Var{})
		stablegob.Register(&Func{})
		stablegob.Register(&Label{})
		stablegob.Register(&Builtin{})
		stablegob.Register(&Nil{})
	})
}
