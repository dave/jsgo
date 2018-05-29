package convert

import (
	"go/types"

	"github.com/dave/jsgo/server/frizz/gotypes"
)

func Type(t types.Type, stack *[]types.Type) gotypes.Type {
	for _, stacked := range *stack {
		if t == stacked {
			return gotypes.Circular{}
		}
	}
	*stack = append(*stack, t)
	defer func() {
		*stack = (*stack)[:len(*stack)-1]
	}()
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
			Elem: Type(t.Elem(), stack),
		}
	case *types.Slice:
		return gotypes.Slice{
			Elem: Type(t.Elem(), stack),
		}
	case *types.Struct:
		var fields []gotypes.Var
		var tags []string
		for i := 0; i < t.NumFields(); i++ {
			if !t.Field(i).Exported() {
				continue
			}
			fields = append(fields, Var(t.Field(i), stack))
			tags = append(tags, t.Tag(i))
		}
		return gotypes.Struct{
			Fields: fields,
			Tags:   tags,
		}
	case *types.Pointer:
		return gotypes.Pointer{
			Elem: Type(t.Elem(), stack),
		}
	case *types.Tuple:
		var vars []gotypes.Var
		for i := 0; i < t.Len(); i++ {
			vars = append(vars, Var(t.At(i), stack))
		}
		return gotypes.Tuple{
			Vars: vars,
		}
	case *types.Signature:
		return gotypes.Signature{
			Recv:     Var(t.Recv(), stack),
			Params:   Type(t.Params(), stack).(gotypes.Tuple),
			Results:  Type(t.Results(), stack).(gotypes.Tuple),
			Variadic: t.Variadic(),
		}
	case *types.Interface:
		var methods []gotypes.Func
		var embeddeds []gotypes.Named
		var allMethods []gotypes.Func
		for i := 0; i < t.NumExplicitMethods(); i++ {
			if !t.ExplicitMethod(i).Exported() {
				continue
			}
			methods = append(methods, Func(t.ExplicitMethod(i), stack))
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			if !t.Embedded(i).Obj().Exported() {
				continue
			}
			embeddeds = append(embeddeds, Type(t.Embedded(i), stack).(gotypes.Named))
		}
		for i := 0; i < t.NumMethods(); i++ {
			if !t.Method(i).Exported() {
				continue
			}
			allMethods = append(allMethods, Func(t.Method(i), stack))
		}
		return gotypes.Interface{
			Methods:    methods,
			Embeddeds:  embeddeds,
			AllMethods: allMethods,
		}
	case *types.Map:
		return gotypes.Map{
			Key:  Type(t.Key(), stack),
			Elem: Type(t.Elem(), stack),
		}
	case *types.Chan:
		return gotypes.Chan{
			Dir:  gotypes.ChanDir(t.Dir()),
			Elem: Type(t.Elem(), stack),
		}
	case *types.Named:
		var methods []gotypes.Func
		for i := 0; i < t.NumMethods(); i++ {
			if !t.Method(i).Exported() {
				continue
			}
			methods = append(methods, Func(t.Method(i), stack))
		}
		var path string
		if t.Obj().Pkg() != nil {
			path = t.Obj().Pkg().Path()
		}
		return gotypes.Named{
			Obj: gotypes.TypeName{
				Obj: gotypes.Obj{
					Pkg:  path,
					Name: t.Obj().Name(),
					Typ:  Type(t.Obj().Type(), stack),
				},
			},
			Type:    Type(t.Underlying(), stack),
			Methods: methods,
		}
	}
	// notest
	return nil
}

func Func(f *types.Func, stack *[]types.Type) gotypes.Func {
	if f == nil {
		// notest
		return gotypes.Func{}
	}
	var path string
	if f.Pkg() != nil {
		path = f.Pkg().Path()
	}
	return gotypes.Func{
		Obj: gotypes.Obj{
			Pkg:  path,
			Name: f.Name(),
			Typ:  Type(f.Type(), stack),
		},
	}
}

func Var(v *types.Var, stack *[]types.Type) gotypes.Var {
	if v == nil {
		return gotypes.Var{}
	}
	var path string
	if v.Pkg() != nil {
		path = v.Pkg().Path()
	}
	return gotypes.Var{
		Obj: gotypes.Obj{
			Pkg:  path,
			Name: v.Name(),
			Typ:  Type(v.Type(), stack),
		},
		Anonymous: v.Anonymous(),
		IsField:   v.IsField(),
	}
}
