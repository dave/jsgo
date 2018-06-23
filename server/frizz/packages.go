package frizz

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"net/http"
	"path/filepath"
	"strings"

	"crypto/sha1"
	"encoding/json"

	"fmt"

	"bytes"
	"io"

	"sort"

	"github.com/dave/frizz/models"
	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/frizz/gotypes"
	"github.com/dave/jsgo/server/frizz/gotypes/convert"
	"github.com/dave/jsgo/server/frizz/messages"
	"github.com/dave/services"
	"github.com/dave/services/constor"
	"github.com/dave/services/getter/get"
	"github.com/dave/services/getter/gettermsg"
	"github.com/dave/services/session"
	"github.com/dave/services/srcimporter"
	"github.com/dave/stablegob"
)

func init() {
	gotypes.RegisterTypesStablegob()
}

func (h *Handler) Packages(ctx context.Context, info messages.GetPackages, req *http.Request, send func(message services.Message), receive chan services.Message) error {

	storer := constor.New(ctx, h.Fileserver, send, config.ConcurrentStorageUploads)
	defer storer.Close()

	// Send a message to the client that downloading step has started.
	send(gettermsg.Downloading{Starting: true})

	var save bool
	if len(info.Tags) == 0 {
		// only save the getter hints if there's no special build tags (this may affect dependency tree)
		// TODO: tweak resolver to incorporate build tags into hints key
		save = true
	}

	gitreq := h.Cache.NewRequest(save)
	if err := gitreq.InitialiseFromHints(ctx, info.Path); err != nil {
		return err
	}

	// set insecure = true in local mode or it will fail if git repo has git protocol
	insecure := config.LOCAL

	s := session.New(info.Tags, assets.Assets, assets.Archives, h.Fileserver, config.ValidExtensions)

	done := map[string]bool{}
	index := messages.PackageIndex{
		Path:    info.Path,
		Tags:    info.Tags,
		Source:  map[string]messages.IndexItem{},
		Objects: map[string]messages.IndexItem{},
	}
	source := map[string]map[string]string{}

	// Start the download process - just like the "go get" command.
	g := get.New(s, send, gitreq)

	g.Callback = func(path string, files map[string]string, standard bool) error {
		if path != info.Path {
			// only return source for main package (not all dependencies)
			return nil
		}
		if done[path] {
			return nil
		}
		done[path] = true

		var hash string
		var unchanged bool

		// internal/cpu not in std.Source? Let's ignore standard in callback signature for now since
		// we don't have it readily accessible in the objects code below, and we don't want standard
		// being different.
		_, foundInStandard := std.Source[path]

		if foundInStandard {
			hash = std.Source[path]
			if cached, ok := info.Source[path]; ok && cached == hash {
				unchanged = true
			}
			if !unchanged {
				send(messages.Source{
					Path:     path,
					Hash:     hash,
					Standard: foundInStandard,
				})
			}
		} else {
			source[path] = files
			s := models.SourcePack{
				Path:  path,
				Files: files,
			}
			sha := sha1.New()
			buf := &bytes.Buffer{}
			mw := io.MultiWriter(sha, buf)
			if err := json.NewEncoder(mw).Encode(s); err != nil {
				return err
			}
			hash = fmt.Sprintf("%x", sha.Sum(nil))
			if cached, ok := info.Source[path]; ok && cached == hash {
				unchanged = true
			}
			if !unchanged {
				storer.Add(constor.Item{
					Message:   path,
					Bucket:    config.Bucket[config.Pkg],
					Name:      fmt.Sprintf("%s.%s.json", path, hash), // Note: hash is a string
					Contents:  buf.Bytes(),
					Mime:      constor.MimeJson,
					Immutable: true,
					Count:     true,
					Send:      true,
					Done: func() {
						send(messages.Source{
							Path:     path,
							Hash:     hash,
							Standard: foundInStandard,
						})
					},
				})
			}
		}

		index.Source[path] = messages.IndexItem{
			Hash:      hash,
			Unchanged: unchanged,
		}

		return nil
	}

	if err := g.Get(ctx, info.Path, false, insecure, false); err != nil {
		return err
	}

	// Parse for types
	fset := token.NewFileSet()
	bctx := s.BuildContext(session.DefaultType, "")
	parsed := []*ast.File{}

	// Files must be in the same order each time
	type file struct{ name, contents string }
	var sorted []file
	for name, contents := range source[info.Path] {
		sorted = append(sorted, file{name, contents})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].name < sorted[j].name })

	for _, f := range sorted {
		if !strings.HasSuffix(f.name, ".go") || strings.HasSuffix(f.name, "_test.go") {
			continue
		}
		match, err := bctx.MatchFile(filepath.Join(bctx.GOPATH, "src", info.Path), f.name)
		if err != nil {
			return err
		}
		if !match {
			continue
		}
		astfile, err := parser.ParseFile(fset, filepath.Join(bctx.GOPATH, "src", info.Path, f.name), []byte(f.contents), 0)
		if err != nil {
			return err
		}
		parsed = append(parsed, astfile)
	}
	packages := map[string]*types.Package{}
	tc := types.Config{
		Importer: srcimporter.New(bctx, fset, packages),
		Error: func(err error) {
			// Ignore errors here - we should be able to load broken code.
		},
	}
	ti := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
	}
	p, err := tc.Check(info.Path, fset, parsed, ti)
	if err != nil {
		// Ignore errors here - we should be able to load broken code.
	}
	packages[info.Path] = p

	objects := map[string]map[string]map[string]gotypes.Object{}
	for _, p := range packages {
		p := p // used in callback

		if p == nil {
			continue
		}

		_, foundInStandard := std.Objects[p.Path()]

		var hash string
		var unchanged bool

		if foundInStandard {
			hash = std.Objects[p.Path()]
			if cached, ok := info.Objects[p.Path()]; ok && cached == hash {
				unchanged = true
			}
			if !unchanged {
				send(messages.Objects{
					Path:     p.Path(),
					Hash:     hash,
					Standard: foundInStandard,
				})
			}
		} else {
			for _, name := range p.Scope().Names() {
				v := p.Scope().Lookup(name)
				if v == nil {
					continue
				}
				if !v.Exported() {
					continue
				}

				object := convert.Object(v)
				path := p.Path()
				name := object.Object().Name
				_, file := filepath.Split(fset.File(v.Pos()).Name())

				if objects[path] == nil {
					objects[path] = map[string]map[string]gotypes.Object{}
				}
				if objects[path][file] == nil {
					objects[path][file] = map[string]gotypes.Object{}
				}
				objects[path][file][name] = object
			}
			pp := models.ObjectPack{
				Path:    p.Path(),
				Name:    p.Name(),
				Objects: objects[p.Path()],
			}
			sha := sha1.New()
			buf := &bytes.Buffer{}
			mw := io.MultiWriter(sha, buf)
			if err := stablegob.NewEncoder(mw).Encode(pp); err != nil {
				return err
			}
			hash = fmt.Sprintf("%x", sha.Sum(nil))
			if cached, ok := info.Objects[p.Path()]; ok && cached == hash {
				unchanged = true
			}
			if !unchanged {
				storer.Add(constor.Item{
					Message:   p.Path(),
					Bucket:    config.Bucket[config.Pkg],
					Name:      fmt.Sprintf("%s.%s.objects.gob", p.Path(), hash), // Note: hash is a string
					Contents:  buf.Bytes(),
					Mime:      constor.MimeBin,
					Immutable: true,
					Count:     true,
					Send:      true,
					Done: func() {
						send(messages.Objects{
							Path:     p.Path(),
							Hash:     hash,
							Standard: foundInStandard,
						})
					},
				})
			}
		}

		index.Objects[p.Path()] = messages.IndexItem{
			Hash:      hash,
			Unchanged: unchanged,
		}

	}

	if err := storer.Wait(); err != nil {
		return err
	}

	if err := gitreq.Close(ctx); err != nil {
		return err
	}

	send(index)

	// Send a message to the client that downloading step has finished.
	send(gettermsg.Downloading{Done: true})
	return nil
}
