package main

import (
	"time"

	"fmt"

	"strings"

	"github.com/dave/jsgo/playground/splitter"
	"github.com/dave/jsgo/server/messages"
	"github.com/go-humble/locstor"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket/websocketjs"
	"github.com/tulir/gopher-ace"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func main() {
	if document.ReadyState() == "loading" {
		document.AddEventListener("DOMContentLoaded", false, func(dom.Event) {
			go run()
		})
	} else {
		go run()
	}
}

func run() {

	applyStyles()

	body := document.GetElementsByTagName("body")[0].(*dom.HTMLBodyElement)
	holder := document.CreateElement("div").(*dom.HTMLDivElement)
	left := document.CreateElement("div").(*dom.HTMLDivElement)
	right := document.CreateElement("div").(*dom.HTMLDivElement)
	body.AppendChild(holder)
	holder.AppendChild(left)
	holder.AppendChild(right)
	holder.Class().Add("split")
	holder.Class().Add("split-horizontal")
	left.Class().Add("split")
	right.Class().Add("split")

	split := splitter.New("split")
	split.Init(
		js.S{left.Underlying(), right.Underlying()},
		js.M{
			"sizes":     split.GetSavedSized([]int{75, 25}),
			"onDragEnd": split.SaveSizes,
		},
	)

	header := document.CreateElement("div").(*dom.HTMLDivElement)
	header.Class().Add("header")
	left.AppendChild(header)

	edit := document.CreateElement("div")
	edit.Class().Add("editor")
	left.AppendChild(edit)
	editor := ace.EditDOM(edit)
	editor.SetOptions(map[string]interface{}{
		"mode": "ace/mode/golang",
	})

	button := document.CreateElement("button").(*dom.HTMLButtonElement)
	button.SetInnerHTML("Compile")
	button.Class().Add("btn")
	button.Class().Add("btn-sm")
	button.Class().Add("btn-primary")
	header.AppendChild(button)
	button.AddEventListener("click", false, func(event dom.Event) {
		event.PreventDefault()
		compile(right, editor.GetValue())
	})

	value, err := locstor.GetItem("code")
	if _, isNotFound := err.(locstor.ItemNotFoundError); err != nil {

		if !isNotFound {
			panic(err)
		}

		defaultCode := `package main

import (
	"log"

	"github.com/dave/ebiten"
	"github.com/dave/ebiten/examples/2048/2048"
)

var (
	game *twenty48.Game
)

func update(screen *ebiten.Image) error {
	if err := game.Update(); err != nil {
		return err
	}
	if ebiten.IsRunningSlowly() {
		return nil
	}
	game.Draw(screen)
	return nil
}

func main() {
	var err error
	game, err = twenty48.NewGame()
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.Run(update, twenty48.ScreenWidth, twenty48.ScreenHeight, 1, "2048 (Ebiten Demo)"); err != nil {
		log.Fatal(err)
	}
}`

		editor.SetValue(defaultCode)
		editor.ClearSelection()
		editor.MoveCursorTo(0, 0)
	}
	if value != "" {
		editor.SetValue(value)
		editor.ClearSelection()
		editor.MoveCursorTo(0, 0)
	}

	var changes int
	editor.OnChange(func(e *js.Object) {
		changes++
		before := changes
		go func() {
			<-time.After(time.Millisecond * 250)
			if before == changes {
				if err := locstor.SetItem("code", editor.GetValue()); err != nil {
					panic(err)
				}
			}
		}()
	})
}

func compile(pane *dom.HTMLDivElement, code string) {
	pane.SetInnerHTML("")
	pre := document.CreateElement("pre")
	pane.AppendChild(pre)

	msg := func(m string) {
		m = strings.TrimSuffix(m, "\n")
		pre.SetInnerHTML(m + "\n" + pre.InnerHTML())
	}
	msgf := func(m string, args ...interface{}) {
		m = strings.TrimSuffix(m, "\n")
		pre.SetInnerHTML(fmt.Sprintf(m, args...) + "\n" + pre.InnerHTML())
	}

	go func() {
		msg("Storing gist...")
		id, err := store(code)
		if err != nil {
			panic(err)
		}
		//id := "d60c1d31cbc3347f0c8485f954bc2f93"
		msg("Gist created: gist.github.com/" + id)

		ws, err := websocketjs.New("wss://compile.jsgo.io/_ws/gist.github.com/" + id)
		//ws, err := websocketjs.New("ws://localhost:8081/_ws/gist.github.com/" + id)
		if err != nil {
			panic(err)
		}
		ws.AddEventListener("open", false, func(ev *js.Object) {
			msg("Compiling...")
		})
		ws.AddEventListener("message", false, func(ev *js.Object) {
			_, p, err := messages.Parse([]byte(ev.Get("data").String()))
			if err != nil {
				panic(err)
			}
			switch p := p.(type) {
			case messages.DownloadPayload:
				if !p.Done && !p.Starting {
					msgf("downloading %s", p.Message)
				}
			case messages.CompilePayload:
				if !p.Done && !p.Starting {
					msgf("compiling %s", p.Message)
				}
			case messages.QueuePayload:
				if !p.Done {
					msgf("queued at position %d", p.Position)
				}
			case messages.ErrorPayload:
				panic(fmt.Sprintf("error: %s %s", p.Message, p.Path))
			case messages.StorePayload:
				if !p.Starting && !p.Done {
					msgf("storing: %d finished, %d unchanged, %d remain", p.Finished, p.Unchanged, p.Remain)
				}
			case messages.CompletePayload:
				msg("complete!")
				iframe := document.CreateElement("iframe").(*dom.HTMLIFrameElement)
				iframe.Src = "https://jsgo.io/gist.github.com/" + id
				iframe.Class().Add("preview")
				pane.SetInnerHTML("")
				pane.AppendChild(iframe)
				ws.Close()
			}
		})
		ws.AddEventListener("close", false, func(ev *js.Object) {
			msg("closed...")
		})
		ws.AddEventListener("error", false, func(ev *js.Object) {
			msg("error...")
		})

	}()

	/**
	repos, _, err := client.Repositories.List(ctx, "", nil)
	if _, ok := err.(*github.RateLimitError); ok {
		log.Println("hit rate limit")
	}
	*/
}
