# getter

The getter package is a fork of [cmd/go/internal/get](https://github.com/golang/go/tree/master/src/cmd/go/internal/get) 
and associated packages. Getter adds modifications in order to pull using the [go-git](https://github.com/src-d/go-git)
Git client into a virtual filesystem. Additionally, various features that are irrelevant to GopherJS 
compiles have been dropped (e.g. tracking C files).