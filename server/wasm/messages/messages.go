package messages

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"

	"github.com/dave/jsgo/server/servermsg"
	"github.com/dave/services"
	"github.com/dave/services/constor/constormsg"
	"github.com/gorilla/websocket"
)

type Payload struct {
	Message services.Message
}

func init() {

	// Commands:
	gob.Register(DeployQuery{})

	// Data messages:
	gob.Register(DeployQueryResponse{})
	gob.Register(DeployFileKey{})
	gob.Register(DeployFile{})
	gob.Register(DeployPayload{})
	gob.Register(DeployDone{})
	gob.Register(DeployClientVersionNotSupported{})

	// Initialise types in servermsg
	servermsg.RegisterTypes()

	// Initialise types in constormsg
	constormsg.RegisterTypes()
}

// Client sends a DeployQuery with all offered files.
// Server responds with DeployQueryResponse, with all required files listed.
// Client sends a DeployFile for each required file.

type DeployQuery struct {
	Version string
	Files   []DeployFileKey
}

type DeployQueryResponse struct {
	Required []DeployFileKey
}

type DeployPayload struct {
	Files []DeployFile
}

type DeployClientVersionNotSupported struct{}

type DeployFileKey struct {
	Type DeployFileType
	Hash string // sha1 hash of contents
}

type DeployFile struct {
	DeployFileKey
	Contents []byte // in the initial CommandDeploy, this is nil
}

type DeployDone struct{}

type DeployFileType string

const (
	DeployFileTypeIndex  DeployFileType = "index"
	DeployFileTypeLoader                = "loader"
	DeployFileTypeWasm                  = "wasm"
)

func Marshal(in services.Message) ([]byte, int, error) {
	p := Payload{in}
	buf := &bytes.Buffer{}
	gzw := gzip.NewWriter(buf)
	if err := gob.NewEncoder(gzw).Encode(p); err != nil {
		return nil, 0, err
	}
	if err := gzw.Close(); err != nil {
		return nil, 0, err
	}
	return buf.Bytes(), websocket.BinaryMessage, nil
}

func Unmarshal(in []byte) (services.Message, error) {
	var p Payload
	gzr, err := gzip.NewReader(bytes.NewBuffer(in))
	if err != nil {
		return nil, err
	}
	if err := gob.NewDecoder(gzr).Decode(&p); err != nil {
		return nil, err
	}
	if err := gzr.Close(); err != nil {
		return nil, err
	}
	return p.Message, nil
}
