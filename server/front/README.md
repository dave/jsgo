
# GCS bucket layout

### Package JS
`/<package-path>/package.<hash>.js`
Package js, hash is the sha1 of the js file.

### Main JS (only for main packages)
`/<package-path>/main.<hash>.js`
Initializer js which downloads all dependencies and runs main(). Hash is the git id of this repo at 
the time it was compiled

### Standard library
`/<package-path>/package.<hash>.js`
Stdlib will be refreshed each time Go / GopherJS is released. Compiled packages will continue to 
use the old versions until they are recompiled.   

# External URLs

### Main url
`jsgo.io/<package-path>`
1) Looks up <package-path> in database, for git id of last compile.
2a) If none found, redirect to the interactive compile page.
2b) If found, shows a page with <script src="<cdn-host>/<package-path>/main.<hash>.js">

### JS url
`jsgo.io/<package-path>.js`
1) Looks up <package-path> in database, for git id of last compile.
2a) If none found, return 404.
2b) If found, redirect to <cdn-host>/<package-path>/main.<hash>.js

### Compile page
`jsgo.io/<package-path>?compile`
Shows interactive compile page. On compile:
1) Gets all dependencies (git only)
2) Compiles all code to JS
3) Uploads package JS to GCS bucket for all dependencies
4) Uploads main JS to GCS bucket
5) Updates database with new package -> git hash