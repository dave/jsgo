package wasm

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/wasm/messages"
	"github.com/dave/services"
	"github.com/dave/services/constor"
	"github.com/dave/services/constor/constormsg"
)

func (h *Handler) DeployQuery(ctx context.Context, info messages.DeployQuery, req *http.Request, send func(services.Message), receive chan services.Message) error {

	var m sync.Mutex
	var required []messages.DeployFileKey
	wg := &sync.WaitGroup{}

	for _, file := range info.Files {
		file := file
		var outer error
		wg.Add(1)
		go func() {
			defer wg.Done()
			bucket, name, _ := details(file.Type, file.Hash)
			exists, err := h.Fileserver.Exists(ctx, bucket, name)
			if err != nil {
				outer = err
				return
			}
			if !exists {
				m.Lock()
				required = append(required, file)
				m.Unlock()
			}
		}()
	}
	wg.Wait()

	send(messages.DeployQueryResponse{Required: required})

	var payload messages.DeployPayload
	select {
	case message := <-receive:
		payload = message.(messages.DeployPayload)
	case <-ctx.Done():
		return nil
	}

	storer := constor.New(ctx, h.Fileserver, send, config.ConcurrentStorageUploads)
	defer storer.Close()

	for _, f := range payload.Files {
		// check the hash is correct
		sha := sha1.New()
		if _, err := io.Copy(sha, bytes.NewBuffer(f.Contents)); err != nil {
			return err
		}
		calculated := fmt.Sprintf("%x", sha.Sum(nil))
		if calculated != f.Hash {
			return fmt.Errorf("hash not consistent for %s", f.Type)
		}
		bucket, name, mime := details(f.Type, f.Hash)
		storer.Add(constor.Item{
			Message:   string(f.Type),
			Name:      name,
			Contents:  f.Contents,
			Bucket:    bucket,
			Mime:      mime,
			Count:     true,
			Immutable: true,
			Send:      true,
		})
	}

	if err := storer.Wait(); err != nil {
		return err
	}

	send(constormsg.Storing{Done: true})

	send(messages.DeployDone{})

	return nil
}

func details(typ messages.DeployFileType, hash string) (bucket, name, mime string) {
	switch typ {
	case messages.DeployFileTypeIndex:
		bucket = config.Bucket[config.Index]
		name = hash
		mime = constor.MimeHtml
	case messages.DeployFileTypeLoader:
		bucket = config.Bucket[config.Pkg]
		name = fmt.Sprintf("%s.js", hash)
		mime = constor.MimeJs
	case messages.DeployFileTypeWasm:
		bucket = config.Bucket[config.Pkg]
		name = fmt.Sprintf("%s.wasm", hash)
		mime = constor.MimeWasm
	}
	return
}
