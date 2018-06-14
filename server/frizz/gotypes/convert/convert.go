package convert

import (
	"go/types"

	"fmt"

	"github.com/dave/jsgo/server/frizz/gotypes"
)

func Object(o types.Object) gotypes.Object {
	if o == nil {
		return nil
	}
	obj := func(object types.Object, definition bool) gotypes.Obj {
		var objpath string
		if o.Pkg() != nil {
			objpath = o.Pkg().Path()
		}
		return gotypes.Obj{
			Identifier: gotypes.Identifier{
				Path: objpath,
				Name: o.Name(),
			},
			Type: Type(o.Type(), definition),
		}
	}
	switch o := o.(type) {
	case *types.PkgName:
		if o == nil {
			return gotypes.PkgName{}
		}
		return gotypes.PkgName{
			Obj:      obj(o, false),
			Imported: o.Imported().Path(),
		}
	case *types.Const:
		if o == nil {
			return gotypes.Const{}
		}
		return gotypes.Const{
			Obj:  obj(o, false),
			Kind: gotypes.ConstKind(o.Val().Kind()),
		}
	case *types.TypeName:
		if o == nil {
			return gotypes.TypeName{}
		}
		return gotypes.TypeName{
			Obj: obj(o, true),
		}
	case *types.Var:
		if o == nil {
			return gotypes.Var{}
		}
		return gotypes.Var{
			Obj:       obj(o, false),
			Anonymous: o.Anonymous(),
			IsField:   o.IsField(),
		}
	case *types.Func:
		if o == nil {
			return gotypes.Func{}
		}
		return gotypes.Func{
			Obj: obj(o, false),
		}
	case *types.Label:
		if o == nil {
			return gotypes.Label{}
		}
		return gotypes.Label{
			Obj: obj(o, false),
		}
	case *types.Builtin:
		if o == nil {
			return gotypes.Builtin{}
		}
		return gotypes.Builtin{
			Obj:       obj(o, false),
			BuiltinId: gotypes.BuiltinNames[o.Name()],
		}
	case *types.Nil:
		if o == nil {
			return gotypes.Nil{}
		}
		return gotypes.Nil{
			Obj: obj(o, false),
		}
	}
	panic(fmt.Sprintf("can't convert %T", o))
}

func Type(t types.Type, definition bool) gotypes.Type {
	if t == nil {
		return nil
	}
	switch t := t.(type) {
	case *types.Basic:
		return gotypes.Basic{
			Kind: gotypes.BasicKind(t.Kind()),
			Info: gotypes.BasicInfo(t.Info()),
			Name: t.Name(),
		}
	case *types.Array:
		return gotypes.Array{
			Len:  t.Len(),
			Elem: Type(t.Elem(), false),
		}
	case *types.Slice:
		return gotypes.Slice{
			Elem: Type(t.Elem(), false),
		}
	case *types.Struct:
		var fields []gotypes.Var
		var tags []string
		for i := 0; i < t.NumFields(); i++ {
			if !t.Field(i).Exported() {
				continue
			}
			fields = append(fields, Object(t.Field(i)).(gotypes.Var))
			tags = append(tags, t.Tag(i))
		}
		return gotypes.Struct{
			Fields: fields,
			Tags:   tags,
		}
	case *types.Pointer:
		return gotypes.Pointer{
			Elem: Type(t.Elem(), false),
		}
	case *types.Tuple:
		var vars []gotypes.Var
		for i := 0; i < t.Len(); i++ {
			vars = append(vars, Object(t.At(i)).(gotypes.Var))
		}
		return gotypes.Tuple{
			Vars: vars,
		}
	case *types.Signature:
		return gotypes.Signature{
			Recv:     Object(t.Recv()).(gotypes.Var),
			Params:   Type(t.Params(), false).(gotypes.Tuple),
			Results:  Type(t.Results(), false).(gotypes.Tuple),
			Variadic: t.Variadic(),
		}
	case *types.Interface:
		var methods []gotypes.Func
		var embeddeds []gotypes.Reference
		var allMethods []gotypes.Func
		for i := 0; i < t.NumExplicitMethods(); i++ {
			if !t.ExplicitMethod(i).Exported() {
				continue
			}
			methods = append(methods, Object(t.ExplicitMethod(i)).(gotypes.Func))
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			if !t.Embedded(i).Obj().Exported() {
				continue
			}
			embeddeds = append(embeddeds, Type(t.Embedded(i), false).(gotypes.Reference))
		}
		for i := 0; i < t.NumMethods(); i++ {
			if !t.Method(i).Exported() {
				continue
			}
			allMethods = append(allMethods, Object(t.Method(i)).(gotypes.Func))
		}
		return gotypes.Interface{
			Methods:    methods,
			Embeddeds:  embeddeds,
			AllMethods: allMethods,
		}
	case *types.Map:
		return gotypes.Map{
			Key:  Type(t.Key(), false),
			Elem: Type(t.Elem(), false),
		}
	case *types.Chan:
		return gotypes.Chan{
			Dir:  gotypes.ChanDir(t.Dir()),
			Elem: Type(t.Elem(), false),
		}
	case *types.Named:
		var path string
		if t.Obj().Pkg() != nil {
			path = t.Obj().Pkg().Path()
		}
		id := gotypes.Identifier{
			Path: path,
			Name: t.Obj().Name(),
		}
		if !definition {
			// Only return a gotypes.Named for type definitions. All other named types return a gotypes.Reference
			return gotypes.Reference(id)
		}
		var methods []gotypes.Func
		for i := 0; i < t.NumMethods(); i++ {
			if !t.Method(i).Exported() {
				continue
			}
			methods = append(methods, Object(t.Method(i)).(gotypes.Func))
		}
		return gotypes.Named{
			Type:    Type(t.Underlying(), false),
			Methods: methods,
		}
	}
	panic(fmt.Sprintf("can't convert %T", t))
}
