package builderjs

import (
	"context"
	"go/token"
	"go/types"
	"sort"

	"go/ast"
	"go/parser"

	"fmt"

	"strings"

	"bytes"
	"crypto/sha1"

	"github.com/gopherjs/gopherjs/compiler"
	"golang.org/x/tools/go/gcimporter15"
)

func BuildPackage(path string, source map[string]map[string]string, deps []*compiler.Archive, minify bool, archives map[string]*compiler.Archive, packages map[string]*types.Package) (*compiler.Archive, error) {

	for _, a := range deps {
		if archives[a.ImportPath] == nil {
			archives[a.ImportPath] = a
		}
		if packages[a.ImportPath] == nil {
			_, p, err := gcimporter.BImportData(token.NewFileSet(), packages, a.ExportData, a.ImportPath)
			if err != nil {
				return nil, err
			}
			packages[a.ImportPath] = p
		}
	}

	fset := token.NewFileSet()

	var importContext *compiler.ImportContext
	importContext = &compiler.ImportContext{
		Packages: packages,
		Import: func(imp string) (*compiler.Archive, error) {
			a, ok := archives[imp]
			if ok {
				return a, nil
			}
			sourceFiles, ok := source[imp]
			if ok {
				// We have the source for this dep
				archive, err := compileFiles(fset, imp, sourceFiles, importContext, minify)
				if err != nil {
					return nil, err
				}
				return archive, nil
			}
			return nil, fmt.Errorf("%s not found", imp)
		},
	}

	archive, err := compileFiles(fset, path, source[path], importContext, minify)
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

func compileFiles(fset *token.FileSet, path string, sourceFiles map[string]string, importContext *compiler.ImportContext, minify bool) (*compiler.Archive, error) {
	var files []*ast.File
	for name, contents := range sourceFiles {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, contents, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	// TODO: Remove this when https://github.com/gopherjs/gopherjs/pull/742 is merged
	// Files must be in the same order to get reproducible JS
	sort.Slice(files, func(i, j int) bool {
		return fset.File(files[i].Pos()).Name() > fset.File(files[j].Pos()).Name()
	})

	archive, err := compiler.Compile(path, files, fset, importContext, minify)
	if err != nil {
		return nil, err
	}
	return archive, nil
}

func GetPackageCode(ctx context.Context, archive *compiler.Archive, minify, initializer bool) (contents []byte, hash []byte, err error) {
	dceSelection := make(map[*compiler.Decl]struct{})
	for _, d := range archive.Declarations {
		dceSelection[d] = struct{}{}
	}
	buf := new(bytes.Buffer)

	if initializer {
		var s string
		if minify {
			s = `$load["%s"]=function(){`
		} else {
			s = `$load["%s"] = function () {` + "\n"
		}
		if _, err := fmt.Fprintf(buf, s, archive.ImportPath); err != nil {
			return nil, nil, err
		}
	}
	if WithCancel(ctx, func() {
		err = compiler.WritePkgCode(archive, dceSelection, minify, &compiler.SourceMapFilter{Writer: buf})
	}) {
		return nil, nil, ctx.Err()
	}
	if err != nil {
		return nil, nil, err
	}

	if minify {
		// compiler.WritePkgCode always finishes with a "\n". In minified mode we should remove this.
		buf.Truncate(buf.Len() - 1)
	}

	if initializer {
		if _, err := fmt.Fprint(buf, "};"); err != nil {
			return nil, nil, err
		}
	}

	sha := sha1.New()
	if _, err := sha.Write(buf.Bytes()); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), sha.Sum(nil), nil
}

// WithCancel executes the provided function, but returns early with true if the context cancellation
// signal was recieved.
func WithCancel(ctx context.Context, f func()) bool {

	finished := make(chan struct{})
	go func() {
		f()
		close(finished)
	}()
	select {
	case <-finished:
		return false
	case <-ctx.Done():
		return true
	}
}
