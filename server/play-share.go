package server

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
)

func playShare(ctx context.Context, info messages.Share, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	send(messages.Storing{Starting: true})

	if config.UseLocal {
		// dummy for local dev
		send(messages.ShareComplete{Hash: "56f9ea337c5f39631fa095e789e44957344e498f"})
		return nil
	}

	buf := &bytes.Buffer{}
	sha := sha1.New()
	w := io.MultiWriter(buf, sha)
	if err := json.NewEncoder(w).Encode(info); err != nil {
		return err
	}
	hash := sha.Sum(nil)

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	storer := compile.NewStorer(ctx, client, send, config.ConcurrentStorageUploads)
	storer.AddSrc("source", fmt.Sprintf("%x.json", hash), buf.Bytes())
	storer.Wait()

	send(messages.Storing{Done: true})

	send(messages.ShareComplete{Hash: fmt.Sprintf("%x", hash)})

	return nil
}
