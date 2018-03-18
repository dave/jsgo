package stores

import (
	"fmt"

	"github.com/dave/flux"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
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

func (a *App) Debug(message ...interface{}) {
	js.Global.Get("console").Call("log", message...)
}

func (a *App) Log(message ...interface{}) {
	m := dom.GetWindow().Document().GetElementByID("message")
	if len(message) == 0 {
		m.SetInnerHTML("jsgo")
		return
	}
	s := fmt.Sprint("jsgo - ", message[0])
	if m.InnerHTML() != s {
		requestAnimationFrame()
		m.SetInnerHTML(s)
		requestAnimationFrame()
	}
	js.Global.Get("console").Call("log", message...)
}

func (a *App) Logf(format string, args ...interface{}) {
	a.Log(fmt.Sprintf(format, args...))
}

func requestAnimationFrame() {
	c := make(chan struct{})
	js.Global.Call("requestAnimationFrame", func() { close(c) })
	<-c
}
