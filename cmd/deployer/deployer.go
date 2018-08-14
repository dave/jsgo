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
	"strings"
	"text/template"

	"github.com/dave/jsgo/assets/std"
	"github.com/dave/jsgo/cmd/cmdconfig"
	"github.com/dave/jsgo/config"
	"github.com/dave/jsgo/server/servermsg"
	"github.com/dave/jsgo/server/wasm/messages"
	"github.com/dave/patsy"
	"github.com/dave/patsy/vos"
	"github.com/dave/services/constor/constormsg"
	"github.com/gorilla/websocket"
	"github.com/pkg/browser"
)

const CLIENT_VERSION = "1.0.0"

func Start(cfg *cmdconfig.Config) error {

	var debug io.Writer
	if cfg.Verbose {
		debug = os.Stdout
	} else {
		debug = ioutil.Discard
	}

	// create a temp dir
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	fpath := filepath.Join(dir, "out.wasm")

	args := []string{"build", "-o", fpath}

	extraFlags := strings.Fields(cfg.Flags)
	for _, f := range extraFlags {
		args = append(args, f)
	}

	path := "."
	if cfg.Path != "" {
		path = cfg.Path
	}
	args = append(args, path)

	fmt.Fprintln(debug, "Compiling...")

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
	wasmUrl := fmt.Sprintf("%s://%s/%x.wasm", config.Protocol[config.Pkg], config.Host[config.Pkg], binarySha.Sum(nil))
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
	loaderUrl := fmt.Sprintf("%s://%s/%x.js", config.Protocol[config.Pkg], config.Host[config.Pkg], loaderSha.Sum(nil))
	indexVars := struct{ Script, Loader string }{
		Script: fmt.Sprintf("%s://%s/wasm_exec.%s.js", config.Protocol[config.Pkg], config.Host[config.Pkg], std.Wasm[true]),
		Loader: loaderUrl,
	}
	indexTemplate := defaultIndexTemplate
	if cfg.Index != "" {
		indexFilename := cfg.Index
		if cfg.Path != "" {
			dir, err := patsy.Dir(vos.Os(), cfg.Path)
			if err != nil {
				return err
			}
			indexFilename = filepath.Join(dir, cfg.Index)
		}
		indexTemplateBytes, err := ioutil.ReadFile(indexFilename)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			indexTemplate, err = template.New("main").Parse(string(indexTemplateBytes))
			if err != nil {
				return err
			}
		}
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
		Version: CLIENT_VERSION,
		Files: []messages.DeployFileKey{
			files[messages.DeployFileTypeWasm].DeployFileKey,
			files[messages.DeployFileTypeIndex].DeployFileKey,
			files[messages.DeployFileTypeLoader].DeployFileKey,
		},
	}

	fmt.Fprintln(debug, "Querying server...")

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
			done = true
		case servermsg.Queueing:
			// don't print
		case servermsg.Error:
			return errors.New(message.Message)
		case messages.DeployClientVersionNotSupported:
			return errors.New("this client version is not supported - try `go get -u github.com/dave/jsgo`")
		default:
			// unexpected
			fmt.Fprintf(debug, "Unexpected message from server: %#v\n", message)
		}
	}

	if len(response.Required) > 0 {

		fmt.Fprintf(debug, "Files required: %d.\n", len(response.Required))
		fmt.Fprintln(debug, "Bundling required files...")

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

		fmt.Fprintf(debug, "Sending payload: %dKB.\n", len(payloadBytes)/1024)

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
			case servermsg.Error:
				return errors.New(message.Message)
			case constormsg.Storing:
				if message.Remain > 0 || message.Finished > 0 {
					fmt.Fprintf(debug, "Storing, %d to go.\n", message.Remain)
				}
			default:
				// unexpected
				fmt.Fprintf(debug, "Unexpected message from server: %#v\n", message)
			}
		}

		fmt.Fprintln(debug, "Sending done.")

	} else {
		fmt.Fprintln(debug, "No files required.")
	}

	outputVars := struct {
		Page   string
		Loader string
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

	if cfg.Open {
		browser.OpenURL(indexUrl)
	}

	return nil
}
