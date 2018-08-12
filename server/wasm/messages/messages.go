package messages

import (
	"bytes"
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

	// Initialise types in servermsg
	servermsg.RegisterTypes()

	// Initialise types in constormsg
	constormsg.RegisterTypes()
}

// Client sends a DeployQuery with all offered files.
// Server responds with DeployQueryResponse, with all required files listed.
// Client sends a DeployFile for each required file.

type DeployQuery struct {
	Files []DeployFileKey
}

type DeployQueryResponse struct {
	Required []DeployFileKey
}

type DeployFileKey struct {
	Type DeployFileType
	Hash string // sha1 hash of contents
}

type DeployFile struct {
	DeployFileKey
	Contents []byte // in the initial CommandDeploy, this is nil
}

type DeployFileType string

const (
	DeployFileTypeIndex  DeployFileType = "index"
	DeployFileTypeLoader                = "loader"
	DeployFileTypeWasm                  = "wasm"
)

func Marshal(in services.Message) ([]byte, int, error) {
	p := Payload{in}
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(p); err != nil {
		return nil, 0, err
	}
	return buf.Bytes(), websocket.BinaryMessage, nil
}

func Unmarshal(in []byte) (services.Message, error) {
	var p Payload
	if err := gob.NewDecoder(bytes.NewBuffer(in)).Decode(&p); err != nil {
		return nil, err
	}
	return p.Message, nil
}
