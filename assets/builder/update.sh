#!/bin/sh
set -e

tmp=$(mktemp -d "${TMPDIR:-/tmp}/gopherjs_playground.XXXXXXXXXX")

cleanup() {
    rm -rf "$tmp"
    exit
}

trap cleanup EXIT SIGHUP SIGINT SIGTERM

go install github.com/gopherjs/gopherjs/...

# The GOPATH workspace where the GopherJS project is.
gopherjsgopath=$(go list -f '{{.Root}}' github.com/gopherjs/gopherjs)

rm -rf assets/static/pkg/
rm -rf assets/static/pkg_min/

rm -rf assets/static/goroot/
rsync -av "$(go env GOROOT)/src" "assets/static/goroot" --exclude *_test.go

mkdir -p assets/static/goroot/src/github.com/gopherjs/gopherjs
rsync -av "$(go env GOPATH)/src/github.com/gopherjs/gopherjs/js" "assets/static/goroot/src/github.com/gopherjs/gopherjs" --exclude *_test.go
rsync -av "$(go env GOPATH)/src/github.com/gopherjs/gopherjs/nosync" "assets/static/goroot/src/github.com/gopherjs/gopherjs" --exclude *_test.go

# Use an empty GOPATH workspace with just gopherjs,
# so that all the standard library packages get written to GOROOT/pkg.
export GOPATH="$tmp/gopath"
mkdir -p "$GOPATH"/src/github.com/gopherjs/gopherjs
cp -a "$gopherjsgopath"/src/github.com/gopherjs/gopherjs/* "$GOPATH"/src/github.com/gopherjs/gopherjs

# minified / non-minified:
gopherjs install -m github.com/gopherjs/gopherjs/js github.com/gopherjs/gopherjs/nosync
gopherjs install github.com/gopherjs/gopherjs/js github.com/gopherjs/gopherjs/nosync
mkdir -p assets/static/pkg/github.com/gopherjs/gopherjs
mkdir -p assets/static/pkg_min/github.com/gopherjs/gopherjs

# minified:
cp "$GOPATH"/pkg/*_js_min/github.com/gopherjs/gopherjs/js.a assets/static/pkg_min/github.com/gopherjs/gopherjs/js.a
cp "$GOPATH"/pkg/*_js_min/github.com/gopherjs/gopherjs/nosync.a assets/static/pkg_min/github.com/gopherjs/gopherjs/nosync.a

# non-minified:
cp "$GOPATH"/pkg/*_js/github.com/gopherjs/gopherjs/js.a assets/static/pkg/github.com/gopherjs/gopherjs/js.a
cp "$GOPATH"/pkg/*_js/github.com/gopherjs/gopherjs/nosync.a assets/static/pkg/github.com/gopherjs/gopherjs/nosync.a

# Make a copy of GOROOT that is user-writeable,
# use it to build and copy out standard library packages.
cp -a "$(go env GOROOT)" "$tmp/goroot"
export GOROOT="$tmp/goroot"

# cd $(go env GOROOT)/src && go list ./...

set +e

# minified:
gopherjs install archive/tar
gopherjs install archive/zip
gopherjs install bufio
#gopherjs install builtin
gopherjs install bytes
gopherjs install cmd/addr2line
gopherjs install cmd/api
gopherjs install cmd/asm
gopherjs install cmd/asm/internal/arch
gopherjs install cmd/asm/internal/asm
gopherjs install cmd/asm/internal/flags
gopherjs install cmd/asm/internal/lex
gopherjs install cmd/cgo
gopherjs install cmd/compile
gopherjs install cmd/compile/internal/amd64
gopherjs install cmd/compile/internal/arm
gopherjs install cmd/compile/internal/arm64
gopherjs install cmd/compile/internal/gc
gopherjs install cmd/compile/internal/mips
gopherjs install cmd/compile/internal/mips64
gopherjs install cmd/compile/internal/ppc64
gopherjs install cmd/compile/internal/s390x
gopherjs install cmd/compile/internal/ssa
gopherjs install cmd/compile/internal/syntax
gopherjs install cmd/compile/internal/test
gopherjs install cmd/compile/internal/types
gopherjs install cmd/compile/internal/x86
gopherjs install cmd/cover
gopherjs install cmd/dist
gopherjs install cmd/doc
gopherjs install cmd/fix
gopherjs install cmd/go
gopherjs install cmd/go/internal/base
gopherjs install cmd/go/internal/bug
gopherjs install cmd/go/internal/buildid
gopherjs install cmd/go/internal/cfg
gopherjs install cmd/go/internal/clean
gopherjs install cmd/go/internal/cmdflag
gopherjs install cmd/go/internal/doc
gopherjs install cmd/go/internal/envcmd
gopherjs install cmd/go/internal/fix
gopherjs install cmd/go/internal/fmtcmd
gopherjs install cmd/go/internal/generate
gopherjs install cmd/go/internal/get
gopherjs install cmd/go/internal/help
gopherjs install cmd/go/internal/list
gopherjs install cmd/go/internal/load
gopherjs install cmd/go/internal/run
gopherjs install cmd/go/internal/str
gopherjs install cmd/go/internal/test
gopherjs install cmd/go/internal/tool
gopherjs install cmd/go/internal/version
gopherjs install cmd/go/internal/vet
gopherjs install cmd/go/internal/web
gopherjs install cmd/go/internal/work
gopherjs install cmd/gofmt
gopherjs install cmd/internal/bio
gopherjs install cmd/internal/browser
gopherjs install cmd/internal/dwarf
gopherjs install cmd/internal/gcprog
gopherjs install cmd/internal/goobj
gopherjs install cmd/internal/obj
gopherjs install cmd/internal/obj/arm
gopherjs install cmd/internal/obj/arm64
gopherjs install cmd/internal/obj/mips
gopherjs install cmd/internal/obj/ppc64
gopherjs install cmd/internal/obj/s390x
gopherjs install cmd/internal/obj/x86
gopherjs install cmd/internal/objabi
gopherjs install cmd/internal/objfile
gopherjs install cmd/internal/src
gopherjs install cmd/internal/sys
gopherjs install cmd/link
gopherjs install cmd/link/internal/amd64
gopherjs install cmd/link/internal/arm
gopherjs install cmd/link/internal/arm64
gopherjs install cmd/link/internal/ld
gopherjs install cmd/link/internal/mips
gopherjs install cmd/link/internal/mips64
gopherjs install cmd/link/internal/ppc64
gopherjs install cmd/link/internal/s390x
gopherjs install cmd/link/internal/x86
gopherjs install cmd/nm
gopherjs install cmd/objdump
gopherjs install cmd/pack
gopherjs install cmd/pprof
gopherjs install cmd/trace
gopherjs install cmd/vet
gopherjs install cmd/vet/internal/cfg
gopherjs install cmd/vet/internal/whitelist
gopherjs install compress/bzip2
gopherjs install compress/flate
gopherjs install compress/gzip
gopherjs install compress/lzw
gopherjs install compress/zlib
gopherjs install container/heap
gopherjs install container/list
gopherjs install container/ring
gopherjs install context
gopherjs install crypto
gopherjs install crypto/aes
gopherjs install crypto/cipher
gopherjs install crypto/des
gopherjs install crypto/dsa
gopherjs install crypto/ecdsa
gopherjs install crypto/elliptic
gopherjs install crypto/hmac
gopherjs install crypto/internal/cipherhw
gopherjs install crypto/md5
gopherjs install crypto/rand
gopherjs install crypto/rc4
gopherjs install crypto/rsa
gopherjs install crypto/sha1
gopherjs install crypto/sha256
gopherjs install crypto/sha512
gopherjs install crypto/subtle
gopherjs install crypto/tls
gopherjs install crypto/x509
gopherjs install crypto/x509/pkix
gopherjs install database/sql
gopherjs install database/sql/driver
gopherjs install debug/dwarf
gopherjs install debug/elf
gopherjs install debug/gosym
gopherjs install debug/macho
gopherjs install debug/pe
gopherjs install debug/plan9obj
gopherjs install encoding
gopherjs install encoding/ascii85
gopherjs install encoding/asn1
gopherjs install encoding/base32
gopherjs install encoding/base64
gopherjs install encoding/binary
gopherjs install encoding/csv
gopherjs install encoding/gob
gopherjs install encoding/hex
gopherjs install encoding/json
gopherjs install encoding/pem
gopherjs install encoding/xml
gopherjs install errors
gopherjs install expvar
gopherjs install flag
gopherjs install fmt
gopherjs install go/ast
gopherjs install go/build
gopherjs install go/constant
gopherjs install go/doc
gopherjs install go/format
gopherjs install go/importer
gopherjs install go/internal/gccgoimporter
gopherjs install go/internal/gcimporter
gopherjs install go/internal/srcimporter
gopherjs install go/parser
gopherjs install go/printer
gopherjs install go/scanner
gopherjs install go/token
gopherjs install go/types
gopherjs install hash
gopherjs install hash/adler32
gopherjs install hash/crc32
gopherjs install hash/crc64
gopherjs install hash/fnv
gopherjs install html
gopherjs install html/template
gopherjs install image
gopherjs install image/color
gopherjs install image/color/palette
gopherjs install image/draw
gopherjs install image/gif
gopherjs install image/internal/imageutil
gopherjs install image/jpeg
gopherjs install image/png
gopherjs install index/suffixarray
#gopherjs install internal/cpu
gopherjs install internal/nettrace
gopherjs install internal/poll
gopherjs install internal/race
gopherjs install internal/singleflight
gopherjs install internal/syscall/windows
gopherjs install internal/syscall/windows/registry
gopherjs install internal/syscall/windows/sysdll
gopherjs install internal/testenv
gopherjs install internal/trace
gopherjs install io
gopherjs install io/ioutil
gopherjs install log
gopherjs install log/syslog
gopherjs install math
gopherjs install math/big
gopherjs install math/bits
gopherjs install math/cmplx
gopherjs install math/rand
gopherjs install mime
gopherjs install mime/multipart
gopherjs install mime/quotedprintable
gopherjs install net
gopherjs install net/http
gopherjs install net/http/cgi
gopherjs install net/http/cookiejar
gopherjs install net/http/fcgi
gopherjs install net/http/httptest
gopherjs install net/http/httptrace
gopherjs install net/http/httputil
gopherjs install net/http/internal
#gopherjs install net/http/pprof
gopherjs install net/internal/socktest
gopherjs install net/mail
gopherjs install net/rpc
gopherjs install net/rpc/jsonrpc
gopherjs install net/smtp
gopherjs install net/textproto
gopherjs install net/url
gopherjs install os
gopherjs install os/exec
gopherjs install os/signal
gopherjs install os/user
gopherjs install path
gopherjs install path/filepath
#gopherjs install plugin
gopherjs install reflect
gopherjs install regexp
gopherjs install regexp/syntax
gopherjs install runtime
#gopherjs install runtime/cgo
gopherjs install runtime/debug
gopherjs install runtime/internal/atomic
gopherjs install runtime/internal/sys
gopherjs install runtime/pprof
gopherjs install runtime/pprof/internal/profile
gopherjs install runtime/race
gopherjs install runtime/trace
gopherjs install sort
gopherjs install strconv
gopherjs install strings
gopherjs install sync
gopherjs install sync/atomic
gopherjs install syscall
gopherjs install testing
gopherjs install testing/internal/testdeps
gopherjs install testing/iotest
gopherjs install testing/quick
gopherjs install text/scanner
gopherjs install text/tabwriter
gopherjs install text/template
gopherjs install text/template/parse
gopherjs install time
gopherjs install unicode
gopherjs install unicode/utf16
gopherjs install unicode/utf8
gopherjs install unsafe

gopherjs install -m archive/tar
gopherjs install -m archive/zip
gopherjs install -m bufio
#gopherjs install -m builtin
gopherjs install -m bytes
gopherjs install -m cmd/addr2line
gopherjs install -m cmd/api
gopherjs install -m cmd/asm
gopherjs install -m cmd/asm/internal/arch
gopherjs install -m cmd/asm/internal/asm
gopherjs install -m cmd/asm/internal/flags
gopherjs install -m cmd/asm/internal/lex
gopherjs install -m cmd/cgo
gopherjs install -m cmd/compile
gopherjs install -m cmd/compile/internal/amd64
gopherjs install -m cmd/compile/internal/arm
gopherjs install -m cmd/compile/internal/arm64
gopherjs install -m cmd/compile/internal/gc
gopherjs install -m cmd/compile/internal/mips
gopherjs install -m cmd/compile/internal/mips64
gopherjs install -m cmd/compile/internal/ppc64
gopherjs install -m cmd/compile/internal/s390x
gopherjs install -m cmd/compile/internal/ssa
gopherjs install -m cmd/compile/internal/syntax
gopherjs install -m cmd/compile/internal/test
gopherjs install -m cmd/compile/internal/types
gopherjs install -m cmd/compile/internal/x86
gopherjs install -m cmd/cover
gopherjs install -m cmd/dist
gopherjs install -m cmd/doc
gopherjs install -m cmd/fix
gopherjs install -m cmd/go
gopherjs install -m cmd/go/internal/base
gopherjs install -m cmd/go/internal/bug
gopherjs install -m cmd/go/internal/buildid
gopherjs install -m cmd/go/internal/cfg
gopherjs install -m cmd/go/internal/clean
gopherjs install -m cmd/go/internal/cmdflag
gopherjs install -m cmd/go/internal/doc
gopherjs install -m cmd/go/internal/envcmd
gopherjs install -m cmd/go/internal/fix
gopherjs install -m cmd/go/internal/fmtcmd
gopherjs install -m cmd/go/internal/generate
gopherjs install -m cmd/go/internal/get
gopherjs install -m cmd/go/internal/help
gopherjs install -m cmd/go/internal/list
gopherjs install -m cmd/go/internal/load
gopherjs install -m cmd/go/internal/run
gopherjs install -m cmd/go/internal/str
gopherjs install -m cmd/go/internal/test
gopherjs install -m cmd/go/internal/tool
gopherjs install -m cmd/go/internal/version
gopherjs install -m cmd/go/internal/vet
gopherjs install -m cmd/go/internal/web
gopherjs install -m cmd/go/internal/work
gopherjs install -m cmd/gofmt
gopherjs install -m cmd/internal/bio
gopherjs install -m cmd/internal/browser
gopherjs install -m cmd/internal/dwarf
gopherjs install -m cmd/internal/gcprog
gopherjs install -m cmd/internal/goobj
gopherjs install -m cmd/internal/obj
gopherjs install -m cmd/internal/obj/arm
gopherjs install -m cmd/internal/obj/arm64
gopherjs install -m cmd/internal/obj/mips
gopherjs install -m cmd/internal/obj/ppc64
gopherjs install -m cmd/internal/obj/s390x
gopherjs install -m cmd/internal/obj/x86
gopherjs install -m cmd/internal/objabi
gopherjs install -m cmd/internal/objfile
gopherjs install -m cmd/internal/src
gopherjs install -m cmd/internal/sys
gopherjs install -m cmd/link
gopherjs install -m cmd/link/internal/amd64
gopherjs install -m cmd/link/internal/arm
gopherjs install -m cmd/link/internal/arm64
gopherjs install -m cmd/link/internal/ld
gopherjs install -m cmd/link/internal/mips
gopherjs install -m cmd/link/internal/mips64
gopherjs install -m cmd/link/internal/ppc64
gopherjs install -m cmd/link/internal/s390x
gopherjs install -m cmd/link/internal/x86
gopherjs install -m cmd/nm
gopherjs install -m cmd/objdump
gopherjs install -m cmd/pack
gopherjs install -m cmd/pprof
gopherjs install -m cmd/trace
gopherjs install -m cmd/vet
gopherjs install -m cmd/vet/internal/cfg
gopherjs install -m cmd/vet/internal/whitelist
gopherjs install -m compress/bzip2
gopherjs install -m compress/flate
gopherjs install -m compress/gzip
gopherjs install -m compress/lzw
gopherjs install -m compress/zlib
gopherjs install -m container/heap
gopherjs install -m container/list
gopherjs install -m container/ring
gopherjs install -m context
gopherjs install -m crypto
gopherjs install -m crypto/aes
gopherjs install -m crypto/cipher
gopherjs install -m crypto/des
gopherjs install -m crypto/dsa
gopherjs install -m crypto/ecdsa
gopherjs install -m crypto/elliptic
gopherjs install -m crypto/hmac
gopherjs install -m crypto/internal/cipherhw
gopherjs install -m crypto/md5
gopherjs install -m crypto/rand
gopherjs install -m crypto/rc4
gopherjs install -m crypto/rsa
gopherjs install -m crypto/sha1
gopherjs install -m crypto/sha256
gopherjs install -m crypto/sha512
gopherjs install -m crypto/subtle
gopherjs install -m crypto/tls
gopherjs install -m crypto/x509
gopherjs install -m crypto/x509/pkix
gopherjs install -m database/sql
gopherjs install -m database/sql/driver
gopherjs install -m debug/dwarf
gopherjs install -m debug/elf
gopherjs install -m debug/gosym
gopherjs install -m debug/macho
gopherjs install -m debug/pe
gopherjs install -m debug/plan9obj
gopherjs install -m encoding
gopherjs install -m encoding/ascii85
gopherjs install -m encoding/asn1
gopherjs install -m encoding/base32
gopherjs install -m encoding/base64
gopherjs install -m encoding/binary
gopherjs install -m encoding/csv
gopherjs install -m encoding/gob
gopherjs install -m encoding/hex
gopherjs install -m encoding/json
gopherjs install -m encoding/pem
gopherjs install -m encoding/xml
gopherjs install -m errors
gopherjs install -m expvar
gopherjs install -m flag
gopherjs install -m fmt
gopherjs install -m go/ast
gopherjs install -m go/build
gopherjs install -m go/constant
gopherjs install -m go/doc
gopherjs install -m go/format
gopherjs install -m go/importer
gopherjs install -m go/internal/gccgoimporter
gopherjs install -m go/internal/gcimporter
gopherjs install -m go/internal/srcimporter
gopherjs install -m go/parser
gopherjs install -m go/printer
gopherjs install -m go/scanner
gopherjs install -m go/token
gopherjs install -m go/types
gopherjs install -m hash
gopherjs install -m hash/adler32
gopherjs install -m hash/crc32
gopherjs install -m hash/crc64
gopherjs install -m hash/fnv
gopherjs install -m html
gopherjs install -m html/template
gopherjs install -m image
gopherjs install -m image/color
gopherjs install -m image/color/palette
gopherjs install -m image/draw
gopherjs install -m image/gif
gopherjs install -m image/internal/imageutil
gopherjs install -m image/jpeg
gopherjs install -m image/png
gopherjs install -m index/suffixarray
#gopherjs install -m internal/cpu
gopherjs install -m internal/nettrace
gopherjs install -m internal/poll
gopherjs install -m internal/race
gopherjs install -m internal/singleflight
gopherjs install -m internal/syscall/windows
gopherjs install -m internal/syscall/windows/registry
gopherjs install -m internal/syscall/windows/sysdll
gopherjs install -m internal/testenv
gopherjs install -m internal/trace
gopherjs install -m io
gopherjs install -m io/ioutil
gopherjs install -m log
gopherjs install -m log/syslog
gopherjs install -m math
gopherjs install -m math/big
gopherjs install -m math/bits
gopherjs install -m math/cmplx
gopherjs install -m math/rand
gopherjs install -m mime
gopherjs install -m mime/multipart
gopherjs install -m mime/quotedprintable
gopherjs install -m net
gopherjs install -m net/http
gopherjs install -m net/http/cgi
gopherjs install -m net/http/cookiejar
gopherjs install -m net/http/fcgi
gopherjs install -m net/http/httptest
gopherjs install -m net/http/httptrace
gopherjs install -m net/http/httputil
gopherjs install -m net/http/internal
#gopherjs install -m net/http/pprof
gopherjs install -m net/internal/socktest
gopherjs install -m net/mail
gopherjs install -m net/rpc
gopherjs install -m net/rpc/jsonrpc
gopherjs install -m net/smtp
gopherjs install -m net/textproto
gopherjs install -m net/url
gopherjs install -m os
gopherjs install -m os/exec
gopherjs install -m os/signal
gopherjs install -m os/user
gopherjs install -m path
gopherjs install -m path/filepath
#gopherjs install -m plugin
gopherjs install -m reflect
gopherjs install -m regexp
gopherjs install -m regexp/syntax
gopherjs install -m runtime
#gopherjs install -m runtime/cgo
gopherjs install -m runtime/debug
gopherjs install -m runtime/internal/atomic
gopherjs install -m runtime/internal/sys
gopherjs install -m runtime/pprof
gopherjs install -m runtime/pprof/internal/profile
gopherjs install -m runtime/race
gopherjs install -m runtime/trace
gopherjs install -m sort
gopherjs install -m strconv
gopherjs install -m strings
gopherjs install -m sync
gopherjs install -m sync/atomic
gopherjs install -m syscall
gopherjs install -m testing
gopherjs install -m testing/internal/testdeps
gopherjs install -m testing/iotest
gopherjs install -m testing/quick
gopherjs install -m text/scanner
gopherjs install -m text/tabwriter
gopherjs install -m text/template
gopherjs install -m text/template/parse
gopherjs install -m time
gopherjs install -m unicode
gopherjs install -m unicode/utf16
gopherjs install -m unicode/utf8
gopherjs install -m unsafe

# minified:
cp -a "$GOROOT"/pkg/*_js_min/* assets/static/pkg_min/
cp -a "$GOROOT"/pkg/*_amd64_js_min/* assets/static/pkg_min/

# non-minified:
cp -a "$GOROOT"/pkg/*_js/* assets/static/pkg/
cp -a "$GOROOT"/pkg/*_amd64_js/* assets/static/pkg/
