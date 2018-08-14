package deployer

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/cmd/cmdconfig"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/servermsg"
	"github.com/dave/jsgo/server/wasm/messages"
	"github.com/gorilla/websocket"
)

func Start(cfg *cmdconfig.Config) error {

	// create a temp dir
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	fpath := filepath.Join(dir, "out.wasm")

	args := []string{"build", "-o", fpath}

	// TODO: more args?

	path := "."
	if cfg.Path != "" {
		path = cfg.Path
	}
	args = append(args, path)

	cmd := exec.Command(cfg.Command, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOARCH=wasm")
	cmd.Env = append(cmd.Env, "GOOS=js")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if len(output) > 0 {
		return fmt.Errorf("%s", string(output))
	}

	binaryBytes, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	binarySha := sha1.New()
	if _, err := io.Copy(binarySha, bytes.NewBuffer(binaryBytes)); err != nil {
		return err
	}

	files := map[messages.DeployFileType]messages.DeployFile{}

	files[messages.DeployFileTypeWasm] = messages.DeployFile{
		DeployFileKey: messages.DeployFileKey{
			Type: messages.DeployFileTypeWasm,
			Hash: fmt.Sprintf("%x", binarySha.Sum(nil)),
		},
		Contents: binaryBytes,
	}

	loaderBuf := &bytes.Buffer{}
	loaderSha := sha1.New()
	wasmUrl := fmt.Sprintf("%s://%s/%x.wasm", config.Protocol[config.Pkg], config.Host[config.Pkg], fmt.Sprintf("%x", binarySha.Sum(nil)))
	loaderVars := struct{ Binary string }{
		Binary: wasmUrl,
	}
	if err := loaderTemplateMin.Execute(io.MultiWriter(loaderBuf, loaderSha), loaderVars); err != nil {
		return err
	}

	files[messages.DeployFileTypeLoader] = messages.DeployFile{
		DeployFileKey: messages.DeployFileKey{
			Type: messages.DeployFileTypeLoader,
			Hash: fmt.Sprintf("%x", loaderSha.Sum(nil)),
		},
		Contents: loaderBuf.Bytes(),
	}

	indexBuf := &bytes.Buffer{}
	indexSha := sha1.New()
	loaderUrl := fmt.Sprintf("%s://%s/%s.js", config.Protocol[config.Pkg], config.Host[config.Pkg], std.Wasm[true])
	indexVars := struct{ Script, Loader string }{
		Script: fmt.Sprintf("%s://%s/wasm_exec.%s.js", config.Protocol[config.Pkg], config.Host[config.Pkg], std.Wasm[true]),
		Loader: loaderUrl,
	}
	if err := indexTemplate.Execute(io.MultiWriter(indexBuf, indexSha), indexVars); err != nil {
		return err
	}

	files[messages.DeployFileTypeIndex] = messages.DeployFile{
		DeployFileKey: messages.DeployFileKey{
			Type: messages.DeployFileTypeIndex,
			Hash: fmt.Sprintf("%x", indexSha.Sum(nil)),
		},
		Contents: indexBuf.Bytes(),
	}

	indexUrl := fmt.Sprintf("%s://%s/%x", config.Protocol[config.Index], config.Host[config.Index], indexSha.Sum(nil))

	message := messages.DeployQuery{
		Files: []messages.DeployFileKey{
			files[messages.DeployFileTypeWasm].DeployFileKey,
			files[messages.DeployFileTypeIndex].DeployFileKey,
			files[messages.DeployFileTypeLoader].DeployFileKey,
		},
	}

	protocol := "wss"
	if config.Protocol[config.Wasm] == "http" {
		protocol = "ws"
	}
	conn, _, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("%s://%s/_wasm/", protocol, config.Host[config.Wasm]),
		http.Header{"Origin": []string{fmt.Sprintf("%s://%s/", config.Protocol[config.Wasm], config.Host[config.Wasm])}},
	)
	if err != nil {
		return err
	}

	messageBytes, messageType, err := messages.Marshal(message)
	if err != nil {
		return err
	}
	if err := conn.WriteMessage(messageType, messageBytes); err != nil {
		return err
	}

	var response messages.DeployQueryResponse
	var done bool
	for !done {
		_, replyBytes, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		message, err := messages.Unmarshal(replyBytes)
		if err != nil {
			return err
		}
		switch message := message.(type) {
		case messages.DeployQueryResponse:
			response = message
			fmt.Printf("%#v\n", message)
			done = true
		case servermsg.Queueing:
			// don't print
		default:
			// unexpected
			fmt.Printf("%#v\n", message)
		}
	}

	var required []messages.DeployFile
	for _, k := range response.Required {
		file := files[k.Type]
		if file.Hash == "" {
			return errors.New("server requested file not found")
		}
		required = append(required, file)
	}

	payload := messages.DeployPayload{Files: required}
	payloadBytes, payloadType, err := messages.Marshal(payload)
	if err != nil {
		return err
	}
	if err := conn.WriteMessage(payloadType, payloadBytes); err != nil {
		return err
	}

	done = false
	for !done {
		_, replyBytes, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		message, err := messages.Unmarshal(replyBytes)
		if err != nil {
			return err
		}
		switch message := message.(type) {
		case messages.DeployDone:
			done = true
		case servermsg.Queueing:
			// don't print
		default:
			// storing messages
			fmt.Printf("%#v\n", message)
		}
	}

	outputVars := struct {
		Page    string
		Loader  string
		Error   bool
		Message string
	}{
		Page:   indexUrl,
		Loader: loaderUrl,
	}

	if cfg.Json {
		out, err := json.Marshal(outputVars)
		if err != nil {
			return err
		}
		fmt.Println(string(out))
	} else {
		tpl, err := template.New("main").Parse(cfg.Template)
		if err != nil {
			return err
		}
		tpl.Execute(os.Stdout, outputVars)
		fmt.Println("")
	}
	return nil
}
