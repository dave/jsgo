package frizz

import (
	"context"
	"net/http"

	"crypto/sha1"
	"encoding/json"

	"fmt"

	"bytes"
	"io"

	"github.com/dave/frizz/models"
	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/frizz/messages"
	"github.com/dave/services"
	"github.com/dave/services/constor"
	"github.com/dave/services/getter/get"
	"github.com/dave/services/getter/gettermsg"
	"github.com/dave/services/session"
)

func (h *Handler) Source(ctx context.Context, info messages.GetSource, req *http.Request, send func(message services.Message), receive chan services.Message) error {

	storer := constor.New(ctx, h.Fileserver, send, config.ConcurrentStorageUploads)
	defer storer.Close()

	// Send a message to the client that downloading step has started.
	send(gettermsg.Downloading{Starting: true})

	var save bool
	if len(info.Tags) == 0 {
		// only save the getter hints if there's no special build tags (this may affect dependency tree)
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
	index := messages.SourceIndex{}

	// Start the download process - just like the "go get" command.
	g := get.New(s, send, gitreq)

	g.Callback = func(path string, files map[string]string, standard bool) error {
		if done[path] {
			return nil
		}
		done[path] = true

		var hash string
		var unchanged bool

		// internal/cpu not in std.Source? Treat as a non-standard package
		_, foundInStandard := std.Source[path]

		if standard && foundInStandard {
			hash = std.Source[path]
			if cached, ok := info.Cache[path]; ok && cached == hash {
				unchanged = true
			}
			if !unchanged {
				send(messages.Source{
					Path:     path,
					Hash:     hash,
					Standard: standard,
				})
			}
		} else {
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
			if cached, ok := info.Cache[path]; ok && cached == hash {
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
							Standard: standard,
						})
					},
				})
			}
		}

		index[path] = messages.SourceIndexItem{
			Hash:      hash,
			Unchanged: unchanged,
		}

		return nil
	}

	if err := g.Get(ctx, info.Path, false, insecure, false); err != nil {
		return err
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
