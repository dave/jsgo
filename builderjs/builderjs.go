package builderjs

import (
	"errors"
	"go/token"
	"go/types"
	"sort"

	"go/ast"
	"go/parser"

	"fmt"

	"github.com/gopherjs/gopherjs/compiler"
	"golang.org/x/tools/go/gcimporter15"
)

func BuildPackage(source map[string]string, deps []*compiler.Archive, minify bool) (*compiler.Archive, error) {

	archives := map[string]*compiler.Archive{}
	packages := map[string]*types.Package{}

	for _, a := range deps {
		archives[a.ImportPath] = a
		_, p, err := gcimporter.BImportData(token.NewFileSet(), packages, a.ExportData, a.ImportPath)
		if err != nil {
			return nil, err
		}
		packages[a.ImportPath] = p
	}

	importPath := "main"

	fset := token.NewFileSet()
	var files []*ast.File
	for name, contents := range source {
		f, err := parser.ParseFile(fset, name, contents, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	importContext := &compiler.ImportContext{
		Packages: packages,
		Import: func(path string) (*compiler.Archive, error) {
			if path == "main" {
				return nil, errors.New("can't import main package")
			}
			a, ok := archives[path]
			if !ok {
				return nil, fmt.Errorf("%s not found", path)
			}
			return a, nil
		},
	}

	// TODO: Remove this when https://github.com/gopherjs/gopherjs/pull/742 is merged
	// Files must be in the same order to get reproducible JS
	sort.Slice(files, func(i, j int) bool {
		return fset.File(files[i].Pos()).Name() > fset.File(files[j].Pos()).Name()
	})

	archive, err := compiler.Compile(importPath, files, fset, importContext, minify)
	if err != nil {
		return nil, err
	}

	/*
		for _, jsFile := range pkg.JSFiles {
			fname := filepath.Join(pkg.Dir, jsFile)
			fs := s.Filesystem(fname)
			code, err := readFile(fs, fname)
			if err != nil {
				return nil, err
			}
			archive.IncJSCode = append(archive.IncJSCode, []byte("\t(function() {\n")...)
			archive.IncJSCode = append(archive.IncJSCode, code...)
			archive.IncJSCode = append(archive.IncJSCode, []byte("\n\t}).call($global);\n")...)
		}
	*/

	/*
		if s.options.Verbose {
			show := true
			if s.options.Standard != nil {
				if _, ok := s.options.Standard[importPath]; ok {
					show = false
				}
			}
			if show {
				fmt.Fprintln(s.options.Log, importPath)
			}
		}
	*/

	return archive, nil
}
