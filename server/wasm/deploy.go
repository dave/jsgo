package wasm

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/wasm/messages"
	"github.com/dave/services"
)

func (h *Handler) DeployQuery(ctx context.Context, info messages.DeployQuery, req *http.Request, send func(services.Message), receive chan services.Message) error {

	var m sync.Mutex
	var required []messages.DeployFileKey
	wg := &sync.WaitGroup{}

	for _, file := range info.Files {
		var outer error
		wg.Add(1)
		go func() {
			defer wg.Done()
			bucket, name := details(file.Type, file.Hash)
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

	return nil
}

func details(typ messages.DeployFileType, hash string) (bucket, name string) {
	switch typ {
	case messages.DeployFileTypeIndex:
		bucket = config.Bucket[config.Index]
		name = hash
	case messages.DeployFileTypeLoader:
		bucket = config.Bucket[config.Pkg]
		name = fmt.Sprintf("%s.js", hash)
	case messages.DeployFileTypeWasm:
		bucket = config.Bucket[config.Pkg]
		name = fmt.Sprintf("%s.wasm", hash)
	}
	return
}
