package builder

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"bytes"

	"crypto/sha1"

	"sort"

	"encoding/hex"

	"context"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/builder/session"
	"github.com/gopherjs/gopherjs/compiler"
	"github.com/gopherjs/gopherjs/compiler/natives"
	"golang.org/x/tools/go/gcexportdata"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

type ImportCError struct {
	pkgPath string
}

func (e *ImportCError) Error() string {
	return e.pkgPath + `: importing "C" is not supported by GopherJS`
}

// Import returns details about the Go package named by the import path. If the
// path is a local import path naming a package that can be imported using
// a standard import path, the returned package will set p.ImportPath to
// that path.
//
// In the directory containing the package, .go and .inc.js files are
// considered part of the package except for:
//
//    - .go files in package documentation
//    - files starting with _ or . (likely editor temporary files)
//    - files with build constraints not satisfied by the context
//
// If an error occurs, Import returns a non-nil error and a nil
// *PackageData.
func (b *Builder) Import(ctx context.Context, path string, mode build.ImportMode, installSuffix string) (*PackageData, error) {
	return b.importWithSrcDir(ctx, path, "", mode, installSuffix)
}

func (b *Builder) importWithSrcDir(ctx context.Context, path string, srcDir string, mode build.ImportMode, installSuffix string) (*PackageData, error) {
	bctx := b.BuildContext(true, installSuffix)
	switch path {
	case "syscall":
		// syscall needs to use a typical GOARCH like amd64 to pick up definitions for _Socklen, BpfInsn, IFNAMSIZ, Timeval, BpfStat, SYS_FCNTL, Flock_t, etc.
		bctx.GOARCH = runtime.GOARCH
		bctx.InstallSuffix = "js"
		if installSuffix != "" {
			bctx.InstallSuffix += "_" + installSuffix
		}
	case "math/big":
		// Use pure Go version of math/big; we don't want non-Go assembly versions.
		bctx.BuildTags = append(bctx.BuildTags, "math_big_pure_go")
	case "crypto/x509", "os/user":
		// These stdlib packages have cgo and non-cgo versions (via build tags); we want the latter.
		bctx.CgoEnabled = false
	}

	var pkg *build.Package
	var err error
	if WithCancel(ctx, func() {
		pkg, err = bctx.Import(path, srcDir, mode)
	}) {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}

	// TODO: Resolve issue #415 and remove this temporary workaround.
	if strings.HasSuffix(pkg.ImportPath, "/vendor/github.com/gopherjs/gopherjs/js") {
		return nil, fmt.Errorf("vendoring github.com/gopherjs/gopherjs/js package is not supported, see https://github.com/gopherjs/gopherjs/issues/415")
	}

	switch path {
	case "os":
		pkg.GoFiles = excludeExecutable(pkg.GoFiles) // Need to exclude executable implementation files, because some of them contain package scope variables that perform (indirectly) syscalls on init.
	case "runtime":
		pkg.GoFiles = []string{"error.go"}
	case "runtime/internal/sys":
		pkg.GoFiles = []string{fmt.Sprintf("zgoos_%s.go", bctx.GOOS), "zversion.go"}
	case "runtime/pprof":
		pkg.GoFiles = nil
	case "internal/poll":
		pkg.GoFiles = exclude(pkg.GoFiles, "fd_poll_runtime.go")
	case "crypto/rand":
		pkg.GoFiles = []string{"rand.go", "util.go"}
		pkg.TestGoFiles = exclude(pkg.TestGoFiles, "rand_linux_test.go") // Don't want linux-specific tests (since linux-specific package files are excluded too).
	}

	if len(pkg.CgoFiles) > 0 {
		return nil, &ImportCError{path}
	}

	// TODO: Is this needed?
	/*
		if _, err := os.Stat(pkg.PkgObj); os.IsNotExist(err) && strings.HasPrefix(pkg.PkgObj, build.Default.GOROOT) {
			// fall back to GOPATH
			firstGopathWorkspace := filepath.SplitList(build.Default.GOPATH)[0] // TODO: Need to check inside all GOPATH workspaces.
			gopathPkgObj := filepath.Join(firstGopathWorkspace, pkg.PkgObj[len(build.Default.GOROOT):])
			if _, err := os.Stat(gopathPkgObj); err == nil {
				pkg.PkgObj = gopathPkgObj
			}
		}
	*/

	jsFiles, err := b.jsFilesFromDir(pkg.Dir)
	if err != nil {
		return nil, err
	}

	return &PackageData{Package: pkg, JSFiles: jsFiles}, nil
}

// excludeExecutable excludes all executable implementation .go files.
// They have "executable_" prefix.
func excludeExecutable(goFiles []string) []string {
	var s []string
	for _, f := range goFiles {
		if strings.HasPrefix(f, "executable_") {
			continue
		}
		s = append(s, f)
	}
	return s
}

// exclude returns files, excluding specified files.
func exclude(files []string, exclude ...string) []string {
	var s []string
Outer:
	for _, f := range files {
		for _, e := range exclude {
			if f == e {
				continue Outer
			}
		}
		s = append(s, f)
	}
	return s
}

// ImportDir is like Import but processes the Go package found in the named
// directory.
func (b *Builder) ImportDir(ctx context.Context, dir string, mode build.ImportMode, installSuffix string) (*PackageData, error) {

	var pkg *build.Package
	var err error
	if WithCancel(ctx, func() {
		pkg, err = b.BuildContext(true, installSuffix).ImportDir(dir, mode)
	}) {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}

	jsFiles, err := b.jsFilesFromDir(pkg.Dir)
	if err != nil {
		return nil, err
	}

	return &PackageData{Package: pkg, JSFiles: jsFiles}, nil
}

// parseAndAugment parses and returns all .go files of given pkg.
// Standard Go library packages are augmented with files in compiler/natives folder.
// If isTest is true and pkg.ImportPath has no _test suffix, package is built for running internal tests.
// If isTest is true and pkg.ImportPath has _test suffix, package is built for running external tests.
//
// The native packages are augmented by the contents of natives.FS in the following way.
// The file names do not matter except the usual `_test` suffix. The files for
// native overrides get added to the package (even if they have the same name
// as an existing file from the standard library). For all identifiers that exist
// in the original AND the overrides, the original identifier in the AST gets
// replaced by `_`. New identifiers that don't exist in original package get added.
func (b *Builder) parseAndAugment(pkg *build.Package, isTest bool, fileSet *token.FileSet) ([]*ast.File, error) {
	var files []*ast.File
	replacedDeclNames := make(map[string]bool)
	funcName := func(d *ast.FuncDecl) string {
		if d.Recv == nil || len(d.Recv.List) == 0 {
			return d.Name.Name
		}
		recv := d.Recv.List[0].Type
		if star, ok := recv.(*ast.StarExpr); ok {
			recv = star.X
		}
		return recv.(*ast.Ident).Name + "." + d.Name.Name
	}
	isXTest := strings.HasSuffix(pkg.ImportPath, "_test")
	importPath := pkg.ImportPath
	if isXTest {
		importPath = importPath[:len(importPath)-5]
	}

	nativesContext := &build.Context{
		GOROOT:   "/",
		GOOS:     "darwin",
		GOARCH:   "js",
		Compiler: "gc",
		JoinPath: path.Join,
		SplitPathList: func(list string) []string {
			if list == "" {
				return nil
			}
			return strings.Split(list, "/")
		},
		IsAbsPath: path.IsAbs,
		IsDir: func(name string) bool {
			dir, err := natives.FS.Open(name)
			if err != nil {
				return false
			}
			defer dir.Close()
			info, err := dir.Stat()
			if err != nil {
				return false
			}
			return info.IsDir()
		},
		HasSubdir: func(root, name string) (rel string, ok bool) {
			panic("not implemented")
		},
		ReadDir: func(name string) (fi []os.FileInfo, err error) {
			dir, err := natives.FS.Open(name)
			if err != nil {
				return nil, err
			}
			defer dir.Close()
			return dir.Readdir(0)
		},
		OpenFile: func(name string) (r io.ReadCloser, err error) {
			return natives.FS.Open(name)
		},
	}
	if nativesPkg, err := nativesContext.Import(importPath, "", 0); err == nil {
		names := nativesPkg.GoFiles
		if isTest {
			names = append(names, nativesPkg.TestGoFiles...)
		}
		if isXTest {
			names = nativesPkg.XTestGoFiles
		}
		for _, name := range names {
			fullPath := path.Join(nativesPkg.Dir, name)
			r, err := nativesContext.OpenFile(fullPath)
			if err != nil {
				panic(err)
			}
			newPath := path.Join(pkg.Dir, "__"+name)
			file, err := parser.ParseFile(fileSet, newPath, r, parser.ParseComments)
			if err != nil {
				panic(err)
			}
			r.Close()
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					replacedDeclNames[funcName(d)] = true
				case *ast.GenDecl:
					switch d.Tok {
					case token.TYPE:
						for _, spec := range d.Specs {
							replacedDeclNames[spec.(*ast.TypeSpec).Name.Name] = true
						}
					case token.VAR, token.CONST:
						for _, spec := range d.Specs {
							for _, name := range spec.(*ast.ValueSpec).Names {
								replacedDeclNames[name.Name] = true
							}
						}
					}
				}
			}
			files = append(files, file)
		}
	}
	delete(replacedDeclNames, "init")

	var errList compiler.ErrorList
	for _, name := range pkg.GoFiles {
		if !filepath.IsAbs(name) {
			name = filepath.Join(pkg.Dir, name)
		}
		fdir, _ := filepath.Split(name)
		fs := b.Filesystem(fdir)
		r, err := fs.Open(name)
		if err != nil {
			return nil, err
		}
		file, err := parser.ParseFile(fileSet, name, r, parser.ParseComments)
		r.Close()
		if err != nil {
			if list, isList := err.(scanner.ErrorList); isList {
				if len(list) > 10 {
					list = append(list[:10], &scanner.Error{Pos: list[9].Pos, Msg: "too many errors"})
				}
				for _, entry := range list {
					errList = append(errList, entry)
				}
				continue
			}
			errList = append(errList, err)
			continue
		}

		switch pkg.ImportPath {
		case "crypto/rand", "encoding/gob", "encoding/json", "expvar", "go/token", "log", "math/big", "math/rand", "regexp", "testing", "time":
			for _, spec := range file.Imports {
				path, _ := strconv.Unquote(spec.Path.Value)
				if path == "sync" {
					if spec.Name == nil {
						spec.Name = ast.NewIdent("sync")
					}
					spec.Path.Value = `"github.com/gopherjs/gopherjs/nosync"`
				}
			}
		}

		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if replacedDeclNames[funcName(d)] {
					d.Name = ast.NewIdent("_")
				}
			case *ast.GenDecl:
				switch d.Tok {
				case token.TYPE:
					for _, spec := range d.Specs {
						s := spec.(*ast.TypeSpec)
						if replacedDeclNames[s.Name.Name] {
							s.Name = ast.NewIdent("_")
						}
					}
				case token.VAR, token.CONST:
					for _, spec := range d.Specs {
						s := spec.(*ast.ValueSpec)
						for i, name := range s.Names {
							if replacedDeclNames[name.Name] {
								s.Names[i] = ast.NewIdent("_")
							}
						}
					}
				}
			}
		}
		files = append(files, file)
	}
	if errList != nil {
		return nil, errList
	}
	return files, nil
}

type Options struct {
	Temporary      billy.Filesystem // Filesystem for temporary Archive storage (optional)
	Unvendor       bool             // Render JS with unvendored paths
	Initializer    bool             // Render JS with deferred initialization
	Log            io.Writer
	Verbose        bool
	Quiet          bool
	CreateMapFile  bool
	MapToLocalDisk bool
	Minify         bool
	Color          bool
	Standard       map[string]map[bool]string
}

func (o *Options) PrintError(format string, a ...interface{}) {
	if o.Color {
		format = "\x1B[31m" + format + "\x1B[39m"
	}
	fmt.Fprintf(os.Stderr, format, a...)
}

func (o *Options) PrintSuccess(format string, a ...interface{}) {
	if o.Color {
		format = "\x1B[32m" + format + "\x1B[39m"
	}
	fmt.Fprintf(os.Stderr, format, a...)
}

func Bytes(in string) []byte {
	if in == "" {
		return nil
	}
	b, err := hex.DecodeString(in)
	if err != nil {
		panic(fmt.Sprintf("invalid hex: %s", in))
	}
	return b
}

type PackageData struct {
	*build.Package
	JSFiles    []string
	IsTest     bool // IsTest is true if the package is being built for running tests.
	SrcModTime time.Time
	UpToDate   bool
}

type Builder struct {
	*session.Session
	options  *Options
	Archives map[string]*compiler.Archive
	Types    map[string]*types.Package
	Callback func(*compiler.Archive) error
}

func New(session *session.Session, options *Options) *Builder {
	if options.Temporary == nil {
		options.Temporary = memfs.New()
	}
	s := &Builder{
		Session:  session,
		options:  options,
		Archives: make(map[string]*compiler.Archive),
	}
	s.Types = make(map[string]*types.Package)
	return s
}

func (b *Builder) InstallSuffix() string {
	if b.options.Minify {
		return "min"
	}
	return ""
}

func (b *Builder) BuildDir(ctx context.Context, packagePath string, importPath string, pkgObj string) (*CommandOutput, error) {

	var buildPkg *build.Package
	var err error
	if WithCancel(ctx, func() {
		buildPkg, err = b.BuildContext(true, b.InstallSuffix()).ImportDir(packagePath, 0)
	}) {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}
	pkg := &PackageData{Package: buildPkg}
	jsFiles, err := b.jsFilesFromDir(pkg.Dir)
	if err != nil {
		return nil, err
	}
	pkg.JSFiles = jsFiles
	archive, err := b.BuildPackage(ctx, pkg)
	if err != nil {
		return nil, err
	}
	if !pkg.IsCommand() {
		return nil, nil
	}
	cp, err := b.WriteCommandPackage(ctx, archive)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func (b *Builder) BuildFiles(ctx context.Context, filenames []string, pkgObj string, packagePath string) (*CommandOutput, error) {
	pkg := &PackageData{
		Package: &build.Package{
			Name:       "main",
			ImportPath: "main",
			Dir:        packagePath,
		},
	}

	for _, file := range filenames {
		if strings.HasSuffix(file, ".inc.js") {
			pkg.JSFiles = append(pkg.JSFiles, file)
			continue
		}
		pkg.GoFiles = append(pkg.GoFiles, file)
	}

	archive, err := b.BuildPackage(ctx, pkg)
	if err != nil {
		return nil, err
	}
	if b.Types["main"].Name() != "main" {
		return nil, fmt.Errorf("cannot build/run non-main package")
	}
	return b.WriteCommandPackage(ctx, archive)
}

func (b *Builder) BuildImportPath(ctx context.Context, path string) (*PackageData, *compiler.Archive, error) {
	return b.buildImportPathWithSrcDir(ctx, path, "")
}

func (b *Builder) buildImportPathWithSrcDir(ctx context.Context, path string, srcDir string) (*PackageData, *compiler.Archive, error) {
	pkg, err := b.importWithSrcDir(ctx, path, srcDir, 0, b.InstallSuffix())
	if err != nil {
		return nil, nil, err
	}

	archive, err := b.BuildPackage(ctx, pkg)
	if err != nil {
		return nil, nil, err
	}

	return pkg, archive, nil
}

func (b *Builder) ImportStandardArchive(ctx context.Context, importPath string) (*compiler.Archive, error) {

	if assets.Archives == nil {
		// assets.Archives may be nil if we don't initialise the assets (for bootstrapping we need to
		// run "go generate" without existing assets).
		return nil, nil
	}

	archivePair, ok := assets.Archives[importPath]
	if !ok {
		return nil, nil
	}

	archive := archivePair[b.options.Minify]
	p, err := gcexportdata.Read(bytes.NewReader(archive.ExportData), token.NewFileSet(), b.Types, importPath)
	if err != nil {
		return nil, err
	}
	b.Types[importPath] = p
	b.Archives[importPath] = archive

	if b.Callback != nil {
		if err := b.Callback(archive); err != nil {
			return nil, err
		}
	}

	for _, p := range archive.Imports {
		if b.Archives[p] != nil {
			continue
		}
		if _, err := b.ImportStandardArchive(ctx, p); err != nil {
			return nil, err
		}
	}

	return archive, nil

}

func (b *Builder) BuildPackage(ctx context.Context, pkg *PackageData) (*compiler.Archive, error) {

	importPath := pkg.ImportPath
	if b.options.Unvendor {
		importPath = UnvendorPath(pkg.ImportPath)
	}

	if archive, ok := b.Archives[importPath]; ok {
		return archive, nil
	}

	// If the path is not in the source collection, and the archive exists in the std lib precompiled
	// archives, load it...
	if !b.HasSource(importPath) {
		archive, err := b.ImportStandardArchive(ctx, importPath)
		if err != nil {
			return nil, err
		}
		if archive != nil {
			return archive, nil
		}
	}

	fileSet := token.NewFileSet()
	files, err := b.parseAndAugment(pkg.Package, pkg.IsTest, fileSet)
	if err != nil {
		return nil, err
	}

	localImportPathCache := make(map[string]*compiler.Archive)
	importContext := &compiler.ImportContext{
		Packages: b.Types,
		Import: func(path string) (*compiler.Archive, error) {
			if archive, ok := localImportPathCache[path]; ok {
				return archive, nil
			}
			_, archive, err := b.buildImportPathWithSrcDir(ctx, path, pkg.Dir)
			if err != nil {
				return nil, pathErr{error: err, path: path}
			}
			localImportPathCache[path] = archive
			return archive, nil
		},
	}

	// TODO: Remove this when https://github.com/gopherjs/gopherjs/pull/742 is merged
	// Files must be in the same order to get reproducible JS
	sort.Slice(files, func(i, j int) bool {
		return fileSet.File(files[i].Pos()).Name() > fileSet.File(files[j].Pos()).Name()
	})

	var archive *compiler.Archive
	if WithCancel(ctx, func() {
		archive, err = compiler.Compile(importPath, files, fileSet, importContext, b.options.Minify)
	}) {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}

	for _, jsFile := range pkg.JSFiles {
		fname := filepath.Join(pkg.Dir, jsFile)
		fs := b.Filesystem(pkg.Dir)
		code, err := readFile(fs, fname)
		if err != nil {
			return nil, err
		}
		archive.IncJSCode = append(archive.IncJSCode, []byte("\t(function() {\n")...)
		archive.IncJSCode = append(archive.IncJSCode, code...)
		archive.IncJSCode = append(archive.IncJSCode, []byte("\n\t}).call($global);\n")...)
	}

	if b.options.Verbose {
		show := true
		if b.options.Standard != nil {
			if _, ok := b.options.Standard[importPath]; ok {
				show = false
			}
		}
		if show {
			fmt.Fprintln(b.options.Log, importPath)
		}
	}

	b.Archives[importPath] = archive

	if b.Callback != nil {
		if err := b.Callback(archive); err != nil {
			return nil, err
		}
	}

	// TODO: Why would PkgObj be empty?
	if pkg.PkgObj == "" {
		return archive, nil
	}

	if err := b.writeLibraryPackage(ctx, archive, pkg.PkgObj); err != nil {
		return nil, err
	}

	return archive, nil
}

type pathErr struct {
	error
	path string
}

func (p pathErr) Path() string {
	return p.path
}

func readArchive(ctx context.Context, fs billy.Filesystem, fpath, path string, types map[string]*types.Package) (*compiler.Archive, error) {
	objFile, err := fs.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer objFile.Close()

	var archive *compiler.Archive
	if WithCancel(ctx, func() {
		archive, err = compiler.ReadArchive(fpath, path, objFile, types)
	}) {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}
	return archive, nil
}

func readFile(fs billy.Filesystem, path string) ([]byte, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *Builder) writeLibraryPackage(ctx context.Context, archive *compiler.Archive, pkgObj string) error {
	if err := b.options.Temporary.MkdirAll(filepath.Dir(pkgObj), 0777); err != nil {
		return err
	}

	objFile, err := b.options.Temporary.Create(pkgObj)
	if err != nil {
		return err
	}
	defer objFile.Close()

	if WithCancel(ctx, func() {
		err = compiler.WriteArchive(archive, objFile)
	}) {
		return ctx.Err()
	}
	return err
}

type CommandOutput struct {
	Path     string
	Packages []*PackageOutput
}

type PackageOutput struct {
	Path     string
	Hash     []byte
	Contents []byte
	Standard bool
	Store    bool
}

func (b *Builder) GetDependencies(ctx context.Context, archive *compiler.Archive) ([]*compiler.Archive, error) {

	importer := func(path string) (*compiler.Archive, error) {
		if archive, ok := b.Archives[path]; ok {
			return archive, nil
		}
		_, archive, err := b.buildImportPathWithSrcDir(ctx, path, "")
		return archive, err
	}

	var deps []*compiler.Archive
	var err error
	if WithCancel(ctx, func() {
		deps, err = compiler.ImportDependencies(archive, importer)
	}) {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}
	return deps, nil
}

func (b *Builder) WriteCommandPackage(ctx context.Context, archive *compiler.Archive) (*CommandOutput, error) {

	deps, err := b.GetDependencies(ctx, archive)
	if err != nil {
		return nil, err
	}

	commandPath, packages, err := b.GetProgramCode(ctx, deps)
	if err != nil {
		return nil, err
	}

	c := &CommandOutput{
		Path:     commandPath,
		Packages: packages,
	}
	return c, nil
}

func (b *Builder) GetProgramCode(ctx context.Context, pkgs []*compiler.Archive) (string, []*PackageOutput, error) {

	mainPkg := pkgs[len(pkgs)-1]
	minify := mainPkg.Minified

	// write packages
	var packageOutputs []*PackageOutput
	for _, pkg := range pkgs {

		// look the path up in the list of pre-stored standard library packages, and use that instead of
		// generating the package code... But only if the package doesn't exist in the source collection.
		var std bool
		var ph map[bool]string
		ph, std = b.options.Standard[pkg.ImportPath]

		if std && !b.HasSource(pkg.ImportPath) {
			packageOutputs = append(packageOutputs, &PackageOutput{
				Path:     pkg.ImportPath,
				Hash:     Bytes(ph[minify]),
				Contents: nil,
				Standard: true,
				Store:    false,
			})
			continue
		}
		contents, hash, err := GetPackageCode(ctx, pkg, minify, b.options.Initializer)
		if err != nil {
			return "", nil, err
		}
		packageOutputs = append(packageOutputs, &PackageOutput{
			Path:     pkg.ImportPath,
			Hash:     hash,
			Contents: contents,
			Standard: std,
			Store:    true,
		})
	}
	return mainPkg.ImportPath, packageOutputs, nil
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

func (b *Builder) jsFilesFromDir(dir string) ([]string, error) {
	fs := b.Filesystem(dir)
	files, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var jsFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".inc.js") && file.Name()[0] != '_' && file.Name()[0] != '.' {
			jsFiles = append(jsFiles, file.Name())
		}
	}
	return jsFiles, nil
}

// hasGopathPrefix returns true and the length of the matched GOPATH workspace,
// iff file has a prefix that matches one of the GOPATH workspaces.
func hasGopathPrefix(file, gopath string) (hasGopathPrefix bool, prefixLen int) {
	gopathWorkspaces := filepath.SplitList(gopath)
	for _, gopathWorkspace := range gopathWorkspaces {
		gopathWorkspace = filepath.Clean(gopathWorkspace)
		if strings.HasPrefix(file, gopathWorkspace) {
			return true, len(gopathWorkspace)
		}
	}
	return false, 0
}

func UnvendorPath(path string) string {
	i, ok := findVendor(path)
	if !ok {
		return path
	}
	return path[i+len("vendor/"):]
}

// FindVendor looks for the last non-terminating "vendor" path element in the given import path.
// If there isn't one, FindVendor returns ok=false.
// Otherwise, FindVendor returns ok=true and the index of the "vendor".
// Copied from cmd/go/internal/load
func findVendor(path string) (index int, ok bool) {
	// Two cases, depending on internal at start of string or not.
	// The order matters: we must return the index of the final element,
	// because the final one is where the effective import path starts.
	switch {
	case strings.Contains(path, "/vendor/"):
		return strings.LastIndex(path, "/vendor/") + 1, true
	case strings.HasPrefix(path, "vendor/"):
		return 0, true
	}
	return 0, false
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
