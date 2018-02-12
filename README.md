# jsgo.io

jsgo.io is a GopherJS serving framework and CDN. 

Please don't use this in production just yet!

### Features

* Compiles your Go to Javascript.  
* Splits the javascript up by Go package.  
* Stores the javascript in a CDN (cloudflare) for you.  
* Aggressively caches the javascript.  
* Serves a page that will run your javascript.  
* Alternatively you can drop a single JS file in your page.   

### How it works

Run your code in a blank page: `https://jsgo.io/<path>`  
Recompile your code: `https://compile.jsgo.io/<path>`  

Try it with a very simple hello world: Open [compile.jsgo.io/github.com/dave/jstest](https://compile.jsgo.io/github.com/dave/jstest)
and click `Compile`. 

After it's finished, you'll be shown a [link to the page](https://jsgo.io/github.com/dave/jstest). If 
you look at the `Network` panel in your browser inspector as the page loads, you'll see the packages 
downloading separately.

### Demos

* https://jsgo.io/github.com/dave/todomvc
* https://jsgo.io/github.com/gopherjs/vecty/example/markdown  
* https://jsgo.io/gist.github.com/d6b70ceef39da20906ddf709d4a054c6  

The power of aggressive caching is apparent when loading pages which share common packages... The [ebiten](https://github.com/hajimehoshi/ebiten)
game library examples are a great demonstration of this:  

* https://jsgo.io/github.com/dave/ebiten/examples/2048
* https://jsgo.io/github.com/dave/ebiten/examples/airship
* https://jsgo.io/github.com/dave/ebiten/examples/alphablending
* https://jsgo.io/github.com/dave/ebiten/examples/audio
* https://jsgo.io/github.com/dave/ebiten/examples/infinitescroll
* https://jsgo.io/github.com/dave/ebiten/examples/rotate
* https://jsgo.io/github.com/dave/ebiten/examples/sprites

### Index

You can customize the HTML by adding a file named `index.jsgo.html` to your package. Use `{{ .Script }}`
as the script src. See [todomvc](https://github.com/dave/todomvc/blob/master/index.jsgo.html) for an example.

### Progress

If the page contains an element with `id="log"`, the bootstrap code will update it with a loading progress. 
To disable this, simply remove the log element from the page.

### Limitations

If there's any non-git repositories in your dependency tree, it will fail. This is unlikely to change. 
Workaround: vendor the dependencies and it'll work fine.  
