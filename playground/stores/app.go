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
	Compile    *CompileStore
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
	a.Compile = NewCompileStore(a)

	a.Dispatcher = flux.NewDispatcher(
		// Notifier:
		a.Notifier,
		// Stores:
		a.Archive,
		a.Editor,
		a.Connection,
		a.Local,
		a.Scanner,
		a.Compile,
	)
}

func (a *App) Dispatch(action flux.ActionInterface) chan struct{} {
	//js.Global.Get("console").Call("log", fmt.Sprintf("%T", action))
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
