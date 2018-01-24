package compiler

import (
	"go/parser"
	"go/types"
	"os"

	"sync"

	"encoding/gob"
	"fmt"

	"io"

	"strings"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/common"
	"github.com/dave/jsgo/config"
	"github.com/gopherjs/gopherjs/compiler"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/loader"
	"gopkg.in/src-d/go-billy.v4"
)

func New(fs billy.Filesystem) *Cache {
	c := &Cache{
		fs:        fs,
		archivesM: &sync.RWMutex{},
		archives:  make(map[string]*compiler.Archive),
	}
	return c
}

type Cache struct {
	fs        billy.Filesystem
	archivesM *sync.RWMutex
	archives  map[string]*compiler.Archive
	prog      *loader.Program
}

type ArchiveInfo struct {
	Path     string
	Standard bool
	Archive  *compiler.Archive
}

func (c *Cache) Compile(path string, logger io.Writer) ([]ArchiveInfo, error) {

	conf := loader.Config{}
	conf.Import(path)
	conf.ParserMode = parser.ParseComments
	conf.Build = common.NewBuildContext(c.fs, true)

	prog, err := conf.Load()
	if err != nil {
		return nil, err
	}
	c.prog = prog

	var importContext *compiler.ImportContext
	importContext = &compiler.ImportContext{
		Packages: make(map[string]*types.Package),
		Import: func(path string) (*compiler.Archive, error) {

			// find in local cache
			if a, ok := c.getArchive(path); ok {
				return a, nil
			}

			pi := c.prog.Package(path)
			importContext.Packages[path] = pi.Pkg

			// find in standard library cache
			a, err := openStaticArchive(path)
			if err != nil {
				return nil, err
			}

			if a != nil {
				c.setArchive(path, a)
				return a, nil
			}

			fmt.Fprintf(logger, "Compiling %s\n", path)

			// compile package
			minify := !config.DEV
			a, err = compiler.Compile(path, pi.Files, c.prog.Fset, importContext, minify)
			if err != nil {
				return nil, err
			}
			c.setArchive(path, a)
			return a, nil
		},
	}

	if _, err := importContext.Import(path); err != nil {
		return nil, err
	}

	orderedArchives, err := c.orderImports(path)
	if err != nil {
		return nil, err
	}

	return orderedArchives, nil
}

func (c *Cache) getArchive(path string) (*compiler.Archive, bool) {
	c.archivesM.RLock()
	defer c.archivesM.RUnlock()
	a, ok := c.archives[path]
	return a, ok
}

func (c *Cache) setArchive(path string, a *compiler.Archive) {
	c.archivesM.Lock()
	defer c.archivesM.Unlock()
	c.archives[path] = a
}

func openStaticArchive(path string) (*compiler.Archive, error) {

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

func (c *Cache) orderImports(path string) ([]ArchiveInfo, error) {

	pkgJs, err := openStaticArchive("github.com/gopherjs/gopherjs/js")
	if err != nil {
		return nil, err
	}
	pkgNosync, err := openStaticArchive("github.com/gopherjs/gopherjs/nosync")
	if err != nil {
		return nil, err
	}

	orderedPackages := []ArchiveInfo{
		{
			Path:     "github.com/gopherjs/gopherjs/js",
			Standard: true,
			Archive:  pkgJs,
		},
		{
			Path:     "github.com/gopherjs/gopherjs/nosync",
			Standard: true,
			Archive:  pkgNosync,
		},
	}

	done := map[string]struct{}{
		"github.com/gopherjs/gopherjs/js":     {},
		"github.com/gopherjs/gopherjs/nosync": {},
	}

	var orderImports func(string) error
	orderImports = func(path string) error {

		done[path] = struct{}{}

		pi := c.prog.Package(path)

		if len(pi.Files) == 0 {
			return nil
		}

		std := strings.HasPrefix(c.prog.Fset.File(pi.Files[0].Pos()).Name(), "/goroot/")

		for _, child := range pi.Pkg.Imports() {
			if _, ok := done[child.Path()]; ok {
				continue
			}
			if err := orderImports(child.Path()); err != nil {
				return err
			}
		}

		var arch *compiler.Archive

		if std {
			var err error
			arch, err = openStaticArchive(path)
			if err != nil {
				return err
			}
			if arch == nil {
				// some packages e.g. internal/cpu, runtime/internal/atomic etc. are not needed by
				// gopherjs, so we don't have them in the static archive cache
				return nil
			}
		} else {
			var ok bool
			arch, ok = c.archives[path]
			if !ok {
				return fmt.Errorf("can't find import %s", path)
			}
		}

		orderedPackages = append(orderedPackages, ArchiveInfo{Path: path, Standard: std, Archive: arch})
		return nil
	}
	if err := orderImports(path); err != nil {
		return nil, err
	}

	return orderedPackages, nil
}

func WriteArchive(w io.Writer, archive *compiler.Archive) error {

	selection := make(map[*compiler.Decl]struct{})
	for _, d := range archive.Declarations {
		selection[d] = struct{}{}
	}

	fmt.Fprintf(w, `$initialisers["%s"] = function () {`, archive.ImportPath)

	if err := compiler.WritePkgCode(archive, selection, false, &compiler.SourceMapFilter{Writer: w}); err != nil {
		return err
	}

	fmt.Fprint(w, "};")

	return nil
}
