package jsgo

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dave/jsgo/assets"
	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/jsgo/messages"
	"github.com/dave/jsgo/server/servermsg"
	"github.com/dave/jsgo/server/store"
	"github.com/dave/services"
	"github.com/dave/services/deployer"
	"github.com/dave/services/getter/get"
	"github.com/dave/services/getter/gettermsg"
	"github.com/dave/services/session"
)

func (h *Handler) Compile(ctx context.Context, info messages.Compile, req *http.Request, send func(services.Message), receive chan services.Message) error {

	path := info.Path

	s := session.New(nil, assets.Assets, assets.Archives, h.Fileserver, config.ValidExtensions)

	// Send a message to the client that downloading step has started.
	send(gettermsg.Downloading{Starting: true})

	gitreq := h.Cache.NewRequest(true)
	if err := gitreq.InitialiseFromHints(ctx, path); err != nil {
		return err
	}

	// set insecure = true in local mode or it will fail if git repo has git protocol
	insecure := config.LOCAL

	// Start the download process - just like the "go get" command.
	if err := get.New(s, send, gitreq).Get(ctx, path, false, insecure, false); err != nil {
		return err
	}

	if err := gitreq.Close(ctx); err != nil {
		return err
	}

	// Send a message to the client that downloading step has finished.
	send(gettermsg.Downloading{Done: true})

	// Start the compile process - this compiles to JS and sends the files to a GCS bucket.
	output, err := deployer.New(s, send, std.Index, std.Prelude, deployerConfig).Deploy(ctx, path, deployer.PathIndex, map[bool]bool{true: true, false: true})
	if err != nil {
		return err
	}

	// Logs the success in the datastore
	h.storeCompile(ctx, send, path, req, output)

	// Send a message to the client that the process has successfully finished
	send(messages.Complete{
		Path:    path,
		Short:   strings.TrimPrefix(path, "github.com/"),
		HashMin: fmt.Sprintf("%x", output[true].MainHash),
		HashMax: fmt.Sprintf("%x", output[false].MainHash),
	})
	return nil
}

func (h *Handler) storeCompile(ctx context.Context, send func(services.Message), path string, req *http.Request, output map[bool]*deployer.DeployOutput) {
	data := store.CompileData{
		Path:    path,
		Time:    time.Now(),
		Min:     getCompileContents(output[true], true),
		Max:     getCompileContents(output[false], false),
		Ip:      req.Header.Get("X-Forwarded-For"),
		Success: true,
	}
	if err := store.StoreCompile(ctx, h.Database, path, data); err != nil {
		// don't save this one to the datastore because it's an error from the datastore.
		send(servermsg.Error{Message: err.Error()})
		return
	}
}

func getCompileContents(c *deployer.DeployOutput, min bool) store.CompileContents {
	val := store.CompileContents{}
	val.Main = fmt.Sprintf("%x", c.MainHash)
	preludeHash := std.Prelude[min]
	val.Packages = []store.CompilePackage{
		{
			Path:     "prelude",
			Hash:     preludeHash,
			Standard: true,
		},
	}
	for _, p := range c.Packages {
		val.Packages = append(val.Packages, store.CompilePackage{
			Path:     p.Path,
			Hash:     fmt.Sprintf("%x", p.Hash),
			Standard: p.Standard,
		})
	}
	return val
}

var deployerConfig = deployer.Config{
	ConcurrentStorageUploads: config.ConcurrentStorageUploads,
	IndexBucket:              config.IndexBucket,
	PkgBucket:                config.PkgBucket,
	Protocol:                 config.Protocol,
	PkgHost:                  config.PkgHost,
}
