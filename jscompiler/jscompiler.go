package jscompiler

import (
	"encoding/gob"
	"fmt"
	"go/build"
	"go/parser"
	"go/types"

	"frizz.io/edit/assets"

	"os"

	"sync"

	"go/token"

	"frizz.io/config"
	"github.com/cskr/pubsub"
	"github.com/dave/patsy/vos"
	"github.com/gopherjs/gopherjs/compiler"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/loader"
)

type JsCompiler struct {
	cacheM *sync.RWMutex
	cache  map[string]*compiler.Archive

	env vos.Env

	ps *pubsub.PubSub
}

func New(env vos.Env) *JsCompiler {
	j := &JsCompiler{}
	j.cacheM = &sync.RWMutex{}
	j.cache = make(map[string]*compiler.Archive)
	j.ps = pubsub.New(1)
	j.env = env
	return j
}

func (j *JsCompiler) getCache(path string) (*compiler.Archive, bool) {
	j.cacheM.RLock()
	defer j.cacheM.RUnlock()
	a, ok := j.cache[path]
	return a, ok
}

func (j *JsCompiler) setCache(path string, a *compiler.Archive) {
	j.cacheM.Lock()
	defer j.cacheM.Unlock()
	j.cache[path] = a
	j.ps.Pub(a, path)
}

// Hint provides a hint that a cascade of requests for packages is coming soon...
func (j *JsCompiler) Main(path string, source []byte) (*loader.Program, error) {

	go j.BuildMain(path, source)

	conf := loader.Config{}
	conf.Fset = token.NewFileSet()
	conf.ParserMode = parser.ImportsOnly
	conf.Build = func() *build.Context { c := build.Default; return &c }() // make a copy of build.Default
	conf.Build.GOPATH = j.env.Getenv("GOPATH")
	conf.Build.BuildTags = []string{"js"}
	conf.AllowErrors = true
	conf.TypeChecker.Error = func(e error) {}
	f, err := parser.ParseFile(conf.Fset, "main.go", source, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	conf.CreateFromFiles(path+"$main", f)

	prog, err := conf.Load()
	if err != nil {
		return nil, err
	}

	return prog, nil
}

func (j *JsCompiler) BuildMain(path string, source []byte) {

	conf := loader.Config{}
	conf.Fset = token.NewFileSet()
	conf.ParserMode = parser.ParseComments
	conf.Build = func() *build.Context { c := build.Default; return &c }() // make a copy of build.Default
	conf.Build.GOPATH = j.env.Getenv("GOPATH")
	conf.Build.BuildTags = []string{"js"}
	conf.AllowErrors = true
	conf.TypeChecker.Error = func(e error) {}
	f, err := parser.ParseFile(conf.Fset, "main.go", source, parser.ParseComments)
	if err != nil {
		panic(err.Error())
	}
	conf.CreateFromFiles(path+"$main", f)
	prog, err := conf.Load()
	if err != nil {
		panic(err.Error())
	}

	var importContext *compiler.ImportContext
	importContext = &compiler.ImportContext{
		Packages: make(map[string]*types.Package),
		Import: func(path string) (*compiler.Archive, error) {

			// find in local cache
			if a, ok := j.getCache(path); ok {
				return a, nil
			}

			pi := prog.Package(path)
			importContext.Packages[path] = pi.Pkg

			// find in standard library cache
			a, err := openArchive(path)
			if err != nil {
				return nil, err
			}
			if a != nil {
				j.setCache(path, a)
				return a, nil
			}

			// compile package
			a, err = compiler.Compile(path, pi.Files, prog.Fset, importContext, !config.DEV)
			if err != nil {
				return nil, err
			}
			j.setCache(path, a)
			return a, nil
		},
	}

	if _, err := importContext.Import(path + "$main"); err != nil {
		panic(err.Error())
	}

}

// Hint provides a hint that a cascade of requests for packages is coming soon...
func (j *JsCompiler) Hint(paths ...string) (*loader.Program, error) {

	go j.Build(paths...)

	conf := loader.Config{}
	for _, path := range paths {
		conf.Import(path)
	}
	conf.ParserMode = parser.ImportsOnly
	conf.Build = func() *build.Context { c := build.Default; return &c }() // make a copy of build.Default
	conf.Build.GOPATH = j.env.Getenv("GOPATH")
	conf.Build.BuildTags = []string{"js"}
	conf.AllowErrors = true
	conf.TypeChecker.Error = func(e error) {}
	prog, err := conf.Load()
	if err != nil {
		return nil, err
	}

	return prog, nil
}

func (j *JsCompiler) Build(paths ...string) {
	conf := loader.Config{}
	for _, path := range paths {
		conf.Import(path)
	}
	conf.ParserMode = parser.ParseComments
	conf.Build = func() *build.Context { c := build.Default; return &c }() // make a copy of build.Default
	conf.Build.GOPATH = j.env.Getenv("GOPATH")
	conf.Build.BuildTags = []string{"js"}
	prog, err := conf.Load()
	if err != nil {
		panic(err.Error())
	}

	var importContext *compiler.ImportContext
	importContext = &compiler.ImportContext{
		Packages: make(map[string]*types.Package),
		Import: func(path string) (*compiler.Archive, error) {

			// find in local cache
			if a, ok := j.getCache(path); ok {
				return a, nil
			}

			pi := prog.Package(path)
			importContext.Packages[path] = pi.Pkg

			// find in standard library cache
			a, err := openArchive(path)
			if err != nil {
				return nil, err
			}
			if a != nil {
				j.setCache(path, a)
				return a, nil
			}

			// compile package
			a, err = compiler.Compile(path, pi.Files, prog.Fset, importContext, !config.DEV)
			if err != nil {
				return nil, err
			}
			j.setCache(path, a)
			return a, nil
		},
	}

	for _, path := range paths {
		if _, err := importContext.Import(path); err != nil {
			panic(err.Error())
		}
	}

}

func openArchive(path string) (*compiler.Archive, error) {

	var filename string
	if config.DEV {
		filename = fmt.Sprintf("pkg/%s.a", path)
	} else {
		filename = fmt.Sprintf("pkg_min/%s.a", path)
	}

	f, err := assets.Assets.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	defer f.Close()

	archive := new(compiler.Archive)
	if err := gob.NewDecoder(f).Decode(archive); err != nil {
		return nil, errors.WithStack(err)
	}
	return archive, nil

}

func (j *JsCompiler) getOrSubscribe(path string) (*compiler.Archive, chan interface{}) {
	j.cacheM.Lock()
	defer j.cacheM.Unlock()
	a, ok := j.cache[path]
	if ok {
		return a, nil
	}
	return nil, j.ps.SubOnce(path)
}

// Get requests a package. This will block until the package is ready.
func (j *JsCompiler) Get(path string) *compiler.Archive {

	a, c := j.getOrSubscribe(path)
	if a != nil {
		return a
	}

	i := <-c
	return i.(*compiler.Archive)

}
