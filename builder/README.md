# Builder

This is a fork of [guthub.com/gopherjs/gopherjs/build @ 178c176a91fe05e3e6c58fa5c989bad19e6cdcb3](https://github.com/gopherjs/gopherjs/tree/178c176a91fe05e3e6c58fa5c989bad19e6cdcb3/build).

Done:
296de816d4fe28b61803072c3f89360e9d1823ff build: Add "purego" build tag to default context.
8fc1f3cabe719f570ed9a95ca3574aca26408855 build: Exclude linux-specific crypto/rand tests.
f681903bc8cc21f8940c141d2beb8660511bed5a build: Make fewer copies of NewBuildContext.
c121b3d6abdf0d7534656ac098f9ae8c3c204abe build: Load js, nosync packages from gopherjspkg rather than GOPATH.
b90dbcb0f8851c1daa21542996b5f23d9779f720 build, compiler/typesutil: Don't use vendor directory to resolve js, nosync.
b24e3563d0efa33b32250411d469ba59fa5cf254 Restore support for testing js package.
