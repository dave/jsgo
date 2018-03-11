package stores

import (
	"github.com/dave/flux"
	"github.com/gopherjs/gopherjs/js"
)

type App struct {
	Dispatcher flux.DispatcherInterface
	Watcher    flux.WatcherInterface
	Notifier   flux.NotifierInterface

	Compiler   *CompilerStore
	Editor     *EditorStore
	Connection *ConnectionStore
	Local      *LocalStore
}

func (a *App) Init() {
	n := flux.NewNotifier()
	a.Notifier = n
	a.Watcher = n
	a.Compiler = NewCompilerStore(a)
	a.Editor = NewEditorStore(a)
	a.Connection = NewConnectionStore(a)
	a.Local = NewLocalStore(a)
	a.Dispatcher = flux.NewDispatcher(
		// Notifier:
		a.Notifier,
		// Stores:
		a.Compiler,
		a.Editor,
		a.Connection,
		a.Local,
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
