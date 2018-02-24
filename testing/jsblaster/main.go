package main

import (
	"context"
	"errors"
	"log"

	"fmt"

	"os"

	"encoding/json"

	"strings"

	"github.com/dave/blast/blaster"
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

	ws, err := websocket.Dial(fmt.Sprintf("wss://compile.jsgo.io/_ws/%s", p.Path), "", "https://compile.jsgo.io")
	//ws, err := websocket.Dial(fmt.Sprintf("ws://localhost:8080/_ws/%s", p.Path), "", "http://localhost:8080")
	if err != nil {
		return map[string]interface{}{"status": "error dialing: " + err.Error()}, err
	}
	defer ws.Close()

	read := make(chan error)
	go func() {
		defer close(read)
		for {

			send := func(err error) {
				select {
				case read <- err:
					// great!
				case <-ctx.Done():
					return
				default:
					// continue
				}
			}

			var raw string
			var msg struct {
				Type    string `json:"type"`
				Payload struct {
					Message string `json:"message"`
				} `json:"payload"`
				Err error
			}
			if err := websocket.Message.Receive(ws, &raw); err != nil {
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
			if msg.Type == "error" {
				if strings.Contains(msg.Payload.Message, "too many git objects") {
					send(errors.New("too many git objects"))
				} else {
					send(errors.New(msg.Payload.Message))
				}
				ws.Close()
				return
			}
			if msg.Type == "complete" {
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
