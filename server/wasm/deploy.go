package wasm

import (
	"context"
	"net/http"

	"github.com/dave/jsgo/server/wasm/messages"
	"github.com/dave/services"
)

func (h *Handler) DeployQuery(ctx context.Context, info messages.DeployQuery, req *http.Request, send func(services.Message), receive chan services.Message) error {

	// TODO
	//h.Fileserver.Exists()

	/*

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
		output, err := deployer.New(s, send, std.Index, std.Prelude, config.DeployerConfig).Deploy(ctx, path, deployer.PathIndex, map[bool]bool{true: true, false: true})
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

	*/
	return nil
}
