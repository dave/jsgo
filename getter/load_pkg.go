package getter

import (
	"fmt"
	"go/build"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dave/jsgo/assets"

	"context"
)

// A Package describes a single package found in a directory.
type Package struct {
	PackagePublic                 // visible in 'go list'
	Internal      PackageInternal // for use inside go command only
}

type PackagePublic struct {
	cache *Session
	// Note: These fields are part of the go command's public API.
	// See list.go. It is okay to add fields, but not to change or
	// remove existing ones. Keep in sync with list.go
	Dir           string `json:",omitempty"` // directory containing package sources
	ImportPath    string `json:",omitempty"` // import path of package in dir
	ImportComment string `json:",omitempty"` // path in import comment on package statement
	Name          string `json:",omitempty"` // package name
	Doc           string `json:",omitempty"` // package documentation string
	Goroot        bool   `json:",omitempty"` // is this package found in the Go root?
	Standard      bool   `json:",omitempty"` // is this package part of the standard Go library?
	Root          string `json:",omitempty"` // Go root or Go path dir containing this package

	// Source files
	// If you add to this list you MUST add to p.AllFiles (below) too.
	// Otherwise file name security lists will not apply to any new additions.
	GoFiles        []string `json:",omitempty"` // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	IgnoredGoFiles []string `json:",omitempty"` // .go sources ignored due to build constraints

	// Dependency information
	Imports []string `json:",omitempty"` // import paths used by this package
	Deps    []string `json:",omitempty"` // all (recursively) imported dependencies

	// Error information
	Incomplete bool            `json:",omitempty"` // was there an error loading this package or dependencies?
	Error      *PackageError   `json:",omitempty"` // error loading this package (not dependencies)
	DepsErrors []*PackageError `json:",omitempty"` // errors loading dependencies
}

// AllFiles returns the names of all the files considered for the package.
// This is used for sanity and security checks, so we include all files,
// even IgnoredGoFiles, because some subcommands consider them.
// The go/build package filtered others out (like foo_wrongGOARCH.s)
// and that's OK.
func (p *Package) AllFiles() []string {
	return StringList(
		p.GoFiles,
		p.IgnoredGoFiles,
	)
}

type PackageInternal struct {
	// Unexported fields are not part of the public API.
	Build   *build.Package
	Imports []*Package // this package's direct imports
	Local   bool       // imported via local path (./ or ../)

	Asmflags   []string // -asmflags for this package
	Gcflags    []string // -gcflags for this package
	Ldflags    []string // -ldflags for this package
	Gccgoflags []string // -gccgoflags for this package
}

type NoGoError struct {
	Package *Package
}

func (e *NoGoError) Error() string {
	// Count files beginning with _ and ., which we will pretend don't exist at all.
	dummy := 0
	for _, name := range e.Package.IgnoredGoFiles {
		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
			dummy++
		}
	}

	if len(e.Package.IgnoredGoFiles) > dummy {
		// Go files exist, but they were ignored due to build constraints.
		return "build constraints exclude all Go files in " + e.Package.Dir
	}
	return "no Go files in " + e.Package.Dir
}

func (p *Package) copyBuild(pp *build.Package) {
	p.Internal.Build = pp

	p.Dir = pp.Dir
	p.ImportPath = pp.ImportPath
	p.ImportComment = pp.ImportComment
	p.Name = pp.Name
	p.Doc = pp.Doc
	p.Root = pp.Root

	p.Goroot = pp.Goroot
	p.Standard = p.Goroot && p.ImportPath != "" && isStandardImportPath(p.ImportPath)
	p.GoFiles = pp.GoFiles
	p.IgnoredGoFiles = pp.IgnoredGoFiles
	// We modify p.Imports in place, so make copy now.
	p.Imports = make([]string, len(pp.Imports))
	copy(p.Imports, pp.Imports)
}

// isStandardImportPath reports whether $GOROOT/src/path should be considered
// part of the standard distribution. For historical reasons we allow people to add
// their own code to $GOROOT instead of using $GOPATH, but we assume that
// code will start with a domain name (dot in the first element).
func isStandardImportPath(path string) bool {
	i := strings.Index(path, "/")
	if i < 0 {
		i = len(path)
	}
	elem := path[:i]
	return !strings.Contains(elem, ".")
}

// A PackageError describes an error loading information about a package.
type PackageError struct {
	ImportStack   []string // shortest path from package named on command line to this one
	Pos           string   // position of error
	Err           string   // the error itself
	IsImportCycle bool     `json:"-"` // the error is an import cycle
	Hard          bool     `json:"-"` // whether the error is soft or hard; soft errors are ignored in some places
}

func (p *PackageError) Error() string {
	// Import cycles deserve special treatment.
	if p.IsImportCycle {
		return fmt.Sprintf("%s\npackage %s\n", p.Err, strings.Join(p.ImportStack, "\n\timports "))
	}
	if p.Pos != "" {
		// Omit import stack. The full path to the file where the error
		// is the most important thing.
		return p.Pos + ": " + p.Err
	}
	if len(p.ImportStack) == 0 {
		return p.Err
	}
	return "package " + strings.Join(p.ImportStack, "\n\timports ") + ": " + p.Err
}

// An ImportStack is a stack of import paths.
type ImportStack []string

func (s *ImportStack) Push(p string) {
	*s = append(*s, p)
}

func (s *ImportStack) Pop() {
	*s = (*s)[0 : len(*s)-1]
}

func (s *ImportStack) Copy() []string {
	return append([]string{}, *s...)
}

// shorterThan reports whether sp is shorter than t.
// We use this to record the shortest import sequence
// that leads to a particular package.
func (sp *ImportStack) shorterThan(t []string) bool {
	s := *sp
	if len(s) != len(t) {
		return len(s) < len(t)
	}
	// If they are the same length, settle ties using string ordering.
	for i := range s {
		if s[i] != t[i] {
			return s[i] < t[i]
		}
	}
	return false // they are equal
}

func (s *Session) ClearPackageCachePartial(args []string) {
	for _, arg := range args {
		p := s.packageCache[arg]
		if p != nil {
			delete(s.packageCache, p.Dir)
			delete(s.packageCache, p.ImportPath)
		}
	}
}

// dirToImportPath returns the pseudo-import path we use for a package
// outside the Go path. It begins with _/ and then contains the full path
// to the directory. If the package lives in c:\home\gopher\my\pkg then
// the pseudo-import path is _/c_/home/gopher/my/pkg.
// Using a pseudo-import path like this makes the ./ imports no longer
// a special case, so that all the code to deal with ordinary imports works
// automatically.
func dirToImportPath(dir string) string {
	return pathpkg.Join("_", strings.Map(makeImportValid, filepath.ToSlash(dir)))
}

func makeImportValid(r rune) rune {
	// Should match Go spec, compilers, and ../../go/parser/parser.go:/isValidImport.
	const illegalChars = `!"#$%&'()*,:;<=>?[\]^{|}` + "`\uFFFD"
	if !unicode.IsGraphic(r) || unicode.IsSpace(r) || strings.ContainsRune(illegalChars, r) {
		return '_'
	}
	return r
}

// LoadImport scans the directory named by path, which must be an import path,
// but possibly a local import path (an absolute file system path or one beginning
// with ./ or ../). A local relative path is interpreted relative to srcDir.
// It returns a *Package describing the package found in that directory.
func (s *Session) LoadImport(ctx context.Context, path, srcDir string, parent *Package, stk *ImportStack, useVendor bool) *Package {
	stk.Push(path)
	defer stk.Pop()

	// Determine canonical identifier for this package.
	// For a local import the identifier is the pseudo-import path
	// we create from the full directory to the package.
	// Otherwise it is the usual import path.
	// For vendored imports, it is the expanded form.
	importPath := path
	origPath := path
	isLocal := build.IsLocalImport(path)
	if isLocal {
		importPath = dirToImportPath(filepath.Join(srcDir, path))
	} else if useVendor {
		// We do our own vendor resolution, because we want to
		// find out the key to use in packageCache without the
		// overhead of repeated calls to buildContext.Import.
		// The code is also needed in a few other places anyway.
		path = s.VendoredImportPath(parent, path)
		importPath = path
	}

	p := s.packageCache[importPath]
	if p != nil {
		p = reusePackage(p, stk)
	} else {
		p = new(Package)
		p.cache = s
		p.Internal.Local = isLocal
		p.ImportPath = importPath
		s.packageCache[importPath] = p

		// Load package.
		// Import always returns bp != nil, even if an error occurs,
		// in order to return partial information.
		var bp *build.Package
		var err error
		buildMode := build.ImportComment
		if !useVendor || path != origPath {
			// Not vendoring, or we already found the vendored path.
			buildMode |= build.IgnoreVendor
		}

		if WithCancel(ctx, func() {
			bp, err = s.buildContext.Import(path, srcDir, buildMode)
		}) {
			err = ctx.Err()
		}
		bp.ImportPath = importPath
		if err == nil && !isLocal && bp.ImportComment != "" && bp.ImportComment != path &&
			!strings.Contains(path, "/vendor/") && !strings.HasPrefix(path, "vendor/") {
			err = fmt.Errorf("code in directory %s expects import %q", bp.Dir, bp.ImportComment)
		}
		p.load(ctx, stk, bp, err)

		if origPath != cleanImport(origPath) {
			p.Error = &PackageError{
				ImportStack: stk.Copy(),
				Err:         fmt.Sprintf("non-canonical import path: %q should be %q", origPath, pathpkg.Clean(origPath)),
			}
			p.Incomplete = true
		}
	}

	// Checked on every import because the rules depend on the code doing the importing.
	if perr := disallowInternal(srcDir, p, stk); perr != p {
		return perr
	}
	if useVendor {
		if perr := disallowVendor(srcDir, origPath, p, stk); perr != p {
			return perr
		}
	}

	if p.Name == "main" && parent != nil && parent.Dir != p.Dir {
		perr := *p
		perr.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         fmt.Sprintf("import %q is a program, not an importable package", path),
		}
		return &perr
	}

	if p.Internal.Local && parent != nil && !parent.Internal.Local {
		perr := *p
		perr.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         fmt.Sprintf("local import %q in non-local package", path),
		}
		return &perr
	}

	return p

}

func cleanImport(path string) string {
	orig := path
	path = pathpkg.Clean(path)
	if strings.HasPrefix(orig, "./") && path != ".." && !strings.HasPrefix(path, "../") {
		path = "./" + path
	}
	return path
}

func (s *Session) isDir(fpath string) bool {

	if strings.HasPrefix(fpath, "/goroot/") {
		fi, err := assets.Assets.Stat(fpath)
		if err != nil {
			return false
		}
		return fi.IsDir()
	}

	fi, err := s.fs.Stat(fpath)
	result := err == nil && fi.IsDir()
	return result
}

// VendoredImportPath returns the expansion of path when it appears in parent.
// If parent is x/y/z, then path might expand to x/y/z/vendor/path, x/y/vendor/path,
// x/vendor/path, vendor/path, or else stay path if none of those exist.
// VendoredImportPath returns the expanded path or, if no expansion is found, the original.
func (s *Session) VendoredImportPath(parent *Package, path string) (found string) {
	if parent == nil || parent.Root == "" {
		return path
	}

	dir := filepath.Clean(parent.Dir)
	root := filepath.Join(parent.Root, "src")

	if !HasFilePathPrefix(dir, root) || len(dir) <= len(root) || dir[len(root)] != filepath.Separator || parent.ImportPath != "command-line-arguments" && !parent.Internal.Local && filepath.Join(root, parent.ImportPath) != dir {
		panic(fmt.Sprintf("unexpected directory layout:\n"+
			"	import path: %s\n"+
			"	root: %s\n"+
			"	dir: %s\n"+
			"	expand root: %s\n"+
			"	expand dir: %s\n"+
			"	separator: %s",
			parent.ImportPath,
			filepath.Join(parent.Root, "src"),
			filepath.Clean(parent.Dir),
			root,
			dir,
			string(filepath.Separator)))
	}

	vpath := "vendor/" + path
	for i := len(dir); i >= len(root); i-- {
		if i < len(dir) && dir[i] != filepath.Separator {
			continue
		}
		// Note: checking for the vendor directory before checking
		// for the vendor/path directory helps us hit the
		// isDir cache more often. It also helps us prepare a more useful
		// list of places we looked, to report when an import is not found.
		if !s.isDir(filepath.Join(dir[:i], "vendor")) {
			continue
		}
		targ := filepath.Join(dir[:i], vpath)
		if s.isDir(targ) && s.hasGoFiles(targ) {
			importPath := parent.ImportPath
			if importPath == "command-line-arguments" {
				// If parent.ImportPath is 'command-line-arguments'.
				// set to relative directory to root (also chopped root directory)
				importPath = dir[len(root)+1:]
			}
			// We started with parent's dir c:\gopath\src\foo\bar\baz\quux\xyzzy.
			// We know the import path for parent's dir.
			// We chopped off some number of path elements and
			// added vendor\path to produce c:\gopath\src\foo\bar\baz\vendor\path.
			// Now we want to know the import path for that directory.
			// Construct it by chopping the same number of path elements
			// (actually the same number of bytes) from parent's import path
			// and then append /vendor/path.
			chopped := len(dir) - i
			if chopped == len(importPath)+1 {
				// We walked up from c:\gopath\src\foo\bar
				// and found c:\gopath\src\vendor\path.
				// We chopped \foo\bar (length 8) but the import path is "foo/bar" (length 7).
				// Use "vendor/path" without any prefix.
				return vpath
			}
			return importPath[:len(importPath)-chopped] + "/" + vpath
		}
	}
	return path
}

// hasGoFiles reports whether dir contains any files with names ending in .go.
// For a vendor check we must exclude directories that contain no .go files.
// Otherwise it is not possible to vendor just a/b/c and still import the
// non-vendored a/b. See golang.org/issue/13832.
func (s *Session) hasGoFiles(dir string) bool {

	var fis []os.FileInfo
	if strings.HasPrefix(dir, "/goroot/") {
		fis, _ = assets.Assets.ReadDir(dir)
	} else {
		fis, _ = s.fs.ReadDir(dir)
	}

	for _, fi := range fis {
		if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".go") {
			return true
		}
	}
	return false
}

// reusePackage reuses package p to satisfy the import at the top
// of the import stack stk. If this use causes an import loop,
// reusePackage updates p's error information to record the loop.
func reusePackage(p *Package, stk *ImportStack) *Package {
	// We use p.Internal.Imports==nil to detect a package that
	// is in the midst of its own loadPackage call
	// (all the recursion below happens before p.Internal.Imports gets set).
	if p.Internal.Imports == nil {
		if p.Error == nil {
			p.Error = &PackageError{
				ImportStack:   stk.Copy(),
				Err:           "import cycle not allowed",
				IsImportCycle: true,
			}
		}
		p.Incomplete = true
	}
	// Don't rewrite the import stack in the error if we have an import cycle.
	// If we do, we'll lose the path that describes the cycle.
	if p.Error != nil && !p.Error.IsImportCycle && stk.shorterThan(p.Error.ImportStack) {
		p.Error.ImportStack = stk.Copy()
	}
	return p
}

// disallowInternal checks that srcDir is allowed to import p.
// If the import is allowed, disallowInternal returns the original package p.
// If not, it returns a new package containing just an appropriate error.
func disallowInternal(srcDir string, p *Package, stk *ImportStack) *Package {
	// golang.org/s/go14internal:
	// An import of a path containing the element “internal”
	// is disallowed if the importing code is outside the tree
	// rooted at the parent of the “internal” directory.

	// There was an error loading the package; stop here.
	if p.Error != nil {
		return p
	}

	// The generated 'testmain' package is allowed to access testing/internal/...,
	// as if it were generated into the testing directory tree
	// (it's actually in a temporary directory outside any Go tree).
	// This cleans up a former kludge in passing functionality to the testing package.
	if strings.HasPrefix(p.ImportPath, "testing/internal") && len(*stk) >= 2 && (*stk)[len(*stk)-2] == "testmain" {
		return p
	}

	// The stack includes p.ImportPath.
	// If that's the only thing on the stack, we started
	// with a name given on the command line, not an
	// import. Anything listed on the command line is fine.
	if len(*stk) == 1 {
		return p
	}

	// Check for "internal" element: three cases depending on begin of string and/or end of string.
	i, ok := findInternal(p.ImportPath)
	if !ok {
		return p
	}

	// Internal is present.
	// Map import path back to directory corresponding to parent of internal.
	if i > 0 {
		i-- // rewind over slash in ".../internal"
	}
	parent := p.Dir[:i+len(p.Dir)-len(p.ImportPath)]
	if HasFilePathPrefix(filepath.Clean(srcDir), filepath.Clean(parent)) {
		return p
	}

	// Internal is present, and srcDir is outside parent's tree. Not allowed.
	perr := *p
	perr.Error = &PackageError{
		ImportStack: stk.Copy(),
		Err:         "use of internal package not allowed",
	}
	perr.Incomplete = true
	return &perr
}

// findInternal looks for the final "internal" path element in the given import path.
// If there isn't one, findInternal returns ok=false.
// Otherwise, findInternal returns ok=true and the index of the "internal".
func findInternal(path string) (index int, ok bool) {
	// Three cases, depending on internal at start/end of string or not.
	// The order matters: we must return the index of the final element,
	// because the final one produces the most restrictive requirement
	// on the importer.
	switch {
	case strings.HasSuffix(path, "/internal"):
		return len(path) - len("internal"), true
	case strings.Contains(path, "/internal/"):
		return strings.LastIndex(path, "/internal/") + 1, true
	case path == "internal", strings.HasPrefix(path, "internal/"):
		return 0, true
	}
	return 0, false
}

// disallowVendor checks that srcDir is allowed to import p as path.
// If the import is allowed, disallowVendor returns the original package p.
// If not, it returns a new package containing just an appropriate error.
func disallowVendor(srcDir, path string, p *Package, stk *ImportStack) *Package {
	// The stack includes p.ImportPath.
	// If that's the only thing on the stack, we started
	// with a name given on the command line, not an
	// import. Anything listed on the command line is fine.
	if len(*stk) == 1 {
		return p
	}

	if perr := disallowVendorVisibility(srcDir, p, stk); perr != p {
		return perr
	}

	// Paths like x/vendor/y must be imported as y, never as x/vendor/y.
	if i, ok := FindVendor(path); ok {
		perr := *p
		perr.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         "must be imported as " + path[i+len("vendor/"):],
		}
		perr.Incomplete = true
		return &perr
	}

	return p
}

// disallowVendorVisibility checks that srcDir is allowed to import p.
// The rules are the same as for /internal/ except that a path ending in /vendor
// is not subject to the rules, only subdirectories of vendor.
// This allows people to have packages and commands named vendor,
// for maximal compatibility with existing source trees.
func disallowVendorVisibility(srcDir string, p *Package, stk *ImportStack) *Package {
	// The stack includes p.ImportPath.
	// If that's the only thing on the stack, we started
	// with a name given on the command line, not an
	// import. Anything listed on the command line is fine.
	if len(*stk) == 1 {
		return p
	}

	// Check for "vendor" element.
	i, ok := FindVendor(p.ImportPath)
	if !ok {
		return p
	}

	// Vendor is present.
	// Map import path back to directory corresponding to parent of vendor.
	if i > 0 {
		i-- // rewind over slash in ".../vendor"
	}
	truncateTo := i + len(p.Dir) - len(p.ImportPath)
	if truncateTo < 0 || len(p.Dir) < truncateTo {
		return p
	}
	parent := p.Dir[:truncateTo]
	if HasFilePathPrefix(filepath.Clean(srcDir), filepath.Clean(parent)) {
		return p
	}

	// Vendor is present, and srcDir is outside parent's tree. Not allowed.
	perr := *p
	perr.Error = &PackageError{
		ImportStack: stk.Copy(),
		Err:         "use of vendored package not allowed",
	}
	perr.Incomplete = true
	return &perr
}

// FindVendor looks for the last non-terminating "vendor" path element in the given import path.
// If there isn't one, FindVendor returns ok=false.
// Otherwise, FindVendor returns ok=true and the index of the "vendor".
//
// Note that terminating "vendor" elements don't count: "x/vendor" is its own package,
// not the vendored copy of an import "" (the empty import path).
// This will allow people to have packages or commands named vendor.
// This may help reduce breakage, or it may just be confusing. We'll see.
func FindVendor(path string) (index int, ok bool) {
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

// load populates p using information from bp, err, which should
// be the result of calling build.Context.Import.
func (p *Package) load(ctx context.Context, stk *ImportStack, bp *build.Package, err error) {
	p.copyBuild(bp)

	if err != nil {
		if _, ok := err.(*build.NoGoError); ok {
			err = &NoGoError{Package: p}
		}
		p.Incomplete = true
		err = ExpandScanner(err)
		p.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         err.Error(),
		}
		return
	}

	// Build augmented import list to add implicit dependencies.
	// Be careful not to add imports twice, just to avoid confusion.
	importPaths := p.Imports
	addImport := func(path string) {
		for _, p := range importPaths {
			if path == p {
				return
			}
		}
		importPaths = append(importPaths, path)
	}

	// The linker loads implicit dependencies.
	if p.Name == "main" {
		for _, dep := range p.cache.LinkerDeps(p) {
			addImport(dep)
		}
	}

	// Check for case-insensitive collision of input files.
	// To avoid problems on case-insensitive files, we reject any package
	// where two different input files have equal names under a case-insensitive
	// comparison.
	inputs := p.AllFiles()
	f1, f2 := FoldDup(inputs)
	if f1 != "" {
		p.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         fmt.Sprintf("case-insensitive file name collision: %q and %q", f1, f2),
		}
		return
	}

	// If first letter of input file is ASCII, it must be alphanumeric.
	// This avoids files turning into flags when invoking commands,
	// and other problems we haven't thought of yet.
	// Also, _cgo_ files must be generated by us, not supplied.
	// They are allowed to have //go:cgo_ldflag directives.
	// The directory scan ignores files beginning with _,
	// so we shouldn't see any _cgo_ files anyway, but just be safe.
	for _, file := range inputs {
		if !SafeArg(file) || strings.HasPrefix(file, "_cgo_") {
			p.Error = &PackageError{
				ImportStack: stk.Copy(),
				Err:         fmt.Sprintf("invalid input file name %q", file),
			}
			return
		}
	}
	if name := pathpkg.Base(p.ImportPath); !SafeArg(name) {
		p.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         fmt.Sprintf("invalid input directory name %q", name),
		}
		return
	}
	if !SafeArg(p.ImportPath) {
		p.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         fmt.Sprintf("invalid import path %q", p.ImportPath),
		}
		return
	}

	// Build list of imported packages and full dependency list.
	imports := make([]*Package, 0, len(p.Imports))
	for i, path := range importPaths {
		if path == "C" {
			continue
		}
		p1 := p.cache.LoadImport(ctx, path, p.Dir, p, stk, true)
		if p.Standard && p.Error == nil && !p1.Standard && p1.Error == nil {
			p.Error = &PackageError{
				ImportStack: stk.Copy(),
				Err:         fmt.Sprintf("non-standard import %q in standard package %q", path, p.ImportPath),
			}
			pos := p.Internal.Build.ImportPos[path]
			if len(pos) > 0 {
				p.Error.Pos = pos[0].String()
			}
		}

		path = p1.ImportPath
		importPaths[i] = path
		if i < len(p.Imports) {
			p.Imports[i] = path
		}

		imports = append(imports, p1)
		if p1.Incomplete {
			p.Incomplete = true
		}
	}
	p.Internal.Imports = imports

	deps := make(map[string]*Package)
	var q []*Package
	q = append(q, imports...)
	for i := 0; i < len(q); i++ {
		p1 := q[i]
		path := p1.ImportPath
		// The same import path could produce an error or not,
		// depending on what tries to import it.
		// Prefer to record entries with errors, so we can report them.
		p0 := deps[path]
		if p0 == nil || p1.Error != nil && (p0.Error == nil || len(p0.Error.ImportStack) > len(p1.Error.ImportStack)) {
			deps[path] = p1
			for _, p2 := range p1.Internal.Imports {
				if deps[p2.ImportPath] != p2 {
					q = append(q, p2)
				}
			}
		}
	}
	p.Deps = make([]string, 0, len(deps))
	for dep := range deps {
		p.Deps = append(p.Deps, dep)
	}
	sort.Strings(p.Deps)
	for _, dep := range p.Deps {
		p1 := deps[dep]
		if p1 == nil {
			panic("impossible: missing entry in package cache for " + dep + " imported by " + p.ImportPath)
		}
		if p1.Error != nil {
			p.DepsErrors = append(p.DepsErrors, p1.Error)
		}
	}

	setError := func(msg string) {
		p.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         msg,
		}
	}

	// Check for case-insensitive collisions of import paths.
	fold := ToFold(p.ImportPath)
	if other := p.cache.foldPath[fold]; other == "" {
		p.cache.foldPath[fold] = p.ImportPath
	} else if other != p.ImportPath {
		setError(fmt.Sprintf("case-insensitive import collision: %q and %q", p.ImportPath, other))
		return
	}
}

// SafeArg reports whether arg is a "safe" command-line argument,
// meaning that when it appears in a command-line, it probably
// doesn't have some special meaning other than its own name.
// Obviously args beginning with - are not safe (they look like flags).
// Less obviously, args beginning with @ are not safe (they look like
// GNU binutils flagfile specifiers, sometimes called "response files").
// To be conservative, we reject almost any arg beginning with non-alphanumeric ASCII.
// We accept leading . _ and / as likely in file system paths.
// There is a copy of this function in cmd/compile/internal/gc/noder.go.
func SafeArg(name string) bool {
	if name == "" {
		return false
	}
	c := name[0]
	return '0' <= c && c <= '9' || 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || c == '.' || c == '_' || c == '/' || c >= utf8.RuneSelf
}

// LinkerDeps returns the list of linker-induced dependencies for main package p.
func (s *Session) LinkerDeps(p *Package) []string {
	// Everything links runtime.
	deps := []string{"runtime"}

	// On ARM with GOARM=5, it forces an import of math, for soft floating point.
	if s.buildContext.GOARCH == "arm" {
		deps = append(deps, "math")
	}

	return deps
}
