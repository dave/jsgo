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
	type spec struct {
		code     string
		expected string
	}
	tests := map[string]spec{
		"simple": {
			`type Foo int`,
			`Foo: gotypes.Named{Type:gotypes.Basic{Kind:2, Info:2, Name:"int"}, Methods:[]gotypes.Func(nil)}`,
		},
		"ignore non-global": {
			`type Foo string
			func f() {
				type Bar string
			}`,
			`Foo: gotypes.Named{Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}, Methods:[]gotypes.Func(nil)}`,
		},
		"ignore non-exported": {
			`type Foo string
			 type bar string`,
			`Foo: gotypes.Named{Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}, Methods:[]gotypes.Func(nil)}`,
		},
		"ignore non-exported methods": {
			`type Foo string
			 func (Foo) bar(){}
			 func (Foo) Baz(){}`,
			`Foo: gotypes.Named{Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}, Methods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Baz"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Variadic:false}}}}}`,
		},
		"ignore non-exported interface methods": {
			`type Foo interface {
				foo()
				Bar()
			}`,
			`Foo: gotypes.Named{Type:gotypes.Interface{Methods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Bar"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Variadic:false}}}}, Embeddeds:[]gotypes.Reference(nil), AllMethods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Bar"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Variadic:false}}}}}, Methods:[]gotypes.Func(nil)}`,
		},
		"ignore non-exported interface embeds": {
			`type foo interface{}
			type Bar interface{}
			type Baz interface {
				foo
				Bar
			}`,
			`Bar: gotypes.Named{Type:gotypes.Interface{Methods:[]gotypes.Func(nil), Embeddeds:[]gotypes.Reference(nil), AllMethods:[]gotypes.Func(nil)}, Methods:[]gotypes.Func(nil)}
			Baz: gotypes.Named{Type:gotypes.Interface{Methods:[]gotypes.Func(nil), Embeddeds:[]gotypes.Reference{gotypes.Reference{Path:"foo", Name:"Bar"}}, AllMethods:[]gotypes.Func(nil)}, Methods:[]gotypes.Func(nil)}`,
		},
		"include exported alias of non-exported type": {
			`type foo struct {
				Bar string
			}
			type Baz foo`,
			`Baz: gotypes.Named{Type:gotypes.Struct{Fields:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Bar"}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:true}}, Tags:[]string{""}}, Methods:[]gotypes.Func(nil)}`,
		},

		// TODO: Will this break things?
		"include exported pointer to non-exported type": {
			`type foo struct {
				Bar string
			}
			type Baz *foo`,
			`Baz: gotypes.Named{Type:gotypes.Pointer{Elem:gotypes.Reference{Path:"foo", Name:"foo"}}, Methods:[]gotypes.Func(nil)}`,
		},

		"two types": {
			`type Foo int64
			type Bar rune`,
			`Bar: gotypes.Named{Type:gotypes.Basic{Kind:5, Info:2, Name:"rune"}, Methods:[]gotypes.Func(nil)}
			Foo: gotypes.Named{Type:gotypes.Basic{Kind:6, Info:2, Name:"int64"}, Methods:[]gotypes.Func(nil)}`,
		},
		"alias": {
			`type Foo int
			type Bar Foo`,
			`Bar: gotypes.Named{Type:gotypes.Basic{Kind:2, Info:2, Name:"int"}, Methods:[]gotypes.Func(nil)}
			Foo: gotypes.Named{Type:gotypes.Basic{Kind:2, Info:2, Name:"int"}, Methods:[]gotypes.Func(nil)}`,
		},
		"struct": {
			`type Foo struct {
				Bar string
				baz string
			}`,
			`Foo: gotypes.Named{Type:gotypes.Struct{Fields:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Bar"}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:true}}, Tags:[]string{""}}, Methods:[]gotypes.Func(nil)}`,
		},
		"array": {
			`type Foo [2]string`,
			`Foo: gotypes.Named{Type:gotypes.Array{Len:2, Elem:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Methods:[]gotypes.Func(nil)}`,
		},
		"slice": {
			`type Foo []int`,
			`Foo: gotypes.Named{Type:gotypes.Slice{Elem:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Methods:[]gotypes.Func(nil)}`,
		},
		"pointer": {
			`type Foo *int`,
			`Foo: gotypes.Named{Type:gotypes.Pointer{Elem:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Methods:[]gotypes.Func(nil)}`,
		},
		"func type": {
			`type Foo func(int)`,
			`Foo: gotypes.Named{Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"", Name:""}, Type:gotypes.Type(nil)}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Anonymous:false, IsField:false}}}, Results:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Variadic:false}, Methods:[]gotypes.Func(nil)}`,
		},
		"interface": {
			`type Foo interface{
				A() string
				B(int, ...string)
			}`,
			`Foo: gotypes.Named{Type:gotypes.Interface{Methods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"A"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:false}}}, Variadic:false}}}, gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"B"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Anonymous:false, IsField:false}, gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Slice{Elem:gotypes.Basic{Kind:17, Info:32, Name:"string"}}}, Anonymous:false, IsField:false}}}, Results:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Variadic:true}}}}, Embeddeds:[]gotypes.Reference(nil), AllMethods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"A"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:false}}}, Variadic:false}}}, gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"B"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Anonymous:false, IsField:false}, gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Slice{Elem:gotypes.Basic{Kind:17, Info:32, Name:"string"}}}, Anonymous:false, IsField:false}}}, Results:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Variadic:true}}}}}, Methods:[]gotypes.Func(nil)}`,
		},
		"interface with embeds": {
			`type Foo interface{
				A() string
			}
			type Bar interface{
				Foo
				B() string
			}`,
			`Bar: gotypes.Named{Type:gotypes.Interface{Methods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"B"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Bar"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:false}}}, Variadic:false}}}}, Embeddeds:[]gotypes.Reference{gotypes.Reference{Path:"foo", Name:"Foo"}}, AllMethods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"A"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:false}}}, Variadic:false}}}, gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"B"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Bar"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:false}}}, Variadic:false}}}}}, Methods:[]gotypes.Func(nil)}
			Foo: gotypes.Named{Type:gotypes.Interface{Methods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"A"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:false}}}, Variadic:false}}}}, Embeddeds:[]gotypes.Reference(nil), AllMethods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"A"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:false}}}, Variadic:false}}}}}, Methods:[]gotypes.Func(nil)}`,
		},
		"map": {
			`type Foo map[string]int`,
			`Foo: gotypes.Named{Type:gotypes.Map{Key:gotypes.Basic{Kind:17, Info:32, Name:"string"}, Elem:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Methods:[]gotypes.Func(nil)}`,
		},
		"chan": {
			`type Foo chan<- int`,
			`Foo: gotypes.Named{Type:gotypes.Chan{Dir:1, Elem:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Methods:[]gotypes.Func(nil)}`,
		},
		"methods": {
			`type Foo struct{}
			func (f Foo) Bar() int { return 1 }`,
			`Foo: gotypes.Named{Type:gotypes.Struct{Fields:[]gotypes.Var(nil), Tags:[]string(nil)}, Methods:[]gotypes.Func{gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Bar"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"f"}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:""}, Type:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Anonymous:false, IsField:false}}}, Variadic:false}}}}}`,
		},
		"func": {
			`func Foo() {}`,
			`Foo: gotypes.Func{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Foo"}, Type:gotypes.Signature{Recv:gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"", Name:""}, Type:gotypes.Type(nil)}, Anonymous:false, IsField:false}, Params:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Results:gotypes.Tuple{Vars:[]gotypes.Var(nil)}, Variadic:false}}}`,
		},
		"recursive": {
			`type Foo struct{ *Foo }`,
			`Foo: gotypes.Named{Type:gotypes.Struct{Fields:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Foo"}, Type:gotypes.Pointer{Elem:gotypes.Reference{Path:"foo", Name:"Foo"}}}, Anonymous:true, IsField:true}}, Tags:[]string{""}}, Methods:[]gotypes.Func(nil)}`,
		},
		"var int": {
			`var Foo = 1`,
			`Foo: gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Foo"}, Type:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Anonymous:false, IsField:false}`,
		},
		"var alias": {
			`type Foo string
			var Bar = Foo("foo")`,
			`Bar: gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Bar"}, Type:gotypes.Reference{Path:"foo", Name:"Foo"}}, Anonymous:false, IsField:false}
			Foo: gotypes.Named{Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}, Methods:[]gotypes.Func(nil)}`,
		},
		"var struct": {
			`var Foo = struct{
				Bar int
				Baz string
			}{
				Bar: 1, 
				Baz: "baz",
			}`,
			`Foo: gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Foo"}, Type:gotypes.Struct{Fields:[]gotypes.Var{gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Bar"}, Type:gotypes.Basic{Kind:2, Info:2, Name:"int"}}, Anonymous:false, IsField:true}, gotypes.Var{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Baz"}, Type:gotypes.Basic{Kind:17, Info:32, Name:"string"}}, Anonymous:false, IsField:true}}, Tags:[]string{"", ""}}}, Anonymous:false, IsField:false}`,
		},
		"const": {
			`const Foo = "foo"`,
			`Foo: gotypes.Const{Obj:gotypes.Obj{Identifier:gotypes.Identifier{Path:"foo", Name:"Foo"}, Type:gotypes.Basic{Kind:24, Info:96, Name:"untyped string"}}, Kind:2}`,
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
			if tn, ok := o.(gotypes.TypeName); ok {
				fmt.Fprintf(buf, "%s: %#v\n", o.Id().Name, tn.Type)
			} else {
				fmt.Fprintf(buf, "%s: %#v\n", o.Id().Name, o)
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
