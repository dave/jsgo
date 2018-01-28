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

Go to [jsgo.io/&lt;path&gt;?compile](https://jsgo.io/github.com/dave/jstest?compile), 
and click the `Compile now` button. 

After it's finished, the page will be available at [jsgo.io/&lt;path&gt;](https://jsgo.io/github.com/dave/jstest), 
or the compile page will give you the URL of the generated JS file.

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