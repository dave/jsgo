package server

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"time"

	"cloud.google.com/go/storage"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/compile"
	"github.com/dave/jsgo/server/messages"
	"github.com/dave/jsgo/server/store"
)

func (h *Handler) playShare(ctx context.Context, info messages.Share, req *http.Request, send func(message messages.Message), receive chan messages.Message) error {

	send(messages.Storing{Starting: true})

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

	storer := compile.NewStorer(ctx, h.Fileserver, send, config.ConcurrentStorageUploads)
	storer.AddSrc("source", fmt.Sprintf("%x.json", hash), buf.Bytes())
	storer.Wait()

	send(messages.Storing{Done: true})

	if err := storeShare(ctx, info.Source, fmt.Sprintf("%x", hash), send, req); err != nil {
		return err
	}

	send(messages.ShareComplete{Hash: fmt.Sprintf("%x", hash)})

	return nil
}

func storeShare(ctx context.Context, source map[string]map[string]string, hash string, send func(messages.Message), req *http.Request) error {
	var count int
	for _, pkg := range source {
		for range pkg {
			count++
		}
	}
	data := store.ShareData{
		Time:  time.Now(),
		Ip:    req.Header.Get("X-Forwarded-For"),
		Files: count,
		Hash:  hash,
	}
	if err := store.StoreShare(ctx, data); err != nil {
		return err
	}
	return nil
}
