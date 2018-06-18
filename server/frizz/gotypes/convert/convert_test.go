package convert

import (
	"go/parser"
	"go/token"
	"testing"

	"go/ast"
	"go/types"

	"bytes"
	"fmt"

	"strings"

	"sort"

	"regexp"

	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/davecgh/go-spew/spew"
)

func TestFoo(t *testing.T) {
	foo := func(object types.Object) {
		if object == nil {
			return
		}
		if object.(*types.Var) == nil {
			return
		}
		fmt.Println(object.Pkg().Name())
	}
	foo((*types.Var)(nil))
}

func TestConvertType(t *testing.T) {
	spew.Config = spew.ConfigState{DisablePointerAddresses: true, DisableMethods: true}
	type spec struct {
		code     string
		expected string
	}
	tests := map[string]spec{
		"simple": {
			`type Foo int`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int} Methods:([]*gotypes.Func)<nil>}`,
		},
		"ignore non-global": {
			`type Foo string
			func f() {
				type Bar string
			}`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string} Methods:([]*gotypes.Func)<nil>}`,
		},
		"ignore non-exported": {
			`type Foo string
			 type bar string`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string} Methods:([]*gotypes.Func)<nil>}`,
		},
		"ignore non-exported methods": {
			`type Foo string
			 func (Foo) bar(){}
			 func (Foo) Baz(){}`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string} Methods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Baz} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Variadic:(bool)false}}}]}`,
		},
		"ignore non-exported interface methods": {
			`type Foo interface {
				foo()
				Bar()
			}`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Interface){Methods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Variadic:(bool)false}}}] Embeddeds:([]*gotypes.Reference)<nil> AllMethods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Variadic:(bool)false}}}]} Methods:([]*gotypes.Func)<nil>}`,
		},
		"ignore non-exported interface embeds": {
			`type foo interface{}
			type Bar interface{}
			type Baz interface {
				foo
				Bar
			}`,
			`Bar: (*gotypes.Named){Type:(*gotypes.Interface){Methods:([]*gotypes.Func)<nil> Embeddeds:([]*gotypes.Reference)<nil> AllMethods:([]*gotypes.Func)<nil>} Methods:([]*gotypes.Func)<nil>}
			Baz: (*gotypes.Named){Type:(*gotypes.Interface){Methods:([]*gotypes.Func)<nil> Embeddeds:([]*gotypes.Reference)[<*>{Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar}}] AllMethods:([]*gotypes.Func)<nil>} Methods:([]*gotypes.Func)<nil>}`,
		},
		"include exported alias of non-exported type": {
			`type foo struct {
				Bar string
			}
			type Baz foo`,
			`Baz: (*gotypes.Named){Type:(*gotypes.Struct){Fields:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)true}] Tags:([]string)[]} Methods:([]*gotypes.Func)<nil>}`,
		},

		// TODO: Will this break things?
		"include exported pointer to non-exported type": {
			`type foo struct {
				Bar string
			}
			type Baz *foo`,
			`Baz: (*gotypes.Named){Type:(*gotypes.Pointer){Elem:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)foo}}} Methods:([]*gotypes.Func)<nil>}`,
		},

		"two types": {
			`type Foo int64
			type Bar rune`,
			`Bar: (*gotypes.Named){Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)5 Info:(gotypes.BasicInfo)2 Name:(string)rune} Methods:([]*gotypes.Func)<nil>}
			Foo: (*gotypes.Named){Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)6 Info:(gotypes.BasicInfo)2 Name:(string)int64} Methods:([]*gotypes.Func)<nil>}`,
		},
		"alias": {
			`type Foo int
			type Bar Foo`,
			`Bar: (*gotypes.Named){Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int} Methods:([]*gotypes.Func)<nil>}
			Foo: (*gotypes.Named){Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int} Methods:([]*gotypes.Func)<nil>}`,
		},
		"struct": {
			`type Foo struct {
				Bar string
				baz string
			}`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Struct){Fields:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)true}] Tags:([]string)[]} Methods:([]*gotypes.Func)<nil>}`,
		},
		"array": {
			`type Foo [2]string`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Array){Len:(int64)2 Elem:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Methods:([]*gotypes.Func)<nil>}`,
		},
		"slice": {
			`type Foo []int`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Slice){Elem:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Methods:([]*gotypes.Func)<nil>}`,
		},
		"pointer": {
			`type Foo *int`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Pointer){Elem:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Methods:([]*gotypes.Func)<nil>}`,
		},
		"func type": {
			`type Foo func(int)`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Signature){Recv:(*gotypes.Var)<nil> Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Anonymous:(bool)false IsField:(bool)false}]} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Variadic:(bool)false} Methods:([]*gotypes.Func)<nil>}`,
		},
		"interface": {
			`type Foo interface{
				A() string
				B(int, ...string)
			}`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Interface){Methods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)B} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Anonymous:(bool)false IsField:(bool)false} <*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Slice){Elem:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}}} Anonymous:(bool)false IsField:(bool)false}]} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Variadic:(bool)true}}} <*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)A} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)false}]} Variadic:(bool)false}}}] Embeddeds:([]*gotypes.Reference)<nil> AllMethods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)B} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Anonymous:(bool)false IsField:(bool)false} <*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Slice){Elem:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}}} Anonymous:(bool)false IsField:(bool)false}]} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Variadic:(bool)true}}} <*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)A} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)false}]} Variadic:(bool)false}}}]} Methods:([]*gotypes.Func)<nil>}`,
		},
		"interface with embeds": {
			`type Foo interface{
				A() string
			}
			type Bar interface{
				Foo
				B() string
			}`,
			`Bar: (*gotypes.Named){Type:(*gotypes.Interface){Methods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)B} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)false}]} Variadic:(bool)false}}}] Embeddeds:([]*gotypes.Reference)[<*>{Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}] AllMethods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)B} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)false}]} Variadic:(bool)false}}} <*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)A} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)false}]} Variadic:(bool)false}}}]} Methods:([]*gotypes.Func)<nil>}
			Foo: (*gotypes.Named){Type:(*gotypes.Interface){Methods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)A} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)false}]} Variadic:(bool)false}}}] Embeddeds:([]*gotypes.Reference)<nil> AllMethods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)A} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)false}]} Variadic:(bool)false}}}]} Methods:([]*gotypes.Func)<nil>}`,
		},
		"map": {
			`type Foo map[string]int`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Map){Key:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string} Elem:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Methods:([]*gotypes.Func)<nil>}`,
		},
		"chan": {
			`type Foo chan<- int`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Chan){Dir:(gotypes.ChanDir)1 Elem:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Methods:([]*gotypes.Func)<nil>}`,
		},
		"methods": {
			`type Foo struct{}
			func (f Foo) Bar() int { return 1 }`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Struct){Fields:([]*gotypes.Var)<nil> Tags:([]string)<nil>} Methods:([]*gotypes.Func)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar} Type:(*gotypes.Signature){Recv:(*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)f} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false} Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Anonymous:(bool)false IsField:(bool)false}]} Variadic:(bool)false}}}]}`,
		},
		"func": {
			`func Foo() {}`,
			`Foo: (*gotypes.Func){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo} Type:(*gotypes.Signature){Recv:(*gotypes.Var)<nil> Params:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Results:(*gotypes.Tuple){Vars:([]*gotypes.Var)<nil>} Variadic:(bool)false}}}`,
		},
		"recursive": {
			`type Foo struct{ *Foo }`,
			`Foo: (*gotypes.Named){Type:(*gotypes.Struct){Fields:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo} Type:(*gotypes.Pointer){Elem:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}}} Anonymous:(bool)true IsField:(bool)true}] Tags:([]string)[]} Methods:([]*gotypes.Func)<nil>}`,
		},
		"var int": {
			`var Foo = 1`,
			`Foo: (*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Anonymous:(bool)false IsField:(bool)false}`,
		},
		"var alias": {
			`type Foo string
			var Bar = Foo("foo")`,
			`Bar: (*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar} Type:(*gotypes.Reference){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo}}} Anonymous:(bool)false IsField:(bool)false}
			Foo: (*gotypes.Named){Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string} Methods:([]*gotypes.Func)<nil>}`,
		},
		"var struct": {
			`var Foo = struct{
				Bar int
				Baz string
			}{
				Bar: 1, 
				Baz: "baz",
			}`,
			`Foo: (*gotypes.Var){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo} Type:(*gotypes.Struct){Fields:([]*gotypes.Var)[<*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Bar} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)2 Info:(gotypes.BasicInfo)2 Name:(string)int}} Anonymous:(bool)false IsField:(bool)true} <*>{Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Baz} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)17 Info:(gotypes.BasicInfo)32 Name:(string)string}} Anonymous:(bool)false IsField:(bool)true}] Tags:([]string)[ ]}} Anonymous:(bool)false IsField:(bool)false}`,
		},
		"const": {
			`const Foo = "foo"`,
			`Foo: (*gotypes.Const){Obj:(gotypes.Obj){Identifier:(gotypes.Identifier){Path:(string)foo Name:(string)Foo} Type:(*gotypes.Basic){Kind:(gotypes.BasicKind)24 Info:(gotypes.BasicInfo)96 Name:(string)untyped string}} Kind:(gotypes.ConstKind)2}`,
		},
	}
	const single = ""
	if single != "" {
		tests = map[string]spec{single: tests[single]}
	}
	for name, test := range tests {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "foo.go", []byte("package foo\n"+test.code), 0)
		if err != nil {
			t.Fatal(err)
		}
		tc := types.Config{}
		info := &types.Info{
			Types: map[ast.Expr]types.TypeAndValue{},
			Defs:  map[*ast.Ident]types.Object{},
		}

		p, err := tc.Check("foo", fset, []*ast.File{f}, info)
		if err != nil {
			t.Fatal(err)
		}

		var defs []gotypes.Object
		for _, v := range info.Defs {
			if v == nil {
				continue
			}
			if v.Parent() != p.Scope() {
				continue
			}
			if !v.Exported() {
				continue
			}
			o := Object(v)
			defs = append(defs, o)
		}
		sort.Slice(defs, func(i, j int) bool { return defs[i].Id().Name < defs[j].Id().Name })
		buf := &bytes.Buffer{}
		for _, o := range defs {
			if tn, ok := o.(*gotypes.TypeName); ok {
				fmt.Fprintf(buf, "%s: %s\n", o.Id().Name, spew.Sprintf("%#v", tn.Type))
			} else {
				fmt.Fprintf(buf, "%s: %s\n", o.Id().Name, spew.Sprintf("%#v", o))
			}
		}
		if strings.TrimSpace(buf.String()) != indent.ReplaceAllString(test.expected, "") {
			t.Fatalf("%s, got:\n%s", name, buf.String())
		}
	}
	if single != "" {
		t.Fatal("single test")
	}
}

var indent = regexp.MustCompile(`(?m)^\s*`)
