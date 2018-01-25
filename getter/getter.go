package getter

import (
	pathpkg "path"
	"path/filepath"

	"fmt"
	"go/build"
	"strings"

	"io"

	"sync"

	"github.com/dave/jsgo/common"
	"golang.org/x/sync/singleflight"
	"gopkg.in/src-d/go-billy.v4"
)

func New(fs billy.Filesystem, log io.Writer) *Cache {
	c := &Cache{}

	c.fs = fs
	c.log = log
	c.packageCache = make(map[string]*Package)
	c.foldPath = make(map[string]string)
	c.downloadCache = make(map[string]bool)
	c.repoRoots = make(map[string]*repoRoot)    // key is the root dir of the repo
	c.repoPackages = make(map[string]*repoRoot) // key is the path of the package. NOTE: not all packages are included, but the ones we're interested in should be.
	c.fetchCache = make(map[string]fetchResult)
	c.buildContext = common.NewBuildContext(fs, false) // getter doesn't need goroot
	return c

}

type Cache struct {
	fs            billy.Filesystem
	log           io.Writer
	packageCache  map[string]*Package
	buildContext  *build.Context
	foldPath      map[string]string
	downloadCache map[string]bool
	repoRoots     map[string]*repoRoot
	repoPackages  map[string]*repoRoot
	fetchGroup    singleflight.Group
	fetchCacheMu  sync.Mutex
	fetchCache    map[string]fetchResult // key is metaImportsForPrefix's importPrefix
}

func (c *Cache) Hashes() map[string]string {
	hashes := map[string]string{}
	for path, root := range c.repoPackages {
		hashes[path] = root.hash
	}
	return hashes
}

func (c *Cache) Get(path string, update bool, insecure bool) error {
	var stk importStack
	return c.get(path, nil, &stk, update, insecure)
}

func (c *Cache) get(path string, parent *Package, stk *importStack, update bool, insecure bool) error {
	load1 := func(path string, useVendor bool) *Package {
		if parent == nil {
			return c.Import(path, "/", nil, stk, false)
		}
		return c.Import(path, parent.Dir, parent, stk, useVendor)
	}
	p := load1(path, false)
	if p.Error != nil && p.Error.Hard {
		return p.Error
	}

	// loadPackage inferred the canonical ImportPath from arg.
	// Use that in the following to prevent hysteresis effects
	// in e.g. downloadCache and packageCache.
	// This allows invocations such as:
	//   mkdir -p $GOPATH/src/github.com/user
	//   cd $GOPATH/src/github.com/user
	//   go get ./foo
	// see: golang.org/issue/9767
	path = p.ImportPath

	// There's nothing to do if this is a package in the standard library.
	if p.Standard {
		return nil
	}

	// Only process each package once.
	// (Unless we're fetching test dependencies for this package,
	// in which case we want to process it again.)
	if c.downloadCache[path] {
		return nil
	}
	c.downloadCache[path] = true

	pkgs := []*Package{p}

	// Download if the package is missing, or update if we're using -u.
	if p.Dir == "" || update {
		// The actual download.
		stk.Push(path)
		err := c.downloadPackage(p, update, insecure)
		if err != nil {
			perr := &PackageError{ImportStack: stk.Copy(), Err: err.Error()}
			stk.Pop()
			return perr
		}
		stk.Pop()

		// Clear all relevant package cache entries before
		// doing any new loads.
		c.clearPackageCachePartial([]string{path})

		pkgs = pkgs[:0]

		// Note: load calls loadPackage or loadImport,
		// which push arg onto stk already.
		// Do not push here too, or else stk will say arg imports arg.
		p := load1(path, false)
		if p.Error != nil {
			return p.Error
		}
		pkgs = append(pkgs, p)

	}

	// Process package, which might now be multiple packages
	// due to wildcard expansion.
	for _, p := range pkgs {
		// Process dependencies, now that we know what they are.
		imports := p.Imports
		for i, path := range imports {
			if path == "C" {
				continue
			}
			// Fail fast on import naming full vendor path.
			// Otherwise expand path as needed for test imports.
			// Note that p.Imports can have additional entries beyond p.Internal.Build.Imports.
			orig := path
			if i < len(p.Build.Imports) {
				orig = p.Build.Imports[i]
			}
			if j, ok := findVendor(orig); ok {
				stk.Push(path)
				err := &PackageError{
					ImportStack: stk.Copy(),
					Err:         "must be imported as " + path[j+len("vendor/"):],
				}
				stk.Pop()
				return err
			}
			// If this is a test import, apply vendor lookup now.
			// We cannot pass useVendor to download, because
			// download does caching based on the value of path,
			// so it must be the fully qualified path already.
			if i >= len(p.Imports) {
				path = c.vendoredImportPath(p, path)
			}
			c.get(path, p, stk, update, insecure)
		}
	}

	return nil
}

func (c *Cache) clearPackageCachePartial(args []string) {
	for _, arg := range args {
		p := c.packageCache[arg]
		if p != nil {
			delete(c.packageCache, p.Dir)
			delete(c.packageCache, p.ImportPath)
		}
	}
}

func (c *Cache) Import(path, srcDir string, parent *Package, stk *importStack, useVendor bool) *Package {
	stk.Push(path)
	defer stk.Pop()

	// Determine canonical identifier for this package.
	// For a local import the identifier is the pseudo-import path
	// we create from the full directory to the package.
	// Otherwise it is the usual import path.
	// For vendored imports, it is the expanded form.
	importPath := path
	origPath := path
	isLocal := isLocalImport(path)
	if isLocal {
		importPath = dirToImportPath(filepath.Join(srcDir, path))
	} else if useVendor {
		// We do our own vendor resolution, because we want to
		// find out the key to use in packageCache without the
		// overhead of repeated calls to buildContext.Import.
		// The code is also needed in a few other places anyway.
		path = c.vendoredImportPath(parent, path)
		importPath = path
	}

	p := c.packageCache[importPath]
	if p != nil {
		p = reusePackage(p, stk)
	} else {
		p = new(Package)
		p.cache = c
		p.Local = isLocal
		p.ImportPath = importPath
		c.packageCache[importPath] = p

		// Load package.
		// Import always returns bp != nil, even if an error occurs,
		// in order to return partial information.
		buildMode := build.ImportComment
		if !useVendor || path != origPath {
			// Not vendoring, or we already found the vendored path.
			buildMode |= build.IgnoreVendor
		}
		bp, err := c.buildContext.Import(path, srcDir, buildMode)
		bp.ImportPath = importPath
		if err == nil && !isLocal && bp.ImportComment != "" && bp.ImportComment != path &&
			!strings.Contains(path, "/vendor/") && !strings.HasPrefix(path, "vendor/") {
			err = fmt.Errorf("code in directory %s expects import %q", bp.Dir, bp.ImportComment)
		}
		p.load(stk, bp, err)

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

	if p.Local && parent != nil && !parent.Local {
		perr := *p
		perr.Error = &PackageError{
			ImportStack: stk.Copy(),
			Err:         fmt.Sprintf("local import %q in non-local package", path),
		}
		return &perr
	}

	return p

}

// reusePackage reuses package p to satisfy the import at the top
// of the import stack stk. If this use causes an import loop,
// reusePackage updates p's error information to record the loop.
func reusePackage(p *Package, stk *importStack) *Package {
	// We use p.InternalImports==nil to detect a package that
	// is in the midst of its own loadPackage call
	// (all the recursion below happens before p.InternalImports gets set).
	if p.InternalImports == nil {
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
