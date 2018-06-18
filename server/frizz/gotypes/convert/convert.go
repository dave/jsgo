package convert

import (
	"go/types"

	"fmt"

	"sort"

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
			return nil
		}
		return &gotypes.PkgName{
			Obj:      obj(o, false),
			Imported: o.Imported().Path(),
		}
	case *types.Const:
		if o == nil {
			return nil
		}
		return &gotypes.Const{
			Obj:  obj(o, false),
			Kind: gotypes.ConstKind(o.Val().Kind()),
		}
	case *types.TypeName:
		if o == nil {
			return nil
		}
		return &gotypes.TypeName{
			Obj: obj(o, true),
		}
	case *types.Var:
		if o == nil {
			return nil
		}
		return &gotypes.Var{
			Obj:       obj(o, false),
			Anonymous: o.Anonymous(),
			IsField:   o.IsField(),
		}
	case *types.Func:
		if o == nil {
			return nil
		}
		return &gotypes.Func{
			Obj: obj(o, false),
		}
	case *types.Label:
		if o == nil {
			return nil
		}
		return &gotypes.Label{
			Obj: obj(o, false),
		}
	case *types.Builtin:
		if o == nil {
			return nil
		}
		return &gotypes.Builtin{
			Obj:       obj(o, false),
			BuiltinId: gotypes.BuiltinNames[o.Name()],
		}
	case *types.Nil:
		if o == nil {
			return nil
		}
		return &gotypes.Nil{
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
		return &gotypes.Basic{
			Kind: gotypes.BasicKind(t.Kind()),
			Info: gotypes.BasicInfo(t.Info()),
			Name: t.Name(),
		}
	case *types.Array:
		return &gotypes.Array{
			Len:  t.Len(),
			Elem: Type(t.Elem(), false),
		}
	case *types.Slice:
		return &gotypes.Slice{
			Elem: Type(t.Elem(), false),
		}
	case *types.Struct:
		var fields []*gotypes.Var
		var tags []string
		for i := 0; i < t.NumFields(); i++ {
			if !t.Field(i).Exported() {
				continue
			}
			var v *gotypes.Var
			if o := Object(t.Field(i)); o != nil {
				v = o.(*gotypes.Var)
			}
			fields = append(fields, v)
			tags = append(tags, t.Tag(i))
		}
		return &gotypes.Struct{
			Fields: fields,
			Tags:   tags,
		}
	case *types.Pointer:
		return &gotypes.Pointer{
			Elem: Type(t.Elem(), false),
		}
	case *types.Tuple:
		var vars []*gotypes.Var
		for i := 0; i < t.Len(); i++ {
			var v *gotypes.Var
			if o := Object(t.At(i)); o != nil {
				v = o.(*gotypes.Var)
			}
			vars = append(vars, v)
		}
		return &gotypes.Tuple{
			Vars: vars,
		}
	case *types.Signature:
		var recv *gotypes.Var
		var params, results *gotypes.Tuple
		if v := Object(t.Recv()); v != nil {
			recv = v.(*gotypes.Var)
		}
		if v := Type(t.Params(), false); v != nil {
			params = v.(*gotypes.Tuple)
		}
		if v := Type(t.Results(), false); v != nil {
			results = v.(*gotypes.Tuple)
		}
		return &gotypes.Signature{
			Recv:     recv,
			Params:   params,
			Results:  results,
			Variadic: t.Variadic(),
		}
	case *types.Interface:
		var methods []*gotypes.Func
		var embeddeds []*gotypes.Reference
		var allMethods []*gotypes.Func
		for i := 0; i < t.NumExplicitMethods(); i++ {
			if !t.ExplicitMethod(i).Exported() {
				continue
			}
			var f *gotypes.Func
			if o := Object(t.ExplicitMethod(i)); o != nil {
				f = o.(*gotypes.Func)
			}
			methods = append(methods, f)
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			if !t.Embedded(i).Obj().Exported() {
				continue
			}
			var ref *gotypes.Reference
			if t := Type(t.Embedded(i), false); t != nil {
				ref = t.(*gotypes.Reference)
			}
			embeddeds = append(embeddeds, ref)
		}
		for i := 0; i < t.NumMethods(); i++ {
			if !t.Method(i).Exported() {
				continue
			}
			var f *gotypes.Func
			if o := Object(t.Method(i)); o != nil {
				f = o.(*gotypes.Func)
			}
			allMethods = append(allMethods, f)
		}
		sort.Slice(methods, func(i, j int) bool {
			return compareId(methods[i].Obj.Identifier, methods[j].Obj.Identifier)
		})
		sort.Slice(embeddeds, func(i, j int) bool {
			return compareId(embeddeds[i].Identifier, embeddeds[j].Identifier)
		})
		sort.Slice(allMethods, func(i, j int) bool {
			return compareId(allMethods[i].Obj.Identifier, allMethods[j].Obj.Identifier)
		})
		return &gotypes.Interface{
			Methods:    methods,
			Embeddeds:  embeddeds,
			AllMethods: allMethods,
		}
	case *types.Map:
		return &gotypes.Map{
			Key:  Type(t.Key(), false),
			Elem: Type(t.Elem(), false),
		}
	case *types.Chan:
		return &gotypes.Chan{
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
			return &gotypes.Reference{id}
		}
		var methods []*gotypes.Func
		for i := 0; i < t.NumMethods(); i++ {
			if !t.Method(i).Exported() {
				continue
			}
			var f *gotypes.Func
			if o := Object(t.Method(i)); o != nil {
				f = o.(*gotypes.Func)
			}
			methods = append(methods, f)
		}
		sort.Slice(methods, func(i, j int) bool {
			return compareId(methods[i].Obj.Identifier, methods[j].Obj.Identifier)
		})
		return &gotypes.Named{
			Type:    Type(t.Underlying(), false),
			Methods: methods,
		}
	}
	panic(fmt.Sprintf("can't convert %T", t))
}

func compareId(i, j gotypes.Identifier) bool {
	return fmt.Sprintf("%q.%s", i.Path, i.Name) > fmt.Sprintf("%q.%s", j.Path, j.Name)
}
