package getter

import (
	"errors"
	"fmt"
	"go/build"
	"go/scanner"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"bytes"
	"unicode/utf8"

	"context"
	"crypto/sha1"
)

type Package struct {
	cache              *Cache
	Dir                string          // directory containing package sources
	ImportPath         string          // import path of package in dir
	Root               string          // Go root or Go path dir containing this package
	Local              bool            // imported via local path (./ or ../)
	Error              *PackageError   // error loading this package (not dependencies)
	Incomplete         bool            // was there an error loading this package or dependencies?
	Imports            []string        // import paths used by this package
	Deps               []string        // all (recursively) imported dependencies
	IgnoredGoFiles     []string        // .go sources ignored due to build constraints
	ImportComment      string          // path in import comment on package statement
	Name               string          // package name
	Doc                string          // package documentation string
	Goroot             bool            // is this package found in the Go root?
	Standard           bool            // is this package part of the standard Go library?
	GoFiles            []string        // .go source files
	DepsErrors         []*PackageError // errors loading dependencies
	BuildID            string          // expected build ID for generated package
	InternalImports    []*Package
	InternalDeps       []*Package
	Build              *build.Package
	InternalGoFiles    []string
	InternalAllGoFiles []string
}

// load populates p using information from bp, err, which should
// be the result of calling build.Context.Import.
func (p *Package) load(ctx context.Context, stk *importStack, bp *build.Package, err error) *Package {
	p.copyBuild(bp)

	// The localPrefix is the path we interpret ./ imports relative to.
	// Synthesized main packages sometimes override this.
	//p.Internal.LocalPrefix = dirToImportPath(p.Dir)

	if err != nil {
		if _, ok := err.(*build.NoGoError); ok {
			err = &NoGoError{Package: p}
		}
		p.Incomplete = true
		err = expandScanner(err)
		p.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         err.Error(),
		}
		return p
	}

	importPaths := p.Imports

	// Everything depends on runtime, except runtime, its internal
	// subpackages, and unsafe.
	if !p.Standard || (p.ImportPath != "runtime" && !strings.HasPrefix(p.ImportPath, "runtime/internal/") && p.ImportPath != "unsafe") {
		importPaths = append(importPaths, "runtime")
	}

	// Runtime and its internal packages depend on runtime/internal/sys,
	// so that they pick up the generated zversion.go file.
	// This can be an issue particularly for runtime/internal/atomic;
	// see issue 13655.
	if p.Standard && (p.ImportPath == "runtime" || strings.HasPrefix(p.ImportPath, "runtime/internal/")) && p.ImportPath != "runtime/internal/sys" {
		importPaths = append(importPaths, "runtime/internal/sys")
	}

	// Build list of full paths to all Go files in the package,
	// for use by commands like go fmt.
	p.InternalGoFiles = stringList(p.GoFiles)
	for i := range p.InternalGoFiles {
		p.InternalGoFiles[i] = filepath.Join(p.Dir, p.InternalGoFiles[i])
	}
	sort.Strings(p.InternalGoFiles)

	p.InternalAllGoFiles = stringList(p.IgnoredGoFiles)
	for i := range p.InternalAllGoFiles {
		p.InternalAllGoFiles[i] = filepath.Join(p.Dir, p.InternalAllGoFiles[i])
	}
	p.InternalAllGoFiles = append(p.InternalAllGoFiles, p.InternalGoFiles...)
	sort.Strings(p.InternalAllGoFiles)

	// Check for case-insensitive collision of input files.
	// To avoid problems on case-insensitive files, we reject any package
	// where two different input files have equal names under a case-insensitive
	// comparison.
	f1, f2 := foldDup(stringList(
		p.GoFiles,
		p.IgnoredGoFiles,
	))
	if f1 != "" {
		p.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         fmt.Sprintf("case-insensitive file name collision: %q and %q", f1, f2),
		}
		return p
	}

	// Build list of imported packages and full dependency list.
	imports := make([]*Package, 0, len(p.Imports))
	deps := make(map[string]*Package)
	save := func(path string, p1 *Package) {
		// The same import path could produce an error or not,
		// depending on what tries to import it.
		// Prefer to record entries with errors, so we can report them.
		p0 := deps[path]
		if p0 == nil || p1.Error != nil && (p0.Error == nil || len(p0.Error.ImportStack) > len(p1.Error.ImportStack)) {
			deps[path] = p1
		}
	}

	for i, path := range importPaths {
		if path == "C" {
			continue
		}
		p1 := p.cache.Import(ctx, path, p.Dir, p, stk, true)
		if p.Standard && p.Error == nil && !p1.Standard && p1.Error == nil {
			p.Error = &PackageError{
				ImportStack: stk.Copy(),
				Err:         fmt.Sprintf("non-standard import %q in standard package %q", path, p.ImportPath),
			}
			pos := p.Build.ImportPos[path]
			if len(pos) > 0 {
				p.Error.Pos = pos[0].String()
			}
		}

		path = p1.ImportPath
		importPaths[i] = path
		if i < len(p.Imports) {
			p.Imports[i] = path
		}

		save(path, p1)
		imports = append(imports, p1)
		for _, dep := range p1.InternalDeps {
			save(dep.ImportPath, dep)
		}
		if p1.Incomplete {
			p.Incomplete = true
		}
	}
	p.InternalImports = imports

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
		p.InternalDeps = append(p.InternalDeps, p1)
		if p1.Error != nil {
			p.DepsErrors = append(p.DepsErrors, p1.Error)
		}
	}

	// unsafe is a fake package.
	//if p.Standard && (p.ImportPath == "unsafe") {
	//	p.Internal.Target = ""
	//}
	//p.Target = p.Internal.Target

	// Check for case-insensitive collisions of import paths.
	fold := toFold(p.ImportPath)
	if other := p.cache.foldPath[fold]; other == "" {
		p.cache.foldPath[fold] = p.ImportPath
	} else if other != p.ImportPath {
		p.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         fmt.Sprintf("case-insensitive import collision: %q and %q", p.ImportPath, other),
		}
		return p
	}

	p.cache.computeBuildID(p)
	return p
}

// ExpandScanner expands a scanner.List error into all the errors in the list.
// The default Error method only shows the first error
// and does not shorten paths.
func expandScanner(err error) error {
	// Look for parser errors.
	if err, ok := err.(scanner.ErrorList); ok {
		// Prepare error with \n before each message.
		// When printed in something like context: %v
		// this will put the leading file positions each on
		// its own line. It will also show all the errors
		// instead of just the first, as err.Error does.
		var buf bytes.Buffer
		for _, e := range err {
			buf.WriteString("\n")
			buf.WriteString(e.Error())
		}
		return errors.New(buf.String())
	}
	return err
}

func (p *Package) copyBuild(pp *build.Package) {
	p.Build = pp

	//if pp.PkgTargetRoot != "" && cfg.BuildPkgdir != "" {
	//	old := pp.PkgTargetRoot
	//	pp.PkgRoot = cfg.BuildPkgdir
	//	pp.PkgTargetRoot = cfg.BuildPkgdir
	//	pp.PkgObj = filepath.Join(cfg.BuildPkgdir, strings.TrimPrefix(pp.PkgObj, old))
	//}

	p.Dir = pp.Dir
	p.ImportPath = pp.ImportPath
	p.ImportComment = pp.ImportComment
	p.Name = pp.Name
	p.Doc = pp.Doc
	p.Root = pp.Root
	//p.ConflictDir = pp.ConflictDir
	//p.BinaryOnly = pp.BinaryOnly

	p.Goroot = pp.Goroot
	p.Standard = p.Goroot && p.ImportPath != "" && isStandardImportPath(p.ImportPath)
	//p.GoFiles = pp.GoFiles
	//p.CgoFiles = pp.CgoFiles
	p.IgnoredGoFiles = pp.IgnoredGoFiles
	//p.CFiles = pp.CFiles
	//p.CXXFiles = pp.CXXFiles
	//p.MFiles = pp.MFiles
	//p.HFiles = pp.HFiles
	//p.FFiles = pp.FFiles
	//p.SFiles = pp.SFiles
	//p.SwigFiles = pp.SwigFiles
	//p.SwigCXXFiles = pp.SwigCXXFiles
	//p.SysoFiles = pp.SysoFiles
	//p.CgoCFLAGS = pp.CgoCFLAGS
	//p.CgoCPPFLAGS = pp.CgoCPPFLAGS
	//p.CgoCXXFLAGS = pp.CgoCXXFLAGS
	//p.CgoFFLAGS = pp.CgoFFLAGS
	//p.CgoLDFLAGS = pp.CgoLDFLAGS
	//p.CgoPkgConfig = pp.CgoPkgConfig
	// We modify p.Imports in place, so make copy now.
	p.Imports = make([]string, len(pp.Imports))
	copy(p.Imports, pp.Imports)
	//p.TestGoFiles = pp.TestGoFiles
	//p.TestImports = pp.TestImports
	//p.XTestGoFiles = pp.XTestGoFiles
	//p.XTestImports = pp.XTestImports
	//if IgnoreImports {
	//	p.Imports = nil
	//	p.TestImports = nil
	//	p.XTestImports = nil
	//}
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

// StringList flattens its arguments into a single []string.
// Each argument in args must have type string or []string.
func stringList(args ...interface{}) []string {
	var x []string
	for _, arg := range args {
		switch arg := arg.(type) {
		case []string:
			x = append(x, arg...)
		case string:
			x = append(x, arg)
		default:
			panic("stringList: invalid argument of type " + fmt.Sprintf("%T", arg))
		}
	}
	return x
}

// FoldDup reports a pair of strings from the list that are
// equal according to strings.EqualFold.
// It returns "", "" if there are no such strings.
func foldDup(list []string) (string, string) {
	clash := map[string]string{}
	for _, s := range list {
		fold := toFold(s)
		if t := clash[fold]; t != "" {
			if s > t {
				s, t = t, s
			}
			return s, t
		}
		clash[fold] = s
	}
	return "", ""
}

// ToFold returns a string with the property that
//	strings.EqualFold(s, t) iff ToFold(s) == ToFold(t)
// This lets us test a large set of strings for fold-equivalent
// duplicates without making a quadratic number of calls
// to EqualFold. Note that strings.ToUpper and strings.ToLower
// do not have the desired property in some corner cases.
func toFold(s string) string {
	// Fast path: all ASCII, no upper case.
	// Most paths look like this already.
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf || 'A' <= c && c <= 'Z' {
			goto Slow
		}
	}
	return s

Slow:
	var buf bytes.Buffer
	for _, r := range s {
		// SimpleFold(x) cycles to the next equivalent rune > x
		// or wraps around to smaller values. Iterate until it wraps,
		// and we've found the minimum value.
		for {
			r0 := r
			r = unicode.SimpleFold(r0)
			if r <= r0 {
				break
			}
		}
		// Exception to allow fast path above: A-Z => a-z
		if 'A' <= r && r <= 'Z' {
			r += 'a' - 'A'
		}
		buf.WriteRune(r)
	}
	return buf.String()
}

// computeBuildID computes the build ID for p, leaving it in p.Internal.BuildID.
// Build ID is a hash of the information we want to detect changes in.
// See the long comment in isStale for details.
func (c *Cache) computeBuildID(p *Package) {
	h := sha1.New()

	// Include the list of files compiled as part of the package.
	// This lets us detect removed files. See issue 3895.
	inputFiles := stringList(
		p.GoFiles,
	)
	for _, file := range inputFiles {
		fmt.Fprintf(h, "file %s\n", file)
	}

	// Include the content of runtime/internal/sys/zversion.go in the hash
	// for package runtime. This will give package runtime a
	// different build ID in each Go release.
	if p.Standard && p.ImportPath == "runtime/internal/sys" {
		zversion := `// auto generated by go tool dist

package sys

const DefaultGoroot = "/usr/local/go"
const TheVersion = "go1.9"
const Goexperiment = ""
const StackGuardMultiplier = 1

`
		fmt.Fprintf(h, "zversion %q\n", zversion)
	}

	// Include the build IDs of any dependencies in the hash.
	// This, combined with the runtime/zversion content,
	// will cause packages to have different build IDs when
	// compiled with different Go releases.
	// This helps the go command know to recompile when
	// people use the same GOPATH but switch between
	// different Go releases. See issue 10702.
	// This is also a better fix for issue 8290.
	for _, p1 := range p.InternalDeps {
		fmt.Fprintf(h, "dep %s %s\n", p1.ImportPath, p1.BuildID)
	}

	p.BuildID = fmt.Sprintf("%x", h.Sum(nil))
}

func cleanImport(path string) string {
	orig := path
	path = pathpkg.Clean(path)
	if strings.HasPrefix(orig, "./") && path != ".." && !strings.HasPrefix(path, "../") {
		path = "./" + path
	}
	return path
}

// disallowInternal checks that srcDir is allowed to import p.
// If the import is allowed, disallowInternal returns the original package p.
// If not, it returns a new package containing just an appropriate error.
func disallowInternal(srcDir string, p *Package, stk *importStack) *Package {
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
	if hasFilePathPrefix(filepath.Clean(srcDir), filepath.Clean(parent)) {
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
func disallowVendor(srcDir, path string, p *Package, stk *importStack) *Package {
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
	if i, ok := findVendor(path); ok {
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
func disallowVendorVisibility(srcDir string, p *Package, stk *importStack) *Package {
	// The stack includes p.ImportPath.
	// If that's the only thing on the stack, we started
	// with a name given on the command line, not an
	// import. Anything listed on the command line is fine.
	if len(*stk) == 1 {
		return p
	}

	// Check for "vendor" element.
	i, ok := findVendor(p.ImportPath)
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
	if hasFilePathPrefix(filepath.Clean(srcDir), filepath.Clean(parent)) {
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
