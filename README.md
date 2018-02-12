# jsgo.io

jsgo.io is a GopherJS compiler, serving framework and CDN.

*Please don't use this in production just yet!* 

### Features

* Compiles your Go to JS.  
* Splits the JS up by Go package.  
* Stores the JS in a CDN for you (GCP / Cloudflare).  
* Aggressively caches the JS.  
* Creates a page on `jsgo.io` that runs your JS.  
* Creates a single bootstrap JS file you can use on your site.

### How it works

Visit `https://compile.jsgo.io/<path>` to compile or recompile your code. If your Go package path starts 
`github.com`, you can omit that.

Here's a very simple [hello world](https://github.com/dave/jstest/blob/master/main.go): Open [compile.jsgo.io/dave/jstest](https://compile.jsgo.io/dave/jstest)
and click `Compile`. 

After it's finished, you'll be shown a link to the page: [jsgo.io/dave/jstest](https://jsgo.io/dave/jstest). 

If you look at the `Network` panel in your browser inspector as the page loads, you'll see all the packages
in the dependency tree downloading separately. The individual packages are aggressively cached, so if 
you recently visited a page that uses the `fmt` package, it won't be downloaded again!  

You'll also see the URL of a JS file, for use on your own site. This small bootstrap loader initiates 
the individual package files to download, and runs the JS.

### Demos

* https://jsgo.io/dave/todomvc
* https://jsgo.io/gopherjs/vecty/example/markdown  
* https://jsgo.io/gist.github.com/d6b70ceef39da20906ddf709d4a054c6  

The power of aggressive caching is apparent when loading pages which share common packages... The examples
in the [ebiten](https://github.com/hajimehoshi/ebiten) game library are a great demonstration of this:  

* https://jsgo.io/dave/ebiten/examples/2048
* https://jsgo.io/dave/ebiten/examples/airship
* https://jsgo.io/dave/ebiten/examples/alphablending
* https://jsgo.io/dave/ebiten/examples/audio
* https://jsgo.io/dave/ebiten/examples/infinitescroll
* https://jsgo.io/dave/ebiten/examples/rotate
* https://jsgo.io/dave/ebiten/examples/sprites

### Index

You can customize the HTML by adding a file named `index.jsgo.html` to your package. Use `{{ .Script }}`
as the script src. See [todomvc](https://github.com/dave/todomvc/blob/master/index.jsgo.html) for an example.

### Progress

If the page contains an element with `id="log"`, the bootstrap code will update it with loading progress. 
To disable this, simply remove the log element from the page.

### Limitations

If there's any non git repositories (e.g. hg, svn or bzr) in your dependency tree, it will fail. This 
is unlikely to change. Workaround: vendor the dependencies and it'll work fine.  
