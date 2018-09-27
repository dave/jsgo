<a href="https://patreon.com/davebrophy" title="Help with my hosting bills using Patreon"><img src="https://img.shields.io/badge/patreon-donate-yellow.svg" style="max-width:100%;"></a>

# jsgo.io

[GopherJS](https://github.com/gopherjs/gopherjs) is an amazing tool, but I've always been frustrated 
by the size of the output. All the packages in the dependency tree (including the standard library) 
are compiled to a single JS file. This can cause the resultant file to be several megabytes. 

I've always thought a better solution would be to split the JS up by package and store it in a centralized 
CDN. This architecture would then allow aggressive caching: If you import `fmt`, it'll be delivered as 
a separate file `fmt.js`, and there's a good chance some of your visitors will already have it in their 
browser cache. Additionally, incremental updates to your app will only change the package you're updating, 
so your visitors won't have to download the entire dependency tree again.

`jsgo.io` makes this simple. 

<img width="75%" src="https://user-images.githubusercontent.com/925351/36342450-1595a85a-13ff-11e8-9ebe-7019c3f4d1af.png">

### Features

* Compiles Go to JS using [GopherJS](https://github.com/gopherjs/gopherjs).  
* Splits the JS up by Go package.  
* Stores the JS in a CDN for you (GCP / Cloudflare).  
* Aggressively caches the JS.  
* Creates a page on `jsgo.io` that runs the JS.  
* Creates a single `loader JS` file you can use on your site.

### How it works

Visit `https://compile.jsgo.io/<path>` to compile or re-compile your package. Here's a [very simple 
hello world](https://compile.jsgo.io/github.com/dave/jstest). Just click `Compile`.

After it's finished, you'll be shown a link to a [page that runs the code](https://jsgo.io/dave/jstest) 
on `jsgo.io`. The compile page will also give you a link to a single JS file on `pkg.jsgo.io` - this 
is the `loader JS` for your package. Add this in a `<script>` tag on your site and it will download 
all the dependencies and execute your package.

URLs on `jsgo.io` that start `github.com` may be abbreviated: `github.com/foo/bar` will be available 
at `jsgo.io/foo/bar` and also `jsgo.io/github.com/foo/bar`. Package URLs on `pkg.jsgo.io` always use 
the full path.  

### Production ready?

The package CDN (everything on `pkg.jsgo.io`) should be considered relatively production ready - it's 
just static JS files in a Google Storage bucket behind a Cloudflare CDN so there's very little that can 
go wrong. Additionally, the URL of each file contains a hash of it's contents, ensuring immutability.

The index pages (everything on `jsgo.io`) should only be used for testing and toy projects. Remember 
you're sharing a domain with everyone else, so the browser environment (cookies, local storage etc.) 
should be used with caution! For anything important, create your own index page on your site and add 
the `loader JS` (on `pkg.jsgo.io`) to a `<script>` tag.  

Ths compile server (everything on `compile.jsgo.io`) should be considered in beta... Please [add an issue](https://github.com/dave/jsgo/issues) 
if it's having trouble compiling your project. 

### Demos

* https://jsgo.io/hajimehoshi/go-inovation
* https://jsgo.io/hajimehoshi/ebiten/examples/2048
* https://jsgo.io/shurcooL/tictactoe/cmd/tictactoe
* https://jsgo.io/dave/todomvc
* https://jsgo.io/gopherjs/vecty/example/markdown
* https://jsgo.io/dave/html2vecty
* https://jsgo.io/dave/zip
* https://jsgo.io/dave/img

The power of aggressive caching is apparent when loading pages which share common packages... The examples
in the [ebiten](https://github.com/hajimehoshi/ebiten) game library are a great demonstration of this:  

* https://jsgo.io/hajimehoshi/ebiten/examples/blocks
* https://jsgo.io/hajimehoshi/ebiten/examples/airship
* https://jsgo.io/hajimehoshi/ebiten/examples/alphablending
* https://jsgo.io/hajimehoshi/ebiten/examples/audio
* https://jsgo.io/hajimehoshi/ebiten/examples/infinitescroll
* https://jsgo.io/hajimehoshi/ebiten/examples/rotate
* https://jsgo.io/hajimehoshi/ebiten/examples/sprites

### Index

You can customize the HTML delivered by the `jsgo.io` page by adding a file named `index.jsgo.html` to 
your package. Use `{{ .Script }}` as the script src. See [todomvc](https://github.com/dave/todomvc/blob/master/index.jsgo.html) 
for an example.

### Progress

If a function `window.jsgoProgress` exists, it will be called repeatedly as packages load. Two parameters 
are supplied: `count` (the number of packages loaded so far) and `total` (the total number of packages).

The default index page on `jsgo.io` is to display a simple `count / total` message in a span. However, 
by supplying a custom `index.jsgo.html`, more complex effects may be created - see the [html2vecty 
example](https://jsgo.io/dave/html2vecty) for a [bootstrap progress bar](https://github.com/dave/html2vecty/blob/master/index.jsgo.html).

### Limitations

If there's any non git repositories (e.g. hg, svn or bzr) in your dependency tree, it will fail. This 
is unlikely to change. Workaround: vendor the dependencies and it'll work fine.  

### How to contact me

If you'd like to chat more about the project, feel free to [add an issue](https://github.com/dave/jsgo/issues), 
mention [@dave](https://github.com/dave/) in your PR, email me or post in the #gopherjs channel of the 
Gophers Slack. I'm happy to help!

### Run locally?

If you'd like to run jsgo locally, take a look at [these instructions](LOCAL.md).