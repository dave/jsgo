# getter

The getter package is partial fork of: 

* [cmd/go/internal/base](https://github.com/golang/go/tree/549cb18a9131221755694c0ccc610ae9a406129d/src/cmd/go/internal/base)
* [cmd/go/internal/get](https://github.com/golang/go/tree/549cb18a9131221755694c0ccc610ae9a406129d/src/cmd/go/internal/get)
* [cmd/go/internal/load](https://github.com/golang/go/tree/549cb18a9131221755694c0ccc610ae9a406129d/src/cmd/go/internal/load)
* [cmd/go/internal/str](https://github.com/golang/go/tree/549cb18a9131221755694c0ccc610ae9a406129d/src/cmd/go/internal/str)
* [cmd/go/internal/web](https://github.com/golang/go/tree/549cb18a9131221755694c0ccc610ae9a406129d/src/cmd/go/internal/web)

At [549cb18a9131221755694c0ccc610ae9a406129d](https://github.com/golang/go/commit/549cb18a9131221755694c0ccc610ae9a406129d) 
(Go 1.10). Getter adds modifications in order to pull using the [go-git](https://github.com/src-d/go-git) 
Git client into a virtual filesystem. Additionally, various features that are irrelevant to GopherJS 
compiles have been dropped (e.g. cgo, command line flags etc.)

I've included the original files from the Go repo as txt files so it's easy to diff them and see the 
changes.
