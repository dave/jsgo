package compile

import (
	"bytes"
	"context"
	"fmt"

	"strings"

	"sync"

	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/services/builder"
	"github.com/dave/services/fileserver/constor"
	"github.com/gopherjs/gopherjs/compiler"
)

func (c *Compiler) Update(ctx context.Context, source map[string]map[string]string, cache map[string]string, min bool) error {

	storer := constor.New(ctx, c.fileserver, config.ConcurrentStorageUploads)
	defer storer.Close()

	c.send(messages.Updating{Starting: true})

	b := builder.New(c.Session, c.defaultOptions(updateWriter{c.send}, min))

	index := messages.Index{}
	done := map[string]bool{}

	b.Callback = func(archive *compiler.Archive) error {

		if done[archive.ImportPath] {
			return nil
		}

		done[archive.ImportPath] = true

		if archive.Name == "main" {
			return nil
		}

		if c.HasSource(archive.ImportPath) {
			// don't return anything if the package is in the source collection
			return nil
		}

		hashPair, standard := std.Index[archive.ImportPath]
		var hash string
		var js []byte
		if standard {
			hash = hashPair[min]
		} else {
			var b []byte
			var err error
			js, b, err = builder.GetPackageCode(ctx, archive, min, true)
			if err != nil {
				return err
			}
			hash = fmt.Sprintf("%x", b)
		}

		var unchanged bool
		if cached, exists := cache[archive.ImportPath]; exists && cached == hash {
			unchanged = true
		}

		index[archive.ImportPath] = messages.IndexItem{
			Hash:      hash,
			Unchanged: unchanged,
		}

		if unchanged {
			// If the dependency is unchanged from the client cache, don't return it as a PlaygroundArchive
			// message
			return nil
		}

		if !standard {
			var wait sync.WaitGroup
			wait.Add(2)
			storer.Add(constor.Item{
				Message:   archive.Name,
				Name:      fmt.Sprintf("%s.%s.js", archive.ImportPath, hash), // Note: hash is a string
				Contents:  js,
				Bucket:    config.PkgBucket,
				Mime:      constor.MimeJs,
				Count:     true,
				Immutable: true,
				Wait:      &wait,
				Changed: func(done bool) {
					messages.SendStoring(c.send, storer.Stats)
				},
			})
			buf := &bytes.Buffer{}
			if err := compiler.WriteArchive(StripArchive(archive), buf); err != nil {
				return err
			}
			storer.Add(constor.Item{
				Message:   "",
				Name:      fmt.Sprintf("%s.%s.ax", archive.ImportPath, hash), // Note: hash is a string
				Contents:  buf.Bytes(),
				Bucket:    config.PkgBucket,
				Mime:      constor.MimeBin,
				Count:     true,
				Immutable: true,
				Wait:      &wait,
				Changed: func(done bool) {
					messages.SendStoring(c.send, storer.Stats)
				},
			})
			wait.Wait()
		}

		c.send(messages.Archive{
			Path:     archive.ImportPath,
			Hash:     hash,
			Standard: standard,
		})
		return nil
	}

	if cachedPrelude, exists := cache["prelude"]; !exists || cachedPrelude != std.Prelude[min] {
		// send the prelude first if it's not in the cache
		c.send(messages.Archive{
			Path:     "prelude",
			Hash:     std.Prelude[min],
			Standard: true,
		})
	}

	// All programs need runtime and it's dependencies
	if _, _, err := b.BuildImportPath(ctx, "runtime"); err != nil {
		return err
	}

	for path := range source {
		if _, _, err := b.BuildImportPath(ctx, path); err != nil {
			return err
		}
	}

	c.send(index)

	c.send(messages.Updating{Done: true})

	return nil
}

func StripArchive(a *compiler.Archive) *compiler.Archive {
	out := &compiler.Archive{
		ImportPath: a.ImportPath,
		Name:       a.Name,
		Imports:    a.Imports,
		ExportData: a.ExportData,
		Minified:   a.Minified,
	}
	for _, d := range a.Declarations {
		// All that's needed in Declarations is FullName (https://github.com/gopherjs/gopherjs/blob/423bf76ba1888a53d4fe3c1a82991cdb019a52ad/compiler/package.go#L187-L191)
		out.Declarations = append(out.Declarations, &compiler.Decl{FullName: d.FullName, Blocking: d.Blocking})
	}
	return out
}

type updateWriter struct {
	send func(messages.Message)
}

func (w updateWriter) Write(b []byte) (n int, err error) {
	w.send(messages.Updating{Message: strings.TrimSuffix(string(b), "\n")})
	return len(b), nil
}
