package main

import (
	"context"
	"errors"
	"log"

	"fmt"

	"os"

	"encoding/json"

	"strings"

	"io"

	"github.com/dave/blast/blaster"
	"github.com/dave/jsgo/server/messages"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/net/websocket"
)

// Set debug to true to dump full stack info on every error.
const debug = true

func main() {

	ctx, cancel := context.WithCancel(context.Background())

	b := blaster.New(ctx, cancel)
	defer b.Exit()

	b.RegisterWorkerType("jsgo", New)

	if err := b.Command(ctx); err != nil {
		if debug {
			log.Fatal(fmt.Printf("%+v", err))
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

// New returns a new dummy worker
func New() blaster.Worker {
	// notest
	return &Worker{}
}

// Worker is the worker type
type Worker struct{}

// Worker is the worker type
type Payload struct {
	// Path to compile
	Path string `mapstructure:"path"`
}

// Send satisfies the blaster.Worker interface
func (w *Worker) Send(ctx context.Context, raw map[string]interface{}) (out map[string]interface{}, err error) {

	var p Payload
	if err := mapstructure.Decode(raw, &p); err != nil {
		return map[string]interface{}{"status": "error decoding payload: " + err.Error()}, err
	}

	ws, err := websocket.Dial("wss://compile.jsgo.io/_pg/", "", "https://compile.jsgo.io")
	//ws, err := websocket.Dial("ws://localhost:8081/_pg/", "", "http://localhost:8080")
	if err != nil {
		return map[string]interface{}{"status": "error dialing: " + err.Error()}, err
	}
	defer ws.Close()

	b, _ := messages.Marshal(messages.Initialise{Path: p.Path, Minify: true})
	if err != nil {
		return map[string]interface{}{"status": "error encoding: " + err.Error()}, err
	}
	if _, err := ws.Write(b); err != nil {
		return map[string]interface{}{"status": "error writing: " + err.Error()}, err
	}
	//fmt.Println(string(b))

	read := make(chan error)
	go func() {
		defer close(read)
		for {

			send := func(err error) {
				//	if err != nil {
				//		fmt.Println(raw, err)
				//	}
				select {
				case read <- err:
					// great!
				case <-ctx.Done():
					select {
					case read <- ctx.Err():
						// great!
					default:
						// continue
					}
					return
				default:
					// continue
				}
			}

			var raw string
			var msg struct {
				Type    string
				Message struct {
					Message string
					Done    bool
				}
			}
			if err := websocket.Message.Receive(ws, &raw); err != nil {
				if err == io.EOF {
					send(nil)
					ws.Close()
					return
				}
				send(err)
				ws.Close()
				return
			}
			//fmt.Println(raw)

			if err := json.Unmarshal([]byte(raw), &msg); err != nil {
				send(err)
				ws.Close()
				return
			}
			if msg.Type == "Error" {
				if strings.Contains(msg.Message.Message, "too many git objects") {
					send(errors.New("too many git objects"))
				} else if strings.Contains(msg.Message.Message, "unrecognized import path") {
					send(errors.New("source"))
				} else if strings.Contains(msg.Message.Message, "no Go files in") {
					send(errors.New("source"))
				} else if strings.Contains(msg.Message.Message, "cannot find package") {
					send(errors.New("source"))
				} else if strings.HasPrefix(msg.Message.Message, "gopath/src/") {
					send(errors.New("source"))
				} else if strings.Contains(msg.Message.Message, "bad status") {
					send(errors.New("bad status"))
				} else {
					if msg.Message.Message == "" {
						send(errors.New(raw))
					} else {
						send(errors.New(msg.Message.Message))
					}
				}
				ws.Close()
				return
			}
			if msg.Type == "Updating" && msg.Message.Done {
				send(nil)
				ws.Close()
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
				// continue
			}
		}
	}()

	select {
	case err := <-read:
		if err != nil {
			return map[string]interface{}{"status": err.Error()}, err
		}
		return map[string]interface{}{"status": "success"}, err
		// Dummy worker - success!
	case <-ctx.Done():
		// Dummy worker - interrupted by context
		err := ctx.Err()
		var status string
		switch err {
		case nil:
			status = "unknown"
			err = errors.New("context done")
		case context.DeadlineExceeded:
			status = "timeout"
		case context.Canceled:
			status = "cancelled"
		default:
			status = err.Error()
		}
		return map[string]interface{}{"status": status}, err
	}
}
