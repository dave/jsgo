# jsgo

jsgo is a GopherJS serving framework and CDN.

Please don't use this in production just yet!

### Features

* Compiles your Go to Javascript.  
* Splits the javascript up by Go package.  
* Stores the javascript in a CDN (cloudflare) for you.  
* Aggressively caches the javascript.  
* Serves a page that will run your javascript.  
* Alternatively you can drop a single JS file in your page.   

### How it works

Start the JS on a blank page: https://jsgo.io/GOPHERJS_PACKAGE_PATH  
Recompile: https://compile.jsgo.io/GOPHERJS_PACKAGE_PATH  

Try it with [github.com/dave/jstest](https://github.com/dave/jstest/blob/master/main.go) - a very simple 
hello world: Open [compile.jsgo.io/github.com/dave/jstest](https://compile.jsgo.io/github.com/dave/jstest), 
and click the `Compile now` button. 

After it's finished, the URL of the compiled JS will be shown, or [jsgo.io/github.com/dave/jstest](https://jsgo.io/github.com/dave/jstest) 
loads the JS in an empty page.

It also works with Github gists: https://jsgo.io/gist.github.com/dave/d6b70ceef39da20906ddf709d4a054c6

### Demos

https://jsgo.io/github.com/gopherjs/vecty/example/markdown
https://jsgo.io/github.com/gopherjs/vecty/example/todomvc

### Limitations

1) The pages served on jsgo.io don't have any css or html - that's why the vecty demos above don't 
look quite right. Workaround: drop the JS file in your own page.  

2) If there's any non-git repositories in your dependency tree, it will fail. This is unlikely to 
change. Workaround: vendor the dependencies and it'll work fine.  

3) The compile page should update the progress as it's happening, but App Engine doesn't support 
unbuffered HTTP responses. This may be fixed soon.  