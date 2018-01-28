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

	"github.com/gopherjs/gopherjs/compiler"
	"github.com/gopherjs/gopherjs/compiler/natives"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
)

type ImportCError struct {
	pkgPath string
}

func (e *ImportCError) Error() string {
	return e.pkgPath + `: importing "C" is not supported by GopherJS`
}

func (s *Session) NewBuildContext(installSuffix string, buildTags []string) *build.Context {
	return &build.Context{
		GOROOT:        "goroot",
		GOPATH:        "gopath",
		GOOS:          build.Default.GOOS,
		GOARCH:        "js",
		InstallSuffix: installSuffix,
		Compiler:      "gc",
		BuildTags:     append(buildTags, "netgo"),
		ReleaseTags:   build.Default.ReleaseTags,
		CgoEnabled:    true, // detect `import "C"` to throw proper error

		// IsDir reports whether the path names a directory.
		// If IsDir is nil, Import calls os.Stat and uses the result's IsDir method.
		IsDir: func(path string) bool {
			fs := s.Filesystem(path)
			fi, err := fs.Stat(path)
			return err == nil && fi.IsDir()
		},

		// HasSubdir reports whether dir is lexically a subdirectory of
		// root, perhaps multiple levels below. It does not try to check
		// whether dir exists.
		// If so, HasSubdir sets rel to a slash-separated path that
		// can be joined to root to produce a path equivalent to dir.
		// If HasSubdir is nil, Import uses an implementation built on
		// filepath.EvalSymlinks.
		HasSubdir: func(root, dir string) (rel string, ok bool) {
			const sep = string(filepath.Separator)
			root = filepath.Clean(root)
			if !strings.HasSuffix(root, sep) {
				root += sep
			}
			dir = filepath.Clean(dir)
			if !strings.HasPrefix(dir, root) {
				return "", false
			}
			return filepath.ToSlash(dir[len(root):]), true
		},

		// ReadDir returns a slice of os.FileInfo, sorted by Name,
		// describing the content of the named directory.
		// If ReadDir is nil, Import uses ioutil.ReadDir.
		ReadDir: func(path string) ([]os.FileInfo, error) {
			fs := s.Filesystem(path)
			return fs.ReadDir(path)
		},

		// OpenFile opens a file (not a directory) for reading.
		// If OpenFile is nil, Import uses os.Open.
		OpenFile: func(path string) (io.ReadCloser, error) {
			fs := s.Filesystem(path)
			return fs.Open(path)
		},
	}
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
func (s *Session) Import(path string, mode build.ImportMode, installSuffix string, buildTags []string) (*PackageData, error) {
	return s.importWithSrcDir(path, "", mode, installSuffix, buildTags)
}

func (s *Session) importWithSrcDir(path string, srcDir string, mode build.ImportMode, installSuffix string, buildTags []string) (*PackageData, error) {
	bctx := s.NewBuildContext(installSuffix, buildTags)
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
	pkg, err := bctx.Import(path, srcDir, mode)
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

	jsFiles, err := s.jsFilesFromDir(pkg.Dir)
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
func (s *Session) ImportDir(dir string, mode build.ImportMode, installSuffix string, buildTags []string) (*PackageData, error) {
	pkg, err := s.NewBuildContext(installSuffix, buildTags).ImportDir(dir, mode)
	if err != nil {
		return nil, err
	}

	jsFiles, err := s.jsFilesFromDir(pkg.Dir)
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
func (s *Session) parseAndAugment(pkg *build.Package, isTest bool, fileSet *token.FileSet) ([]*ast.File, error) {
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
		GOOS:     build.Default.GOOS,
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
		fs := s.Filesystem(name)
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

	// Filesystem for GOROOT - should contain all Standard Library source
	Root billy.Filesystem

	// Filesystem for GOPATH (optional)
	Path billy.Filesystem

	// Filesystem for temporary Archive storage (optional)
	Temporary billy.Filesystem

	Unvendor       bool // Render JS with unvendored paths
	Initializer    bool // Render JS with deferred initialization
	Log            io.Writer
	Verbose        bool
	Quiet          bool
	CreateMapFile  bool
	MapToLocalDisk bool
	Minify         bool
	Color          bool
	BuildTags      []string
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

type PackageData struct {
	*build.Package
	JSFiles    []string
	IsTest     bool // IsTest is true if the package is being built for running tests.
	SrcModTime time.Time
	UpToDate   bool
}

type Session struct {
	options  *Options
	Archives map[string]*compiler.Archive
	Types    map[string]*types.Package
}

// Gets either the Gopath or Goroot filesystems depending on the path
func (s *Session) Filesystem(fpath string) billy.Filesystem {
	p := filepath.Clean(fpath)
	p = strings.TrimPrefix(p, string(filepath.Separator))
	parts := strings.Split(p, string(filepath.Separator))
	dir := parts[0]
	switch dir {
	case "gopath":
		return s.options.Path
	case "goroot":
		return s.options.Root
	}
	panic(fmt.Sprintf("Top dir should be goroot or gopath, got %s", fpath))
}

func NewSession(options *Options) *Session {
	if options.Root == nil {
		panic("Need to specify Goroot in options (should contain full standard library source)")
	}
	if options.Path == nil {
		options.Path = memfs.New()
	}
	if options.Temporary == nil {
		options.Temporary = memfs.New()
	}

	s := &Session{
		options:  options,
		Archives: make(map[string]*compiler.Archive),
	}
	s.Types = make(map[string]*types.Package)
	return s
}

func (s *Session) InstallSuffix() string {
	if s.options.Minify {
		return "min"
	}
	return ""
}

func (s *Session) BuildDir(packagePath string, importPath string, pkgObj string, standard StandardFunc) (*CommandOutput, error) {
	buildPkg, err := s.NewBuildContext(s.InstallSuffix(), s.options.BuildTags).ImportDir(packagePath, 0)
	if err != nil {
		return nil, err
	}
	pkg := &PackageData{Package: buildPkg}
	jsFiles, err := s.jsFilesFromDir(pkg.Dir)
	if err != nil {
		return nil, err
	}
	pkg.JSFiles = jsFiles
	archive, err := s.BuildPackage(pkg)
	if err != nil {
		return nil, err
	}
	if !pkg.IsCommand() {
		return nil, nil
	}
	cp, err := s.WriteCommandPackage(archive, standard)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func (s *Session) BuildFiles(filenames []string, pkgObj string, packagePath string, standard StandardFunc) (*CommandOutput, error) {
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

	archive, err := s.BuildPackage(pkg)
	if err != nil {
		return nil, err
	}
	if s.Types["main"].Name() != "main" {
		return nil, fmt.Errorf("cannot build/run non-main package")
	}
	return s.WriteCommandPackage(archive, standard)
}

func (s *Session) BuildImportPath(path string) (*compiler.Archive, error) {
	_, archive, err := s.buildImportPathWithSrcDir(path, "")
	return archive, err
}

func (s *Session) buildImportPathWithSrcDir(path string, srcDir string) (*PackageData, *compiler.Archive, error) {
	pkg, err := s.importWithSrcDir(path, srcDir, 0, s.InstallSuffix(), s.options.BuildTags)
	if err != nil {
		return nil, nil, err
	}

	archive, err := s.BuildPackage(pkg)
	if err != nil {
		return nil, nil, err
	}

	return pkg, archive, nil
}

func (s *Session) BuildPackage(pkg *PackageData) (*compiler.Archive, error) {
	if archive, ok := s.Archives[pkg.ImportPath]; ok {
		return archive, nil
	}

	if pkg.PkgObj != "" {

		// TODO: Is this needed? I don't think so?
		/*
			for _, importedPkgPath := range pkg.Imports {
				// Ignore all imports that aren't mentioned in import specs of pkg.
				// For example, this ignores imports such as runtime/internal/sys and runtime/internal/atomic.
				ignored := true
				for _, pos := range pkg.ImportPos[importedPkgPath] {
					importFile := filepath.Base(pos.Filename)
					for _, file := range pkg.GoFiles {
						if importFile == file {
							ignored = false
							break
						}
					}
					if !ignored {
						break
					}
				}

				if importedPkgPath == "unsafe" || ignored {
					continue
				}
				if _, _, err := s.buildImportPathWithSrcDir(importedPkgPath, pkg.Dir); err != nil {
					return nil, err
				}
			}
		*/

		if _, err := s.options.Temporary.Stat(pkg.PkgObj); err == nil {
			// package object exists, load from disk
			pkg.UpToDate = true

			archive, err := readArchive(s.options.Temporary, pkg.PkgObj, pkg.ImportPath, s.Types)
			if err != nil {
				return nil, err
			}

			s.Archives[pkg.ImportPath] = archive
			return archive, err
		}
	}

	fileSet := token.NewFileSet()
	files, err := s.parseAndAugment(pkg.Package, pkg.IsTest, fileSet)
	if err != nil {
		return nil, err
	}

	localImportPathCache := make(map[string]*compiler.Archive)
	importContext := &compiler.ImportContext{
		Packages: s.Types,
		Import: func(path string) (*compiler.Archive, error) {
			if archive, ok := localImportPathCache[path]; ok {
				return archive, nil
			}
			_, archive, err := s.buildImportPathWithSrcDir(path, pkg.Dir)
			if err != nil {
				return nil, err
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

	importPath := pkg.ImportPath
	if s.options.Unvendor {
		importPath = UnvendorPath(pkg.ImportPath)
	}

	archive, err := compiler.Compile(importPath, files, fileSet, importContext, s.options.Minify)
	if err != nil {
		return nil, err
	}

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

	if s.options.Verbose {
		fmt.Fprintln(s.options.Log, pkg.ImportPath)
	}

	s.Archives[pkg.ImportPath] = archive

	// TODO: Why would PkgObj be empty?
	if pkg.PkgObj == "" {
		return archive, nil
	}

	if err := s.writeLibraryPackage(archive, pkg.PkgObj); err != nil {
		return nil, err
	}

	return archive, nil
}

func readArchive(fs billy.Filesystem, fpath, path string, types map[string]*types.Package) (*compiler.Archive, error) {
	objFile, err := fs.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer objFile.Close()

	archive, err := compiler.ReadArchive(fpath, path, objFile, types)
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

func (s *Session) writeLibraryPackage(archive *compiler.Archive, pkgObj string) error {
	if err := s.options.Temporary.MkdirAll(filepath.Dir(pkgObj), 0777); err != nil {
		return err
	}

	objFile, err := s.options.Temporary.Create(pkgObj)
	if err != nil {
		return err
	}
	defer objFile.Close()

	return compiler.WriteArchive(archive, objFile)
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
}

func (s *Session) WriteCommandPackage(archive *compiler.Archive, standard StandardFunc) (*CommandOutput, error) {
	deps, err := compiler.ImportDependencies(archive, func(path string) (*compiler.Archive, error) {
		if archive, ok := s.Archives[path]; ok {
			return archive, nil
		}
		_, archive, err := s.buildImportPathWithSrcDir(path, "")
		return archive, err
	})
	if err != nil {
		return nil, err
	}

	commandPath, packages, err := GetProgramCode(deps, s.options.Initializer, standard)
	if err != nil {
		return nil, err
	}

	c := &CommandOutput{
		Path:     commandPath,
		Packages: packages,
	}
	return c, nil
}

type StandardFunc func(path string, min bool) (hash []byte, ok bool)

func GetProgramCode(pkgs []*compiler.Archive, initializer bool, standard StandardFunc) (string, []*PackageOutput, error) {

	mainPkg := pkgs[len(pkgs)-1]
	minify := mainPkg.Minified

	// write packages
	var packageOutputs []*PackageOutput
	for _, pkg := range pkgs {
		if standard != nil {
			if hash, ok := standard(pkg.ImportPath, minify); ok {
				packageOutputs = append(packageOutputs, &PackageOutput{
					Path:     pkg.ImportPath,
					Hash:     hash,
					Standard: true,
				})
				continue
			}
		}
		contents, hash, err := GetPackageCode(pkg, minify, initializer)
		if err != nil {
			return "", nil, err
		}
		packageOutputs = append(packageOutputs, &PackageOutput{
			Path:     pkg.ImportPath,
			Contents: contents,
			Hash:     hash,
		})
	}
	return mainPkg.ImportPath, packageOutputs, nil
}

func GetPackageCode(archive *compiler.Archive, minify, initializer bool) (contents []byte, hash []byte, err error) {
	dceSelection := make(map[*compiler.Decl]struct{})
	for _, d := range archive.Declarations {
		dceSelection[d] = struct{}{}
	}
	buf := new(bytes.Buffer)

	if initializer {
		if _, err := fmt.Fprintf(buf, `$initialisers["%s"] = function () {`, archive.ImportPath); err != nil {
			return nil, nil, err
		}
	}

	if err := compiler.WritePkgCode(archive, dceSelection, minify, &compiler.SourceMapFilter{Writer: buf}); err != nil {
		return nil, nil, err
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

func (s *Session) jsFilesFromDir(dir string) ([]string, error) {
	fs := s.Filesystem(dir)
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
