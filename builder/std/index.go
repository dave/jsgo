package std

import builder "github.com/dave/jsgo/builder"

var Index = map[string]builder.PackageHash{
	"archive/tar": {
		HashMax: "2f363c52d17420f1fc1ead597d66569e2667990d",
		HashMin: "5043861eddbc06dc9e8d9027fab8cf500ae0a069",
	},
	"archive/zip": {
		HashMax: "2aec1182f534196a78841c370db046a5a4e3ab9e",
		HashMin: "168639cde1132e23aeaeeb350aaea8ff7256f2a9",
	},
	"bufio": {
		HashMax: "8382b0e66d2ebaaf03a8d64a4d4dfb94c5fa416c",
		HashMin: "af7b6f34095585c748cb0ef2c48d959da019bdef",
	},
	"bytes": {
		HashMax: "8a33b0fdde15abec61da2175f2c331195fd94c75",
		HashMin: "59417a7e20d727ae1aece5e8e308ed2bfbdb2a00",
	},
	"cmd/addr2line": {
		HashMax: "d45e2a2cd499750b8a573471daa2d099cbfbe40e",
		HashMin: "9de495a9541e7a61cc83d2389205acbdb4b66b13",
	},
	"cmd/api": {
		HashMax: "b3083e41bdbbce9ec27805f20c665438473506b4",
		HashMin: "25793ac6b03d2d8f5c282290b306c0d79c9de775",
	},
	"cmd/asm": {
		HashMax: "21f65d07836c9663bbe69543c7da2deb6afc8e67",
		HashMin: "0cc68744f468cedb910f68f738d946bd2849f96a",
	},
	"cmd/asm/internal/arch": {
		HashMax: "e1fef02f3dca3b29f05c2a0311a80b4b50dfc0a5",
		HashMin: "c71843f4fc8e3c1bc796822b4a8318cab189174f",
	},
	"cmd/asm/internal/asm": {
		HashMax: "1a1be05c3641438c40e8d52f981078f9e0fd7543",
		HashMin: "e502714a432c63eb5e7b01d7eb62b58d90e9d771",
	},
	"cmd/asm/internal/flags": {
		HashMax: "ae561941c9c5cccfed5086c8268cc6c456b54c39",
		HashMin: "bdac94299eabd511ff2c310371d8a60a5af77427",
	},
	"cmd/asm/internal/lex": {
		HashMax: "0b88efa1d27772df3ff28ca1b20031880ee9ec47",
		HashMin: "7cd5bc0bd33c0ea01681f1ce57167fb3c731e888",
	},
	"cmd/buildid": {
		HashMax: "4c822910caf90e44bb5bb4828268e1a0e65cb597",
		HashMin: "aa9521cbcf38427da8e1842e9556e7fb04ae2a8d",
	},
	"cmd/cgo": {
		HashMax: "b326b227c2c23538f7869f73ffb06344f95d8bd3",
		HashMin: "be8f311178f8c1922643b11b657164e39f6ea0b8",
	},
	"cmd/compile": {
		HashMax: "5d8558b4d66c39213d1edfb5a64871c8063674e8",
		HashMin: "735a0cb6db163779a5997e7efec2803fee502f43",
	},
	"cmd/compile/internal/amd64": {
		HashMax: "b3351c08b7ca83fa820d23fc893b5885ad2123c0",
		HashMin: "5dd9b0a78a56a448ffe55a2144aee7b64aeb89be",
	},
	"cmd/compile/internal/arm": {
		HashMax: "fde17291003ef2c51e4bb3a485ab05466162f310",
		HashMin: "442183ef0278c66177e4feccab899966416f36ce",
	},
	"cmd/compile/internal/arm64": {
		HashMax: "7f1e44a86ebef636fe6fe636ee769e233b7f5692",
		HashMin: "2ab5887ede37f1a4c53f3b214253b51a7be26ceb",
	},
	"cmd/compile/internal/gc": {
		HashMax: "a0e450e55ff484708189708b4801950ffa3f9415",
		HashMin: "4630353f900f5c6c4a2506e934eea0e46b2869e9",
	},
	"cmd/compile/internal/mips": {
		HashMax: "b25b79dba18f052ef5627bf7bf25ab394fc90162",
		HashMin: "de816777540be15b78c84115d0361a8c11f8fc93",
	},
	"cmd/compile/internal/mips64": {
		HashMax: "0c61f45a64c26c28f0a6000de6f638732096481a",
		HashMin: "d2f680c80c034c89316ef92a6bc98f912611d641",
	},
	"cmd/compile/internal/ppc64": {
		HashMax: "418d2be4c42f123299922d3802e7193688a154bd",
		HashMin: "d75e19047e807ebd55562f20cb78870ae9fe3c7b",
	},
	"cmd/compile/internal/s390x": {
		HashMax: "d40384853e0c6fea00542914d19f7b3c4d8ebb26",
		HashMin: "b0e326346c75891eb26ab4ef9665853954a1f11f",
	},
	"cmd/compile/internal/ssa": {
		HashMax: "6a6278b23ba6b243f5bb402788bc6a5033275662",
		HashMin: "8b9407c8b6911d5bf2cbb00cf078377f02b7f771",
	},
	"cmd/compile/internal/syntax": {
		HashMax: "b3924aeb13a77c1e5c695f7d89bdb6fd368c3f74",
		HashMin: "287587607d2746b8ac0913f413f1d5b73cba8c50",
	},
	"cmd/compile/internal/test": {
		HashMax: "be81a3174788f01ab6bc4372c35f36bbb4b59656",
		HashMin: "0c60886ab4c2be1920d2d6d533d6c8a3f53bebb0",
	},
	"cmd/compile/internal/types": {
		HashMax: "11e1d8634e94c5131e89e5ae0fbb691db9da4cc9",
		HashMin: "590a22f13c80143a5317b17acfa22f99c67b4e29",
	},
	"cmd/compile/internal/x86": {
		HashMax: "45847d43f4140739bf29247918809201891bc660",
		HashMin: "2d72c381a41dd00fa39875a6430c0ba345504c5c",
	},
	"cmd/cover": {
		HashMax: "a1352b52fd2710cba931144d15499ea13df88ae2",
		HashMin: "fb2d027ffea396765f5e5aade3ef9f4549a7c257",
	},
	"cmd/dist": {
		HashMax: "7470e54ab63be3599684b1278c68e0b635bc1259",
		HashMin: "d1ba849064e224fc82c312205c609cd6034f225b",
	},
	"cmd/doc": {
		HashMax: "94ba15caf58a12f027da1b97e7b897c52a9a6e21",
		HashMin: "5137f1f7c4336976293d6411317012c16ddce63e",
	},
	"cmd/fix": {
		HashMax: "9ede20543289beb5da3cdd4714f4d94a39fee6dd",
		HashMin: "f2462979fb929f3c31033f53d6716c863b260652",
	},
	"cmd/go": {
		HashMax: "f5cbf99eaed2e8d5d423a48c56e817b1e55adba0",
		HashMin: "52afd425a0ab488142638dea5a61b78e3496b1cb",
	},
	"cmd/go/internal/base": {
		HashMax: "887e4b29c1dad469b85ca81467890ee9ee5db2af",
		HashMin: "9ccfeae12a3f2bef55ef0d186d865d7886144874",
	},
	"cmd/go/internal/bug": {
		HashMax: "d9bfe467ca24cc93c52cccf962b6bea899b60e97",
		HashMin: "8a304f4771a5e531c40e9af60a86205ce91f490f",
	},
	"cmd/go/internal/cache": {
		HashMax: "0e2c0555541a0516e2ffe084ae04f760888e5d54",
		HashMin: "596673bdc9d7f8d3df30c5b868cc73c66da33f4d",
	},
	"cmd/go/internal/cfg": {
		HashMax: "009e8415278850e313c7a43e577264a1f91fef16",
		HashMin: "2e093543aa28f9b90c59203b3feadeeea1b6e301",
	},
	"cmd/go/internal/clean": {
		HashMax: "421eb3e9998dc8cc7a940d8ed05be4ece0bb998a",
		HashMin: "b1c8c9e69cf2aa89742afb937f7c77ea82d8e091",
	},
	"cmd/go/internal/cmdflag": {
		HashMax: "6064ca3e696d8cf9a31e7191b2bc0a6fec4530e6",
		HashMin: "c15db4e7926df62554754595c9e1cc0eb149766f",
	},
	"cmd/go/internal/doc": {
		HashMax: "24656d70989a85d6dc1cf41fe6f6acda2c33ec42",
		HashMin: "d21252d7c7ebe6692fb330152df1e6aa26da3f15",
	},
	"cmd/go/internal/envcmd": {
		HashMax: "916650ab5202ed1bd3c2dac3d7924bff36908fbf",
		HashMin: "3ed3b1781581ba278dbf5532bdacd40cc4eed231",
	},
	"cmd/go/internal/fix": {
		HashMax: "4760700ecea05e8a2dc02dcc6d900ba0111d80f5",
		HashMin: "37278fc96ddb35a39b137d93719f660ee87a710a",
	},
	"cmd/go/internal/fmtcmd": {
		HashMax: "74ccf598a1b09a253515b1ed69e46f81b8e7d613",
		HashMin: "744133bfc5f4283f4a0ac0a6f3139e7b70ccd1ce",
	},
	"cmd/go/internal/generate": {
		HashMax: "2399a5a38d9ece8a2c4ad8f07d84b515c1d4e387",
		HashMin: "ba146dc1d97591143206cd55cf38a6528cdfd6dc",
	},
	"cmd/go/internal/get": {
		HashMax: "5de2706ee7b05ecdfe60946512d2068ac9cb2a37",
		HashMin: "b6278db118cbe1ac5f51911abff9861e542b5893",
	},
	"cmd/go/internal/help": {
		HashMax: "c21e25ec55d27ef4a893fde739eb644a6f096313",
		HashMin: "0e163d72cd11a3ba298266d92eb1549d20057eaa",
	},
	"cmd/go/internal/list": {
		HashMax: "217c2badf30f41bdf359f9cdc1ab37559bc482d8",
		HashMin: "25a8693f8c37ad56db8713a79d7f7b499b7acd44",
	},
	"cmd/go/internal/load": {
		HashMax: "fbc1374d2538d0f8fd0d39678f2567743cfabe84",
		HashMin: "b31b85cdc0d65f835b6390acfc60f235de22ee06",
	},
	"cmd/go/internal/run": {
		HashMax: "b2b136160fb7721a92857f8c54592cb461baaedb",
		HashMin: "0dee0bc226485e80e42209610b7bc578ceba645b",
	},
	"cmd/go/internal/str": {
		HashMax: "d6563e5f0616c67069a015c322a87c83fdf2e8c5",
		HashMin: "1d0183722838921b7db947dad04a3d24d8ebb481",
	},
	"cmd/go/internal/test": {
		HashMax: "542741b78ad46a4ffe6db8d4f530735845799c74",
		HashMin: "24d4951c44b8eecc034c606dc5fd9981b6d5cbae",
	},
	"cmd/go/internal/tool": {
		HashMax: "9ed4ba61e5ecac337feddc59500d328205538559",
		HashMin: "7ab0180ee4ac1b45c5912c6d76d8c4c0adccbbb4",
	},
	"cmd/go/internal/version": {
		HashMax: "4b9c0cde97e0923f3c36f2b5067b40adfef03bdd",
		HashMin: "d2c8862c5a92895524b86724ab7e9cfe07bd1f88",
	},
	"cmd/go/internal/vet": {
		HashMax: "4c3b342376baee229cee1163c3f9f7b8e8468224",
		HashMin: "2e25a416b14aa78c5a79d64e75a28f9d65f7849b",
	},
	"cmd/go/internal/web": {
		HashMax: "7083efa56a98d07792bcba65b56af86e9d05611e",
		HashMin: "38db1639981e5e8272c4b2c28516b8a71ed09f04",
	},
	"cmd/go/internal/work": {
		HashMax: "7162ee5be6da43cf818c9fbec3ae1bcbd9abc18f",
		HashMin: "f4fd30fb00ac5b861a7407031e171ee504ec8be9",
	},
	"cmd/gofmt": {
		HashMax: "a3d5bec23c0bb2b2c5d1e36c3afaa93d1025ea9f",
		HashMin: "fae412b6a06ae80e38e35c1d8c4b839a89afb16e",
	},
	"cmd/internal/bio": {
		HashMax: "8034a758f1b0a7d831875ee00030174cc032d053",
		HashMin: "7f478fe091cf5219e29fe73b88dad99b6d884c17",
	},
	"cmd/internal/browser": {
		HashMax: "8f402df0c84ad41a7890f4c228fe349e5350d08b",
		HashMin: "b7e5f4e4156c400a9141ae3df5d5aaac093a503d",
	},
	"cmd/internal/buildid": {
		HashMax: "c9cdd6ea0707d4ca178cc98b4b4204c46e8b0834",
		HashMin: "2c729e0a3ee35b1c04e6e2fa398b06bf39925409",
	},
	"cmd/internal/dwarf": {
		HashMax: "802d790adc4d5a1faa697230bf4d3517c0696cec",
		HashMin: "5ba80da5dffc1891bc93417902fb14dc1b85aa77",
	},
	"cmd/internal/edit": {
		HashMax: "b884ff2d934f8f68b7d5b99148d4fc8559e3c430",
		HashMin: "87dd32dc1afb20cb23c7871eb93ca171aae37397",
	},
	"cmd/internal/gcprog": {
		HashMax: "debfa7b4217a9d59964c1fee4609a5e1fd1d38d5",
		HashMin: "b8a3fa62d1fe42e11ad388c1dd31ecdfd62e84f6",
	},
	"cmd/internal/goobj": {
		HashMax: "efa8d209b5fe73e40604739ae1b80a30399d3885",
		HashMin: "9a3c039dfbc77977edb0149c4432b089a75b5dc2",
	},
	"cmd/internal/obj": {
		HashMax: "2aa785d022d49f85d9e3e4680d00ead0db7af04c",
		HashMin: "eb6c6caeb1363a50940c02b80551c9837a3d0964",
	},
	"cmd/internal/obj/arm": {
		HashMax: "3719f346aae8fb2c19982fa4e88aaadaf8dd17b1",
		HashMin: "c1b50b89ba2b24d485957d1da91017fed0d1a42f",
	},
	"cmd/internal/obj/arm64": {
		HashMax: "4cb4a5cee06e09c3055dd0d004c7d5ec99802281",
		HashMin: "9fd4b5b75ddf3cbb7277aa69da365e8274e64bf0",
	},
	"cmd/internal/obj/mips": {
		HashMax: "6456bfb86459e1c4f9ee315a955da3062754d879",
		HashMin: "41f185bc529e12f957a08edb48057789d92e2a4f",
	},
	"cmd/internal/obj/ppc64": {
		HashMax: "2b43a3927637ec99bd5a5138b551ee227d5c6d9b",
		HashMin: "70a33d6000d483c201ea009c7d9efeda45a33605",
	},
	"cmd/internal/obj/s390x": {
		HashMax: "42d0dd2adcb1e306edfa91c0b0071596db109e05",
		HashMin: "4b336e581b52f041585a847494b753ff7678f71e",
	},
	"cmd/internal/obj/x86": {
		HashMax: "e00c5f2c0f48ee0e16c61cce057f8017051bb935",
		HashMin: "cc5f14676564e92dd753c80d704f3458b9617895",
	},
	"cmd/internal/objabi": {
		HashMax: "d1fd20c8573944769f071a703e662ed5f50498ab",
		HashMin: "d931e9d4184cb5a868189d6ff58c32d1471f2d0f",
	},
	"cmd/internal/objfile": {
		HashMax: "2eb201524649befac7489740b54a37fc4fabf827",
		HashMin: "6625fb5261f3bc653dfebbc9d6235cd9101e009f",
	},
	"cmd/internal/src": {
		HashMax: "9ab0fc20197a79ed8709a276e1b5ee8ebfe0b65b",
		HashMin: "dda785c428c5b211e9ecbaa474e53b4cda9dae1f",
	},
	"cmd/internal/sys": {
		HashMax: "e81e37782d6b3dbb2e8cc895d0a76b5ebfea5e07",
		HashMin: "3314c68cfafd44efbe802195ec734a2e14c53209",
	},
	"cmd/internal/test2json": {
		HashMax: "48684d3d00bb8ea5bd5d1a4678a590623328bd86",
		HashMin: "d8db2f6bcc83c3cf51929560f3d3e8a0db4b1204",
	},
	"cmd/link": {
		HashMax: "af41d53c8c9426795df793c60676a90e7dbd32b7",
		HashMin: "da15e83a49a33492bcfdcd55d86b4be8dc51e771",
	},
	"cmd/link/internal/amd64": {
		HashMax: "6b4a61337cb4388ea56c48d53ea4f42af3e90a1d",
		HashMin: "a4d6ecd5a8f98f6d7b410f82af9de08a9ccd42e6",
	},
	"cmd/link/internal/arm": {
		HashMax: "7cfd5031282f6a598669d958736cddd345bf647c",
		HashMin: "882dcc7c084cfd94ab53df616d7c010dd28efc2e",
	},
	"cmd/link/internal/arm64": {
		HashMax: "b714b400fdc88203d254c5f22f10cb8ea073e98c",
		HashMin: "09da05dd0e28ac7681ab9a14097aa43355e6cc73",
	},
	"cmd/link/internal/ld": {
		HashMax: "23db5b4ba28a0311a970d5f3a725a6e9165e94f8",
		HashMin: "343a10d6a725108efe4734dc56575a0df1587a70",
	},
	"cmd/link/internal/loadelf": {
		HashMax: "b5c8f358d9536eb901128f3dd4e5ecf85d1e07b3",
		HashMin: "bf2cb8321a6ea801b43020b77b1ceb67d1564c43",
	},
	"cmd/link/internal/loadmacho": {
		HashMax: "053c5f0fa033d126c458aa9b97ef5ed8d1f44973",
		HashMin: "f85f2c445b99d672b2525a167fb45c9b26626215",
	},
	"cmd/link/internal/loadpe": {
		HashMax: "dc0f5c5427709072c24a6c6edff93927d94d5739",
		HashMin: "a51b6eea6c6e5cb4e0b45d0a790e8b20775f9abd",
	},
	"cmd/link/internal/mips": {
		HashMax: "08b77f86a4e3d427e39928eef8fbe87313079276",
		HashMin: "e16ba44d7c933fd52d236373eadda2e05287b521",
	},
	"cmd/link/internal/mips64": {
		HashMax: "d93b01d801e5ff53494f0467634497998fed1ed1",
		HashMin: "ed24afa3df02550215bdb011b47cdc6c323929b5",
	},
	"cmd/link/internal/objfile": {
		HashMax: "fc335a77748073c16c84f0e59e93cfcca4d00b45",
		HashMin: "3a66c4e1c89bd2d8e7e9cec90e3832effb961b10",
	},
	"cmd/link/internal/ppc64": {
		HashMax: "9d45b047cc39fea9f3dbf7be6ff8d79832901a39",
		HashMin: "b1f36600e80d62ac543230f8403effd1aa2e8c36",
	},
	"cmd/link/internal/s390x": {
		HashMax: "9474db5836a79c019a8b00f241fccd9b1c8f1c35",
		HashMin: "3615d82d9ea6cc077579289dda9ab686e80bca7e",
	},
	"cmd/link/internal/sym": {
		HashMax: "add3ffc4413ee93f91ea1f2fa66a5269ba2c4aa3",
		HashMin: "b84cbf10a5daaf7381fecab803f8ecb9e47f2514",
	},
	"cmd/link/internal/x86": {
		HashMax: "96e1a51f5afdef9e54b846955990525ccf870eed",
		HashMin: "f715c1a3ec61dfed37ff04e6703ab185395b531f",
	},
	"cmd/nm": {
		HashMax: "f57a4feaebde90d1661461abdebbee6ed35ca3b9",
		HashMin: "d96cf7d685c3d3516aa2223faad3e01bd3d018b3",
	},
	"cmd/objdump": {
		HashMax: "9b74ce3dcb0d5eae8eae6307287e032a2287876f",
		HashMin: "758ccb1e37ad9bcccd340fe571f4710576a2d3d3",
	},
	"cmd/pack": {
		HashMax: "d3a0e65c965584740d1c3ae194b34911c14354a8",
		HashMin: "a74f6f25997df441c4941f8c49700934efb26b16",
	},
	"cmd/pprof": {
		HashMax: "f5458ed20ed4a85c38362eb74f552b7641913297",
		HashMin: "eaff8050b84469b030f7e01f5441f721db5944b3",
	},
	"cmd/test2json": {
		HashMax: "277deb3a6e2355a8d47ff7d45420943b0798957e",
		HashMin: "3ceda72b6a2e803fcd4bfd531ef5b58f054a4b70",
	},
	"cmd/trace": {
		HashMax: "c86dc939a6cefd9f9aefa13350be34853e67375e",
		HashMin: "88bd767b1fb78dab0e3405228c7b6b370bb72ceb",
	},
	"cmd/vet": {
		HashMax: "ee375d571268638df2ff0030291ecdd551b370a0",
		HashMin: "bbfcd654c75cfd032242293d918e332727f14468",
	},
	"cmd/vet/internal/cfg": {
		HashMax: "1a3c51505d807431562478b9d29cc3bbde4f6b9e",
		HashMin: "b4d514f087353a7480dbcbe87e08965311cc7c38",
	},
	"cmd/vet/internal/whitelist": {
		HashMax: "4705ecda11b5f48c5f77a7b51587b61f0e60410a",
		HashMin: "f2f1f4b5e360e69632c549922ee08af6f35fa77f",
	},
	"compress/bzip2": {
		HashMax: "35c0e500103dfaa7de477d638c9b124e102786d7",
		HashMin: "e87c27ce7c13a81b309a7f394420a36710ff3d1c",
	},
	"compress/flate": {
		HashMax: "8be7c2b0d0f7d41dfe5bc8fed0384cc81560340a",
		HashMin: "88127c69af18c27746bf71c9c96b5b40ee713ff4",
	},
	"compress/gzip": {
		HashMax: "ff4e75936813291f71756fc1aa579bf5126b1dca",
		HashMin: "f179185368ab93f353d47a80f0ea65b792ea3456",
	},
	"compress/lzw": {
		HashMax: "d66261ff2173f9ebd0a83673238bd99486e5d5f1",
		HashMin: "cab168573a264b74220b87530fb42a6324a0f5f3",
	},
	"compress/zlib": {
		HashMax: "02ba10d585af2eb73640b6b73cd026c7fd9da80b",
		HashMin: "220676c953efe651b822188e2526dfa68b70cbed",
	},
	"container/heap": {
		HashMax: "3af8c3bca901e7b0ad60e9023ea6b2231a952a60",
		HashMin: "17cab0b0d7a6567b58ddcdd2d5892ec0a1eaca10",
	},
	"container/list": {
		HashMax: "87869951c1eb94765939f852e8c3347ba9620c45",
		HashMin: "5e66d35b38e931e136f3556321a1fcbb3340845a",
	},
	"container/ring": {
		HashMax: "813382204a675121a4cf37790376b99593855fa7",
		HashMin: "783177ff6cd358aef050082184b8d48b631e2c0b",
	},
	"context": {
		HashMax: "bb730a6cfafc898779de6da07fcab2f6b3ee463b",
		HashMin: "a5a65446f7618f8cfd7f8fcb8c46243146b23bd0",
	},
	"crypto": {
		HashMax: "3627c6d889db16ce33e566b25ca8b42c8f609368",
		HashMin: "24f7dfe5c434a016f13337693f12e7ddf941193b",
	},
	"crypto/aes": {
		HashMax: "97788263bbdf6072b255b23908663f71f642023c",
		HashMin: "ba51a4164ae8e1eaa520a640f2ecc1436a33ef5e",
	},
	"crypto/cipher": {
		HashMax: "fecc739ac7bffaa6acbe4d2248313d70d0b2e7ce",
		HashMin: "15723eac1799a3295cc748a62738f16f5311dad6",
	},
	"crypto/des": {
		HashMax: "b1c4adfff12d89ed8cbee1d9d57ae7581649fa29",
		HashMin: "8f6b198b6a8022a0c54a36edd301e8626b3fb290",
	},
	"crypto/dsa": {
		HashMax: "5b6ee88255d78de25279f7d3380e307710a5fc77",
		HashMin: "5095cd33513416d26bca3cd9ef4435aaaa68d9fd",
	},
	"crypto/ecdsa": {
		HashMax: "c3311871d032a6f04c438d4b44b825551716d07e",
		HashMin: "6cbd92f57c8806a994c68f78972a383b149a84f9",
	},
	"crypto/elliptic": {
		HashMax: "7603b1d6026934535fb88444f0884a5d82c9b715",
		HashMin: "dd3d23f29cdb1f80a679052d42d6ae881361826b",
	},
	"crypto/hmac": {
		HashMax: "592d87c87760ad74007ccca68b6bc7adfde5d3ed",
		HashMin: "4247b03fc21b03cdd7a65f2b1dd06e5ac6b08094",
	},
	"crypto/internal/cipherhw": {
		HashMax: "dd92b1f432010ea68184adb0f70ba6b2ab7d27da",
		HashMin: "d85d18bab23b914cd5d43fa922dabb29fdaed153",
	},
	"crypto/md5": {
		HashMax: "583afc2009ec3c3fbd997c7497607ed904e3fdf4",
		HashMin: "70a3f92342fe3e9a2adf5b34b83ec0eae2cccacf",
	},
	"crypto/rand": {
		HashMax: "d3ef51818e2d845e86a5cc1b5c0aac7fa0e14c51",
		HashMin: "97cf272f515b3ccbd3172d2cd356c285ed69c791",
	},
	"crypto/rc4": {
		HashMax: "6e6493aaaddb727acc606bdc98112293370d931a",
		HashMin: "477d30ceb86e4d050938fffc30f92397c4ec9cfe",
	},
	"crypto/rsa": {
		HashMax: "225ec3d29e1a2491a7c2c9f1d2004a79c927dc1c",
		HashMin: "4884b2f147bae789224b003df7d8f8f1330b81ac",
	},
	"crypto/sha1": {
		HashMax: "9d02cebfd590eafaf09ed3ba4254680d29d2b434",
		HashMin: "e5e5adf26a663620b3a3a928bef1ed537dbf45d7",
	},
	"crypto/sha256": {
		HashMax: "d64891570d623ddeeca84464772dbfed521607f3",
		HashMin: "2602244ddc250fc512aaf9a772fbb513fb84d5ed",
	},
	"crypto/sha512": {
		HashMax: "473a4a2f108778f5b97d4b5cdadcb15d4176239a",
		HashMin: "d145748e62b7132724adcfeb4dfe32337b382121",
	},
	"crypto/subtle": {
		HashMax: "a22090f4110edfabf5918a46835608c49b10372a",
		HashMin: "42ad0ea07cf40d63550c7b4f50384bf047d09a05",
	},
	"crypto/tls": {
		HashMax: "ee617cfa6dc1d7536c19914dc94103b041d87b4a",
		HashMin: "a81d0ce38710c0c4650483e9afcadfc72057fa18",
	},
	"crypto/x509": {
		HashMax: "083c735b9c672775a4bb27979d09010e0370059a",
		HashMin: "b7d15032993fd3f9a7bdd728220ed1d5874c547f",
	},
	"crypto/x509/pkix": {
		HashMax: "4db9110f44db3d45aa8e460fff111b9ba1737e08",
		HashMin: "3ffcc8dd6461bb9da1cde86ede4c4fe356776604",
	},
	"database/sql": {
		HashMax: "071791fe7580263219ab1d168c751edd1ba7377a",
		HashMin: "573c84c7ec35f7a76f9a12f035dc2b0363f77086",
	},
	"database/sql/driver": {
		HashMax: "ef40ae3ffb4199303abce4514e6aaa8456755536",
		HashMin: "3550ff9f766486d16f05fb3a1862582b2a5ba25c",
	},
	"debug/dwarf": {
		HashMax: "f992f25e01e05d6ba1a612ee7a20a7d4c1d1d1d0",
		HashMin: "2d9ef9eb4614e2e661620aee0cd9e2feeaed3b4e",
	},
	"debug/elf": {
		HashMax: "3017f27eb2248ace3f4919e46b71458596d24e6d",
		HashMin: "5577cf8cdc941513779d489d8a94059619b39f9a",
	},
	"debug/gosym": {
		HashMax: "b12a5ebbec1c8803978460685fbececa175f1b2a",
		HashMin: "6bdb172ce73ff10de01f59bea673c6f84941b2fd",
	},
	"debug/macho": {
		HashMax: "70f1d12082282cd9d5485b971d30fa99fc47c62a",
		HashMin: "8bd391394514a2f299781ab46fc0df6a4ff84614",
	},
	"debug/pe": {
		HashMax: "8ecfc0013b597bcd7c4deae06d463689614846da",
		HashMin: "991d5e06ee87b3929b802d11bc3e29f6c5142e6d",
	},
	"debug/plan9obj": {
		HashMax: "0fc092094925b2549795dba0162a969b997b646a",
		HashMin: "dd7f054eac52b5bff37c9206265244e279260290",
	},
	"encoding": {
		HashMax: "143404db3ab9eca915cab1bb864051c875ad4e89",
		HashMin: "1811b6814ac6550df6abfd50728c78181c1f4c75",
	},
	"encoding/ascii85": {
		HashMax: "5a3e7462249b9a89ef55a576af45c783b65a2643",
		HashMin: "2206af09c1c61a5dcf178b7a3d37579ef8f45e01",
	},
	"encoding/asn1": {
		HashMax: "d8f601823ebcf986b527fd32e7fd7f0557dcc36d",
		HashMin: "268781c4080b487098ecc6f4139824cb383a1f12",
	},
	"encoding/base32": {
		HashMax: "50e1eb151054935e2d08de178d9cec95fa2a863b",
		HashMin: "4ea186b8017a876997f6eb15d450ca64cd2d2ebb",
	},
	"encoding/base64": {
		HashMax: "570ada3661d1db25cb550b6e7c4182e575134236",
		HashMin: "89f2f8c2c01b968ee5c05db54f94d82e54f05698",
	},
	"encoding/binary": {
		HashMax: "7c5ad27119f7918b2d443ad41238340617033485",
		HashMin: "91adbee6f3556726e98b48b22a0999808b312ea0",
	},
	"encoding/csv": {
		HashMax: "1bfc8f039e321ed35eb7784b24893649a8177307",
		HashMin: "20e7f1b8ebbdb9a875a0af57d0b14ad80e57534a",
	},
	"encoding/gob": {
		HashMax: "c4e0ece70d4d2a59b2db66186b548c245e5bbd5a",
		HashMin: "5ca1fbef186c6794a8d8ebe089058880e1b2b87f",
	},
	"encoding/hex": {
		HashMax: "ae909d155b71406a229810a24bad40f879127107",
		HashMin: "762bf66267b44bf2bad6b16e521457a71b6a1aa4",
	},
	"encoding/json": {
		HashMax: "9937419d5f0c44c1f9a72d8dabc3a718415744a3",
		HashMin: "090d7869a2786625cf562d25b78623e6eb5050b4",
	},
	"encoding/pem": {
		HashMax: "ed8ff201aeb4cb84dd04b7b5250f01e017718ba8",
		HashMin: "c3a15365fb4d689e9e4cd9057d7506964b417067",
	},
	"encoding/xml": {
		HashMax: "5e9557d5d32817e56a42c0ec8801ef2990190259",
		HashMin: "1a122dc0c699a5b460b10124e5295ac9a921227d",
	},
	"errors": {
		HashMax: "886e1794220ce12c736a2353097f1ef03d402fc5",
		HashMin: "fcf1373327065f532a6495922919487ba0f6b76a",
	},
	"expvar": {
		HashMax: "f00282550995d43a1a33919cc36f4626a05e2ca5",
		HashMin: "a8026b2bc8657e2f256350e4ec42d4c46e84ae6d",
	},
	"flag": {
		HashMax: "7ae6dda1613922af60e696288b1c54085b26cf59",
		HashMin: "3bf76f309e56662dd91fb596d556ce0340da3cb1",
	},
	"fmt": {
		HashMax: "8616bd00c6257cdf3bec21b8b2772325abd6de5c",
		HashMin: "5459137b2b97cf1eaa52c12fb7bf0024e0cfd229",
	},
	"github.com/google/pprof/driver": {
		HashMax: "838f8adf4de49a481ceb794fd3bd8871ccfef895",
		HashMin: "6f7cb40cad7cc45d6e6d9798416b776938a634f5",
	},
	"github.com/google/pprof/internal/binutils": {
		HashMax: "2540570aa2f15a3003344329a5f176e8a9831a27",
		HashMin: "5f762647fc4f7bb95c70a91a341fd80345f395d2",
	},
	"github.com/google/pprof/internal/driver": {
		HashMax: "ea0cfa0bb160e1a2f1c5eb4edb3a112785aaa120",
		HashMin: "e864e39e2f47ba7ae9f6bc130809367cf18cdef4",
	},
	"github.com/google/pprof/internal/elfexec": {
		HashMax: "a67a33a42a74ada5418f037c4d3abee8905841b7",
		HashMin: "7190a458b62b4a0e40d8d21627ecb5b206de87a4",
	},
	"github.com/google/pprof/internal/graph": {
		HashMax: "7e97e57dc278248c20e982f8bf04cde14a4c14f7",
		HashMin: "1714e4bb02e29b4a3597613a3b57900a0812d01a",
	},
	"github.com/google/pprof/internal/measurement": {
		HashMax: "8d1bc39cfd22cd3aae300858d149c9d9af751c9c",
		HashMin: "5e877d3be77790e45ea4f29b9a0268365629d6df",
	},
	"github.com/google/pprof/internal/plugin": {
		HashMax: "3a164574a84a5c5f5751fda99e863f8ebd375a43",
		HashMin: "d5dce14f05c73e427d3a382306af2c63864d0398",
	},
	"github.com/google/pprof/internal/report": {
		HashMax: "a29d3f757efc6ecc641b97903ba33777461c8984",
		HashMin: "53e750bf80b5e7af9ad5ed688c07c8911803e0e4",
	},
	"github.com/google/pprof/internal/symbolizer": {
		HashMax: "4d393927114a90ff2adfd5c90ca0309461fe3c57",
		HashMin: "e591f47006f364bd4bebbf1e22ef7adff009d146",
	},
	"github.com/google/pprof/internal/symbolz": {
		HashMax: "773ac2291a0e07120b344adbc793d9a7781b303b",
		HashMin: "071cfe4ca4daeed1d880a0beac673788069de143",
	},
	"github.com/google/pprof/profile": {
		HashMax: "2b0afda721f2d1485a7c7470c6eea25abb0940b7",
		HashMin: "566d348b8cf1a3464ae8ef3d2a8935ad9e3ef0b0",
	},
	"github.com/google/pprof/third_party/svg": {
		HashMax: "a521a0f391eeb4d2ebfe108ba64f39e4243d598c",
		HashMin: "613d5c8ae69127350146b3dcab18e3ab50355727",
	},
	"github.com/gopherjs/gopherjs/js": {
		HashMax: "e17e9f2ef094b1b68816d4aa6dbe1fe69e23c4ac",
		HashMin: "39354bfa4a1f9eb296652d4dad3ebe086781c369",
	},
	"github.com/gopherjs/gopherjs/nosync": {
		HashMax: "cb1117653705f66fc625c08e6878bf6f8b5813f1",
		HashMin: "7b69f2e75543aafbe7dcf126b72e01b91bb1ca6d",
	},
	"github.com/ianlancetaylor/demangle": {
		HashMax: "fedf977884692a029eef057a88f1e2a02d9da1d7",
		HashMin: "1476301864314ba3b2b019015590bbda39f35d43",
	},
	"go/ast": {
		HashMax: "f827cc00fcc80df84b02fb4d437439d80d901a3a",
		HashMin: "df1ebcd19caeb16ae67b64f0934cc4dc53bc88a1",
	},
	"go/build": {
		HashMax: "7a9c9e8d2df8fbc1a7a8833ddfae9f3c7c374e5c",
		HashMin: "d5314e04f121e96cd52a2cfda1eabf5cab1dfde7",
	},
	"go/constant": {
		HashMax: "3be5ecbe956f06691d0de1aaae9620e2ca789f80",
		HashMin: "b0bb4dce1f533f0707f7d1df519f45f47243e308",
	},
	"go/doc": {
		HashMax: "1334baf26b6cdb6d87bcc41bed7e5819b8acae12",
		HashMin: "c903d65d2fb3a1566bce43e60a01c8eca5b33fa1",
	},
	"go/format": {
		HashMax: "c0519e86e1ab6251a0a4ace0587bc0c70366ad90",
		HashMin: "bc1e81d441bdbea07e4032333876a3a1f221d865",
	},
	"go/importer": {
		HashMax: "654f12f5e97e57e92b03fc13dd9190cf32cdb3e1",
		HashMin: "17c0014f567eb0bf40b7cd733fd5494b51814ef8",
	},
	"go/internal/gccgoimporter": {
		HashMax: "d46b86ed6b19b01c00b692126eb38dd0342459c0",
		HashMin: "92ea22d835d413c2b4d9665605d2eb81d6998bfe",
	},
	"go/internal/gcimporter": {
		HashMax: "646fe9f6346384d1b6d294f8ec07c2b1268f4c65",
		HashMin: "3cb99d0696d5fc5ed1b3e79fe9ada49ac9176b03",
	},
	"go/internal/srcimporter": {
		HashMax: "3a2df023037d9e776a169a44af12efe98adce57d",
		HashMin: "f9377ed81dc75f253408e2eb3f5bf6c858642e3f",
	},
	"go/parser": {
		HashMax: "6c6aa858add211bfae2d22c3aa6d236422448b08",
		HashMin: "4424212a8114e6df4079e2e9bc3dbbc2a12818d8",
	},
	"go/printer": {
		HashMax: "2c3bffafb881d48abbe91b0f6170034ce84ec3d1",
		HashMin: "5afa0ad690165aeda89777291d1ab727e00b0115",
	},
	"go/scanner": {
		HashMax: "8047084ca78344d33ec51df84f147b748fb29e7a",
		HashMin: "3b57b2f41f61d7eea40eeb5c1f593255dcfd8e58",
	},
	"go/token": {
		HashMax: "5e6ec872db9689312f224c76225b79e69fc4341f",
		HashMin: "161e2a27c7f1937ed331867643ddb2157021875f",
	},
	"go/types": {
		HashMax: "a83e70eb0aa6fe54c82dd4974a52263ee615a0e0",
		HashMin: "116ca5237d6774ddb4f8c478f016e53b353645d0",
	},
	"golang.org/x/arch/arm/armasm": {
		HashMax: "28ec95612fa23b58154574c062d1ab0cfb9fab33",
		HashMin: "ebf5bec454c0a509eb635ba3530a28b6fe00af8b",
	},
	"golang.org/x/arch/arm64/arm64asm": {
		HashMax: "f9ac74b833f15658cc666d07c64ffa9a12d33350",
		HashMin: "8920c2a1fc9eacf7ca518fb204e4cc0cd42a6cbd",
	},
	"golang.org/x/arch/ppc64/ppc64asm": {
		HashMax: "a9a32224f2ef4aacef582dd50fca4f07b658a4a4",
		HashMin: "c47d33438395dfd034c0681e2da42e8613fa9423",
	},
	"golang.org/x/arch/x86/x86asm": {
		HashMax: "e250d587d52dfe1072ddf1db5e5e99cbda3510e2",
		HashMin: "e8cb573d5082db5994f23fd23b2a9670c3f67e9a",
	},
	"golang_org/x/crypto/chacha20poly1305": {
		HashMax: "158a3be7ad003b3fbac39f8991ddbac41d526255",
		HashMin: "bbd59c273b1e35c522b7c702d82643de01757dc1",
	},
	"golang_org/x/crypto/chacha20poly1305/internal/chacha20": {
		HashMax: "e260b9ad4eb715cb385af0afd7a13fd599a2a23d",
		HashMin: "2254ddc8e0334badd2ccf67f52f01354be288e84",
	},
	"golang_org/x/crypto/cryptobyte": {
		HashMax: "933af7f40063059269436a7b0913a5be81699763",
		HashMin: "4710ea1609c65e924bf61931e1df6b7c55cc2a41",
	},
	"golang_org/x/crypto/cryptobyte/asn1": {
		HashMax: "96824d927cc566ee8dd42d462988f60a8cf5b818",
		HashMin: "1ddb4d0a039a0211fffe37e373b467cf64d701b1",
	},
	"golang_org/x/crypto/curve25519": {
		HashMax: "736649f2dca06e3438758681ce9c712e0e49e120",
		HashMin: "3ec53a1655cccdfaf591a9716366286bc084f494",
	},
	"golang_org/x/crypto/poly1305": {
		HashMax: "d2cbc2f86b02d2d19810940771e8a936dc1135a9",
		HashMin: "9b672cb8968f1a995507fb0dd66043032c14be4b",
	},
	"golang_org/x/net/http2/hpack": {
		HashMax: "352efdfd1eeaad453334de673a279ca51ba4a0df",
		HashMin: "11c6351c6e9006fac2296dde8ebf38307c278f05",
	},
	"golang_org/x/net/idna": {
		HashMax: "ccc5826b6a6b7625cb8afb72c8a56be13b448b56",
		HashMin: "8151d9bb1de1f3e8fed4cf29c69c52c9524db8f3",
	},
	"golang_org/x/net/lex/httplex": {
		HashMax: "567313d98e73b5b72af2e1daa43e408925c91ce2",
		HashMin: "567be060a07f6f521393af076355299886f893d6",
	},
	"golang_org/x/net/proxy": {
		HashMax: "38a2509475e52d5426886766237026c48f50d03a",
		HashMin: "6f60c7841f34a525501a0596834650496ff7d2eb",
	},
	"golang_org/x/net/route": {
		HashMax: "ab8ee5e14ebc07e21120b356bd54c8d7ce33b4aa",
		HashMin: "816b2da55f633ccd1c4fb4c3b9e1810d12148493",
	},
	"golang_org/x/text/secure/bidirule": {
		HashMax: "3092270e2cce97d2482240b3fbd1f8dd403d692a",
		HashMin: "22b9f4ae1729d5540842edbfdbff008170a13e43",
	},
	"golang_org/x/text/transform": {
		HashMax: "ca7ffeb6f50548142374341e241870a2df2e89d5",
		HashMin: "d39ed0bbdccf845450b73dd5825814c6bca62787",
	},
	"golang_org/x/text/unicode/bidi": {
		HashMax: "ea07026c2ce668bab5f5e6a4c747b52809c90343",
		HashMin: "1aad1dbc5baad968c67e373f1765b51c1be767bf",
	},
	"golang_org/x/text/unicode/norm": {
		HashMax: "6cd36cfff5768c57252aa9f6cbcee4a3074122f3",
		HashMin: "a78ff99902b9e5b72bbda2e1dd2f1f3a81c82c3f",
	},
	"hash": {
		HashMax: "e329c0e33783da782d2383ec4d0d16d30f67892b",
		HashMin: "089192c7e86e9e29a8b8250f1d82d364695f2929",
	},
	"hash/adler32": {
		HashMax: "643565f88fc5c6e7c8403aa19340e7846b8442cf",
		HashMin: "cc1a2745cfcfe4cc54dea02f7a3aea60b6924eb6",
	},
	"hash/crc32": {
		HashMax: "bdda3df0563a40c3e4ff4e985b77f571c68ae4be",
		HashMin: "5218d030061047f0063894108e8079c6b7a71f4d",
	},
	"hash/crc64": {
		HashMax: "d68ec4056784d844ff5cd0d9c63b759d1da05c1b",
		HashMin: "f9e97747ad5f1f7a99620cdb519e1d461162abe1",
	},
	"hash/fnv": {
		HashMax: "f39947aac7ad2942139f28153b117c2f840530d6",
		HashMin: "d3b78b9f26c92110e6c5409122f435ea519b186a",
	},
	"html": {
		HashMax: "719f600b345be5c2c4be25b349cc1d847a431fc8",
		HashMin: "fa0c758d50e488338e7506fdb5f706d2aa4055ff",
	},
	"html/template": {
		HashMax: "fcddd6a12116ddc42fef576a7cab2f85e1349599",
		HashMin: "cca6742b05096f8c8cb1108d741e597788ec2932",
	},
	"image": {
		HashMax: "34a465ef7e86aae83ba17c3a5c377bffe7bc386f",
		HashMin: "9eb25575a4477d9c727af69ce6abab046356f9f1",
	},
	"image/color": {
		HashMax: "d78fd1b07064e4e067e2a87cb11a873e625ca5a0",
		HashMin: "e441f3f9d72629b107ef67bbfdfa085bbb8b1ae9",
	},
	"image/color/palette": {
		HashMax: "0de8f292d07e6094d805ec810bdef501778cf1fe",
		HashMin: "c636e93134517345f43e5fa90a2012c490e8afed",
	},
	"image/draw": {
		HashMax: "e81b7b737cbada078c2f20497f26ef6cb3f3ccab",
		HashMin: "6e10d4bf2e4835010d53cb7e1f9f405026c74d84",
	},
	"image/gif": {
		HashMax: "55d81bfb231413340b706304e4348a7c5c19a428",
		HashMin: "4583775d2dbe5dbf821390fcd21b04bb06ff289a",
	},
	"image/internal/imageutil": {
		HashMax: "edab6640939a1d4ac39c6796b57f40fcf920e6fb",
		HashMin: "1bf9fe08bdf2009b831579c3337af6f38f3b6ddf",
	},
	"image/jpeg": {
		HashMax: "ea33b3989148401eeadafbb291b3ec0fea493070",
		HashMin: "e670f5fca6e95cc506a1565a0ddc21aa8d5cdfd7",
	},
	"image/png": {
		HashMax: "6f8ef452e76ebc5046be2b0e21b9c0b46c8ed8a3",
		HashMin: "f1ecf1fcde8c32e10038b77ba7d9761c06dcd212",
	},
	"index/suffixarray": {
		HashMax: "890ed1296740e98a0e2303c543676fa3ce90f926",
		HashMin: "b0d9b0df9fdd8547c0e42a5e89ca0ca290e078fb",
	},
	"internal/nettrace": {
		HashMax: "2fd8d800a101b64c845d0e151128823ac50ce685",
		HashMin: "e9c84746c6725f43758190d87b82b9740a228a9a",
	},
	"internal/poll": {
		HashMax: "b5acbc85467f8475e14097477c4384ecf14a5c61",
		HashMin: "ac5db25177718553ce486ea6bd2840295d1ef0dd",
	},
	"internal/race": {
		HashMax: "4bb49c054eb6aa21f9aa217c6b57e666c7681bd7",
		HashMin: "4b904d32201027c0e8aacbbf5be4cf9c0f597478",
	},
	"internal/singleflight": {
		HashMax: "adbeb331486bac127116a205872988c8e12a18b7",
		HashMin: "c24fe550d86ea741b6924ce008683e53a991bd18",
	},
	"internal/syscall/windows": {
		HashMax: "0440b2d46c7e3c57135c73af5c503bd2acb3ea77",
		HashMin: "89a046e7bb32351a4950faaac943e5a78b9c2fd0",
	},
	"internal/syscall/windows/registry": {
		HashMax: "218bda913da80b0a6742490f20447552c94433f4",
		HashMin: "a4e02b0521e120ddb452a6a7a3e9e1c26189980b",
	},
	"internal/syscall/windows/sysdll": {
		HashMax: "e4b9ab2be819105115d55112fa71a150afb3f3f0",
		HashMin: "ef1f4115504a10511487cfb92f9fe66c1654faf5",
	},
	"internal/testenv": {
		HashMax: "d0b94da10d1080b4bc54f98a22c3958da3fce264",
		HashMin: "271369e45d1006f7d219343bdc8677e9921794b2",
	},
	"internal/testlog": {
		HashMax: "c89b0ab1b63117b7bdde78d16c680333ab6705e8",
		HashMin: "b7dba9ac9afa6ce7497409ea657fa8f045dd62f6",
	},
	"internal/trace": {
		HashMax: "40a3487cc4cf6f2d855ddd93ceec54b42a8f285a",
		HashMin: "525be5d6e18cf3f2bb52fa7c04b84a549c8f4b9b",
	},
	"io": {
		HashMax: "a356ccfed4c84bd032b9c833db40fe184f7c3aec",
		HashMin: "50ac6c47220a6a85f3ee127fdcbc93a3af79f201",
	},
	"io/ioutil": {
		HashMax: "0b654b63f14d91ffef3819644e7f8f8f728f3ad5",
		HashMin: "0c67d95af16e2e1d05f5bbf8efe41f24d34d1ca4",
	},
	"log": {
		HashMax: "069aa90045e4ec36abce2942fc21c03da52efc32",
		HashMin: "df725f7c78cd8635da2e17b89c3ba7f332578038",
	},
	"log/syslog": {
		HashMax: "076b0550faaacd946e5a8f54be8a8713336b1752",
		HashMin: "eae41a90a491dc634bbaeb566aa96fd0bdbcd9d9",
	},
	"math": {
		HashMax: "83c89fbd0dd522ebecd0decc08888d4a6db35565",
		HashMin: "56402b8c14f7e1a5df1ee0e7997b6c9d7504f8c2",
	},
	"math/big": {
		HashMax: "0941cd3d717cbc6216f752b8cb1e2a6a8aad3190",
		HashMin: "d8cb19a6d8bc97fedeef77f1686a440eca667097",
	},
	"math/bits": {
		HashMax: "bc287d1385da9be1725cb05c9431887d23639461",
		HashMin: "f601cad97a89015871e3ee769167c8dec5605432",
	},
	"math/cmplx": {
		HashMax: "96fdf25495b492854ec5c4bf91f74ea14abd517c",
		HashMin: "3d28e52a49ff8ad668d52b142edc1c7417206a51",
	},
	"math/rand": {
		HashMax: "6c0eea276107c40b0fe6ca6db3267ad2e722e636",
		HashMin: "a5039d85ff05c6b76d87439ab94ebe3272096b85",
	},
	"mime": {
		HashMax: "b1d339c7bc3114187dc1345cb3974c6c9ca1ae5d",
		HashMin: "ee3b165b989167bb6f95953e2e2acb1cf485b061",
	},
	"mime/multipart": {
		HashMax: "c51db198cc95761e4a716ccb8a00a20825cb8e7e",
		HashMin: "fa9d60690f7a9f0554d19ca4a89bd7d1e6281de3",
	},
	"mime/quotedprintable": {
		HashMax: "caf88113afe6f49ff622ed21e8ff275cb2d9e6ba",
		HashMin: "7b236f837ecd01f6c62e6cd70e73ddb3bb3f84f6",
	},
	"net": {
		HashMax: "82b2bb1f377c1ea2bb377c0c7b9ca8ecc4f01fa5",
		HashMin: "406c7e8b1f4e43a69f865b2a3fed8dca0e47504d",
	},
	"net/http": {
		HashMax: "a952952d2127536b22325b7ac442d5f393a07b4f",
		HashMin: "289dfa4d8fc563caccff6c1c1536f9afcde4a8a3",
	},
	"net/http/cgi": {
		HashMax: "0c57295bc3c285dd973ac1a9d553af8eb946f571",
		HashMin: "079a14d8846188727ea43d8973e62e3a218be913",
	},
	"net/http/cookiejar": {
		HashMax: "a3c22142d9ff6a9f3427c69d8879477966d486eb",
		HashMin: "f91e4128ac27dd9091c444787fb79609a6af7b33",
	},
	"net/http/fcgi": {
		HashMax: "3078a6ee171faf004fa6f73c21c2172e7ef6d9c9",
		HashMin: "19a53fb78039ba51dc8fbf9945e38d6945f8179d",
	},
	"net/http/httptest": {
		HashMax: "f6e07c557668ec308091788903eee69263bee4ea",
		HashMin: "a5ae8b765b8525d7b3e4ec3f18c66adfd1423c01",
	},
	"net/http/httptrace": {
		HashMax: "3c74477bb9f9e923cd39153cd01f64d082e1365d",
		HashMin: "6f17e91314abe7a88c557fcd3f945e43a3206506",
	},
	"net/http/httputil": {
		HashMax: "0a8a3dd312447ece75a9cdb42f3f7e6e95399157",
		HashMin: "b3c15a3ff480697983e892b924eab40addcf0c25",
	},
	"net/http/internal": {
		HashMax: "0ec6017bf1d5eacbfb3ba87ad4b1cf2a8c65e1ef",
		HashMin: "f2cdda5c0ec22f02d12fa4aae338ba1291e1d532",
	},
	"net/internal/socktest": {
		HashMax: "ce134b9c67f538a14732ff254896fb18d3adbadb",
		HashMin: "d53e230c2cf82ee3aa40ee54e590ef879c11f08d",
	},
	"net/mail": {
		HashMax: "1b06cbe6fba33b09506b03dde4c64f4a88a87de3",
		HashMin: "e733540f6e4b216c6cbe3ba71310535607ec7d53",
	},
	"net/rpc": {
		HashMax: "1e29fbd0a1ad3837a2ed8ba2278543afa543c8dd",
		HashMin: "e35062facf94abaf393a2cde87d453e9e920aede",
	},
	"net/rpc/jsonrpc": {
		HashMax: "7a9f2dd0f2d6c4d2ae6fde3a96352a66fef8ffa3",
		HashMin: "fca77e7cf35126b3e60cf2e9e6c98dbbe038ba7d",
	},
	"net/smtp": {
		HashMax: "9eebb7193cb1b3192f8e49381906c624d00d8ae5",
		HashMin: "fda572be43fa6d5dff3cef57eaa6211095481b12",
	},
	"net/textproto": {
		HashMax: "6ddddebd284871413d2a2e22584fc9905e04c59e",
		HashMin: "aa1da92904dcd8228f014cbe4dd82b1a6fa076eb",
	},
	"net/url": {
		HashMax: "f559df32ded0906c2b1004052c2f39597c660dba",
		HashMin: "c30a5661312be570b1fb48bee8d718002c91ef67",
	},
	"os": {
		HashMax: "e5ab041d1f34409da2f12e0ffc4b5d18c25fc5ef",
		HashMin: "eda355186b7a1f9b1c6536b3f92ac418cb58f1ad",
	},
	"os/exec": {
		HashMax: "ebb0ad152d903e8477d8c681f0f7a5b905f148fa",
		HashMin: "3c159113d8c6d701fdd6de1ff926800198c6e045",
	},
	"os/signal": {
		HashMax: "bb3c78be12271c70fe1c873857b8aac5d85b3d07",
		HashMin: "4a10f569ddc81a9520e94f12c4924246181f7785",
	},
	"os/user": {
		HashMax: "59462bc7234b0f5d8a28f2c31204da632bc7fbd9",
		HashMin: "d0014bee0e147ab746ce013932c386483be0c485",
	},
	"path": {
		HashMax: "07831caed0377467fea67da31731303a38960bc9",
		HashMin: "549632e142893a702850e27669501a110c13293f",
	},
	"path/filepath": {
		HashMax: "b469d2f481343d844c173ad4ba4575f4ffcefaa8",
		HashMin: "41a29d90379a996352190da4b2bc0e41680b359c",
	},
	"reflect": {
		HashMax: "eb15785d92a6881948cab8fe00d7ad902d9927c9",
		HashMin: "6363e9011e1af1ef492e5afcd4131677dab5ebbd",
	},
	"regexp": {
		HashMax: "022814cec6c7a4807c15ab56e02604c9e25e128b",
		HashMin: "5f55153ea7b0bac5dceb304d401914b7e971ac67",
	},
	"regexp/syntax": {
		HashMax: "fcb4ae624cc0dbb0aafd66aa380ea07d193ccab1",
		HashMin: "5fb3abcac76baf3c6cece87524c57483a199d576",
	},
	"runtime": {
		HashMax: "82c28d7180772af6630a3f6c4324f9f5bb0bad03",
		HashMin: "ddeece056c3bd4f9ec250ef6abbbd3050c9226e4",
	},
	"runtime/debug": {
		HashMax: "a66aa19119358393b4a6de1d691e0f64330e6aa5",
		HashMin: "d1ddb1506f727e312acb848526fb2909ed0f8f3d",
	},
	"runtime/internal/atomic": {
		HashMax: "77637424435070057c6da2c53422aca6b95a1765",
		HashMin: "bcd9afcdef801c8998c190e152922a654353525e",
	},
	"runtime/internal/sys": {
		HashMax: "6aa147a4764955fc2b9f08c87b00a155da6b73e4",
		HashMin: "a6bcb0392db5f0dc426e861e8012d672913a64e2",
	},
	"runtime/pprof": {
		HashMax: "448d5bcb45a316cf321605cfced4e16552db3447",
		HashMin: "3a0979d118263f987bcdfc947e6bae6263786ee7",
	},
	"runtime/pprof/internal/profile": {
		HashMax: "78db24bde403d9369cd54e1d9e0edddf8c4b18de",
		HashMin: "8f34df40166382a61533353116490bc8d2e7c2f2",
	},
	"runtime/race": {
		HashMax: "edf622129d968ed6f15013b836b4545db530b2fc",
		HashMin: "18f278690c499dcf64453954c512ddf8c981a31c",
	},
	"runtime/trace": {
		HashMax: "e23255766e694e51668b88604f630be19869771f",
		HashMin: "b9f2391b9330b3dbc92a9ec32e5b34b8a847f4c7",
	},
	"sort": {
		HashMax: "d5c5b64e6850a7df28539c0540b491f16f409938",
		HashMin: "be12e2b31fdc7e64a036d76e0617dee7385775e9",
	},
	"strconv": {
		HashMax: "04bb5f4747ca0d5d48cd6b096df7c61ea390745e",
		HashMin: "fb78180d5410bd903e465ef8f5ed868975f2ac05",
	},
	"strings": {
		HashMax: "f6c1a32a3d550f26ad13ca0f8fe71fd322ea2eb0",
		HashMin: "b1a846048926e905fbed1b674d77b1423d21d5f1",
	},
	"sync": {
		HashMax: "f0bebac94836c10e78111fa27e3a4f367e4bc8d0",
		HashMin: "fb1a0f415af1335379c070cb7cf49ebd5b4828b0",
	},
	"sync/atomic": {
		HashMax: "ca7fbafeeba47780ba6cc479d89429d010ac56ca",
		HashMin: "668c4ed2afa24b85d3b13bad10d21e7f4073a68a",
	},
	"syscall": {
		HashMax: "fac198ad9a73a55cd967b9037338afee62f00af4",
		HashMin: "f91571697620e7e5aab481cca98d447b63e61320",
	},
	"testing": {
		HashMax: "a3524c1d91b5beee8f0c88c693e750c76bbeb1c4",
		HashMin: "8f9510b39cd7467db3d5b47cc926cbea6d271b2a",
	},
	"testing/internal/testdeps": {
		HashMax: "bda5d793e5228b517d4f379cc7000c121346130b",
		HashMin: "0f289b68d47294d111c4238b9231acf61e0ed08b",
	},
	"testing/iotest": {
		HashMax: "83ea335c9825c934ff86a644d3073f01f8c9647b",
		HashMin: "e7c860e87d61c9aede1fa65b384a26246dea6af5",
	},
	"testing/quick": {
		HashMax: "fdd953202c679b61619fe0b85c811bdc0300f48e",
		HashMin: "d0bc607767a2eb202ad1f125a5132de898b08a91",
	},
	"text/scanner": {
		HashMax: "c23d8387b045fa42ddb5352f90c036764cbe5d6a",
		HashMin: "8e435f8972e043d17183acfef42f8dd2a0c3a193",
	},
	"text/tabwriter": {
		HashMax: "6b50d49f3257c9da6276ef0ce664d4ad160ca162",
		HashMin: "7ba395007afe0f3519c2914f4c311e639904863c",
	},
	"text/template": {
		HashMax: "37e27ef86e53fa83a4709ab01a06286c186b0f59",
		HashMin: "4c6046020a60eb90a16a6a0774bd586aba3373b5",
	},
	"text/template/parse": {
		HashMax: "805894702fe0c8fee935a88d29bb9d37bc911b5d",
		HashMin: "ff8b730b53650f82d8b13be5eead33cef4c7a9ba",
	},
	"time": {
		HashMax: "9f6822457b56b74ce3ab3dbf99d903f0374dc2d2",
		HashMin: "0dea4fc8a578a8b5342360973e5a86e225901d47",
	},
	"unicode": {
		HashMax: "a923d46c00f7505ecf2efdc426cb7dfc3651c19d",
		HashMin: "c290ac7bf1acceb70a84559474ed23f841c5c500",
	},
	"unicode/utf16": {
		HashMax: "cad2c09cbfef4b755810abc40b6c58bb4eaaa7af",
		HashMin: "40cde41561d9378c3c58c8b1831a188250a2b778",
	},
	"unicode/utf8": {
		HashMax: "61d6548fa99e69136f41869f01bdbe4dc28d5232",
		HashMin: "28ed2ed385543cedf0327804577f8e15781553c3",
	},
	"unsafe": {
		HashMax: "5a535455f5561432f06987f3ed4c103353691c20",
		HashMin: "f6cbd69d1a94141ebde424922312b4f98cdc46dd",
	},
}
