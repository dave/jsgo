package stores

import (
	"github.com/dave/flux"
	"github.com/gopherjs/gopherjs/js"
)

type App struct {
	Dispatcher flux.DispatcherInterface
	Watcher    flux.WatcherInterface
	Notifier   flux.NotifierInterface

	Archive    *ArchiveStore
	Editor     *EditorStore
	Connection *ConnectionStore
	Local      *LocalStore
	Scanner    *ScannerStore
}

func (a *App) Init() {

	n := flux.NewNotifier()
	a.Notifier = n
	a.Watcher = n

	a.Archive = NewArchiveStore(a)
	a.Editor = NewEditorStore(a)
	a.Connection = NewConnectionStore(a)
	a.Local = NewLocalStore(a)
	a.Scanner = NewScannerStore(a)

	a.Dispatcher = flux.NewDispatcher(
		// Notifier:
		a.Notifier,
		// Stores:
		a.Archive,
		a.Editor,
		a.Connection,
		a.Local,
		a.Scanner,
	)
}

func (a *App) Dispatch(action flux.ActionInterface) chan struct{} {
	return a.Dispatcher.Dispatch(action)
}

func (a *App) Watch(key interface{}, f func(done chan struct{})) {
	a.Watcher.Watch(key, f)
}

func (a *App) Delete(key interface{}) {
	a.Watcher.Delete(key)
}

func (a *App) Fail(err error) {
	// TODO: improve this
	js.Global.Call("alert", err.Error())
}
