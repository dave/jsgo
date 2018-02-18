package main

import (
	"context"
	"errors"
	"log"

	"fmt"

	"os"

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
type Worker struct {
	// Path to compile
	Path string `mapstructure:"path"`
}

// Start satisfies the blaster.Starter interface
func (w *Worker) Start(ctx context.Context, raw map[string]interface{}) error {

	if err := mapstructure.Decode(raw, w); err != nil {
		return err
	}

	return nil
}

// Send satisfies the blaster.Worker interface
func (w *Worker) Send(ctx context.Context, raw map[string]interface{}) (out map[string]interface{}, err error) {

	ws, err := websocket.Dial(fmt.Sprintf("wss://compile.jsgo.io/_ws%s", w.Path), "", "https://compile.jsgo.io")
	if err != nil {
		return map[string]interface{}{"status": "error dialing: " + err.Error()}, err
	}
	defer ws.Close()

	read := make(chan error)
	go func() {
		defer close(read)
		for {
			var msg struct {
				Type    string `json:"type"`
				Payload struct {
					Message string `json:"message"`
				} `json:"payload"`
				Err error
			}
			if err := websocket.JSON.Receive(ws, &msg); err != nil {
				read <- err
				ws.Close()
				return
			}
			if msg.Type == "error" {
				read <- errors.New(msg.Payload.Message)
				ws.Close()
				return
			}
			if msg.Type == "complete" {
				read <- nil
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
