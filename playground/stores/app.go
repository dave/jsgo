package stores

import (
	"fmt"
	"strconv"

	"honnef.co/go/js/dom"

	"strings"

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
	Share      *ShareStore
	Get        *GetStore
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
	a.Share = NewShareStore(a)
	a.Get = NewGetStore(a)

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
		a.Share,
		a.Get,
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

func (a *App) Log(args ...interface{}) {
	m := dom.GetWindow().Document().GetElementByID("message")
	var message string
	if len(args) > 0 {
		message = strings.TrimSuffix(fmt.Sprintln(args...), "\n")
	}
	if m.InnerHTML() != message {
		if message != "" {
			js.Global.Get("console").Call("log", "Status", strconv.Quote(message))
		}
		requestAnimationFrame()
		m.SetInnerHTML(message)
		requestAnimationFrame()
	}
}

func (a *App) Logf(format string, args ...interface{}) {
	a.Log(fmt.Sprintf(format, args...))
}

func requestAnimationFrame() {
	c := make(chan struct{})
	js.Global.Call("requestAnimationFrame", func() { close(c) })
	<-c
}
