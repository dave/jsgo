package store

import (
	"github.com/dave/jsgo/playground/connection"
	"github.com/dave/jsgo/playground/dispatcher"
	"github.com/dave/locstor"
)

// Listeners is the listeners that will be invoked when the store changes.
var Listeners = dispatcher.NewListenerRegistry()

var (
	stor = locstor.NewDataStore(locstor.JSONEncoding)
	conn = connection.New()
)

func init() {
	go func() {
		for message := range conn.Receive {
			dispatcher.Dispatch(message)
		}
	}()
}
