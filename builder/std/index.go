package std

import builder "github.com/dave/jsgo/builder"

var Index = map[string]builder.PackageHash{
	"archive/tar": {
		HashMax: "0405504c098a59cfb2689a33a3b8f36b26df9a91",
		HashMin: "298e444746dfd851117438e480f9b47bb0458407",
	},
	"archive/zip": {
		HashMax: "6be3b58b340f9ed726894f322e0dd1cd20de1f96",
		HashMin: "d6ebc3a14e3d1d5f8d7d08e1a34a0fd8bfd53eae",
	},
	"bufio": {
		HashMax: "c320bf84ee813fbc19584f8a1aedb049fe9af15f",
		HashMin: "b1fa0a9358e245aaef2166e9f30d183ca16da293",
	},
	"bytes": {
		HashMax: "c973a46447b6ec05a54fb8750dd3fdbc57f098f2",
		HashMin: "c6afcb2e9d573615ab2a506022920c39d3d593fb",
	},
	"cmd/addr2line": {
		HashMax: "d45e2a2cd499750b8a573471daa2d099cbfbe40e",
		HashMin: "9de495a9541e7a61cc83d2389205acbdb4b66b13",
	},
	"cmd/api": {
		HashMax: "d32af0ad7d6dd2ae10ac81aa05999966277906e5",
		HashMin: "f5871239f5fc5566b0cc9fddc0ac001590e7bdfc",
	},
	"cmd/asm": {
		HashMax: "e91d1cd328245d0a579e3daad790126385e8b80b",
		HashMin: "db49fb30f6687789a3428d7209ffd65216786950",
	},
	"cmd/asm/internal/arch": {
		HashMax: "281c7807526f5c7fb05132a079ec9a4b015d9c1c",
		HashMin: "2294e4998f89ff15dd9b62e97fac520672a54d2c",
	},
	"cmd/asm/internal/asm": {
		HashMax: "70de72200df710148ee80dd885ceb884be9e5f55",
		HashMin: "fbc5abc25b72f544aa332a67a2403731cc15dd68",
	},
	"cmd/asm/internal/flags": {
		HashMax: "ac4bd469854c0841ce245dfd8e99df297f09e8f1",
		HashMin: "a6a2e47ffbd7bfd030f9d6de7b616566d998446a",
	},
	"cmd/asm/internal/lex": {
		HashMax: "c2763730d5bac8ecd6f4966d07cc0c5abfbcbe85",
		HashMin: "39b9d7a4b8635f45a3d97b9538f7e63f25189168",
	},
	"cmd/cgo": {
		HashMax: "c70c3e64e3a4044ee34e5755a719461731eb0a61",
		HashMin: "ca237b69461b2dc45fb6fe106c448315d202735a",
	},
	"cmd/compile": {
		HashMax: "5d8558b4d66c39213d1edfb5a64871c8063674e8",
		HashMin: "735a0cb6db163779a5997e7efec2803fee502f43",
	},
	"cmd/compile/internal/amd64": {
		HashMax: "6baab08b85aa8166b9d20653f3be94e6f507c8b7",
		HashMin: "69cb3a9d5eeaafc6e5ead3196bbfca99568da8cc",
	},
	"cmd/compile/internal/arm": {
		HashMax: "e21af52ba6b45311e20d024e9b589d6b55232150",
		HashMin: "508da629d03f0249793670013eee9a544ec0a554",
	},
	"cmd/compile/internal/arm64": {
		HashMax: "c601fe93aa5ba4464447d737f7c9be3a48b8e887",
		HashMin: "6618f6aeef1de8a5b0f407657d7b14a4956d6c10",
	},
	"cmd/compile/internal/gc": {
		HashMax: "80e20646d1bf9504da295c893770616982cb8828",
		HashMin: "e6e3cbf44e16c90f9fdd4c30142330731424130b",
	},
	"cmd/compile/internal/mips": {
		HashMax: "738704217102217f685b347304a675dcf2174f83",
		HashMin: "46006af7f547ff99edadd149376b677e49a302dd",
	},
	"cmd/compile/internal/mips64": {
		HashMax: "c288ee2e71998a0de7e98a0c97fec1a2d0f08b91",
		HashMin: "f493173c40f63022edbe0d141b9d97f781ad56c8",
	},
	"cmd/compile/internal/ppc64": {
		HashMax: "519dc292e4a44859a69ebb80fd6a5f459edd8c06",
		HashMin: "dcd3b80c6343a99dad052d0f1ebcfe1614f1de9f",
	},
	"cmd/compile/internal/s390x": {
		HashMax: "1f7ff5114974b44d633c42895ba018b66c4f1c99",
		HashMin: "e82e60159ce0354f64ea5a42d6941615eddfa9ec",
	},
	"cmd/compile/internal/ssa": {
		HashMax: "0c4053c9b962fac0fd80715eb5fd75fe81d5dd63",
		HashMin: "31c01d95bca961e0d46895f4dce596a499bf630a",
	},
	"cmd/compile/internal/syntax": {
		HashMax: "ad547d1af3525f0f627e7b42429a97d9c32ea0b4",
		HashMin: "9b10ce8224f10ffd9ee0d0fba24a19898b0b828d",
	},
	"cmd/compile/internal/test": {
		HashMax: "be81a3174788f01ab6bc4372c35f36bbb4b59656",
		HashMin: "0c60886ab4c2be1920d2d6d533d6c8a3f53bebb0",
	},
	"cmd/compile/internal/types": {
		HashMax: "718f1288d46dbfc8feea4d647be06751cdb57b07",
		HashMin: "6df0d79c0f712b657a8bb7cf494b343246f101de",
	},
	"cmd/compile/internal/x86": {
		HashMax: "1506566b6082a80ce4a6b294bf758b2dca955e63",
		HashMin: "31c28334be1479cd7f5c6d6815af2c0409c49e2c",
	},
	"cmd/cover": {
		HashMax: "39369f808447f891cd91e58443f623588855b87e",
		HashMin: "8a8d03a2b98da2bb118bacc377b355bfff8e58d7",
	},
	"cmd/dist": {
		HashMax: "8d6ef7b34dddde999a71a58eba5051d206b5ca24",
		HashMin: "a7e044a1268f6fb66d5c698dba6445b9bdbfb592",
	},
	"cmd/doc": {
		HashMax: "2b6fe8ddf0f50b6a9884f2e39f4fd1b296359c55",
		HashMin: "8fc0656a87e787fa63e271b5caac11ead0b179c3",
	},
	"cmd/fix": {
		HashMax: "bf4eb79e69a21f0411f0be6c1abcc92aa8a3e85d",
		HashMin: "94640ea33d13c84b410f17562b28eaf0a8127d35",
	},
	"cmd/go": {
		HashMax: "74b919231b47d7a22f6650a641863003285e0e4e",
		HashMin: "4c7b9b668ced4c16c326a949e49f17558864965e",
	},
	"cmd/go/internal/base": {
		HashMax: "e5f9e0661ecc6a2346508f907de74d6a977b30a3",
		HashMin: "61e8f044f6cb50dd19c24a657a86a2f18708be36",
	},
	"cmd/go/internal/bug": {
		HashMax: "bc3c6bd95e289b4064c12dd4a4099abd5d63a9cf",
		HashMin: "efeaf83764fa65469d94e9d5f6e7ff99b995388e",
	},
	"cmd/go/internal/buildid": {
		HashMax: "4682d8b0055688612361b428263fd03cf315b764",
		HashMin: "44f9dfe7c742f5dd9585f0ae3a921c09db70dc44",
	},
	"cmd/go/internal/cfg": {
		HashMax: "67062236981811ccbd7c2511c9964c9f6fd778d4",
		HashMin: "73be2466e8329949cad3ac12459ebf5a57819aa9",
	},
	"cmd/go/internal/clean": {
		HashMax: "1cad04b5676e0c725ed817a93a3dad8786e48c15",
		HashMin: "40b0ad75fbc3fbb41178c8cde868a277eb562cb4",
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
		HashMax: "0afd553af5f1791966fdbb7aa79d9b1461a79fe4",
		HashMin: "739b2f184a0238e994a4224924ca1a6daef5a14b",
	},
	"cmd/go/internal/fix": {
		HashMax: "33f92caff8df32319ec1a00e55c60c77c2160d2a",
		HashMin: "ad25a547d3f3e0c57927fa3d294f3dbb1a124d6c",
	},
	"cmd/go/internal/fmtcmd": {
		HashMax: "5a48229eaa69c57c99e0dba5e5d0c87237129a69",
		HashMin: "c779615b97992ed446609b7e037d406b9bc8ecc4",
	},
	"cmd/go/internal/generate": {
		HashMax: "7a23d909f7253e4759b910c6cf872418919b6f80",
		HashMin: "d1ca5a345998111ab29347c4ece072cc01dca178",
	},
	"cmd/go/internal/get": {
		HashMax: "a46f952ab67cafa4509bde7cfcd124d74a471046",
		HashMin: "fe0f1833c0e2a0fae9ab95b43d9a8e638c2b5ce8",
	},
	"cmd/go/internal/help": {
		HashMax: "862a7e1b2699d57a62b1676780a3be47df06859a",
		HashMin: "a17415e308fb70b835c9e8da236f7a34339582c6",
	},
	"cmd/go/internal/list": {
		HashMax: "3fa968d51d8b98e3dccc9d28e5cbd44d98f43d95",
		HashMin: "0e67eab950e363e8392be57903edbab03ca69052",
	},
	"cmd/go/internal/load": {
		HashMax: "bbd880001240a1317a3184e0b881647f475b58cf",
		HashMin: "bdbdda1e4017ec07bfefdcfe58804aae396c42ab",
	},
	"cmd/go/internal/run": {
		HashMax: "a98c65aa9e9eeea8ec7469d14d91dbb2c977a46a",
		HashMin: "e1013957d83b8b783ed75354a865b91abff317e5",
	},
	"cmd/go/internal/str": {
		HashMax: "f718c3990ab8c029ae73104d8a300a12ab66e9b2",
		HashMin: "d3548bcf103d2475a014f02a667bd9d5ae09c904",
	},
	"cmd/go/internal/test": {
		HashMax: "f953463b1e789a54b97b2179f3f797b1ed3d28d9",
		HashMin: "ce1287f5d8fd4e4e105021d235811cd8635889be",
	},
	"cmd/go/internal/tool": {
		HashMax: "380bc59f2b9b7cd56ca8c8cc9b9333d11f5b1e6c",
		HashMin: "3c1874cbe7b93c0f44ddb10249eceb3f04732c61",
	},
	"cmd/go/internal/version": {
		HashMax: "4b9c0cde97e0923f3c36f2b5067b40adfef03bdd",
		HashMin: "d2c8862c5a92895524b86724ab7e9cfe07bd1f88",
	},
	"cmd/go/internal/vet": {
		HashMax: "51440ba2c7cd926a542054bde32951743d83424b",
		HashMin: "a05e14ca3d489adb1abb2c7e55adceacb231eaed",
	},
	"cmd/go/internal/web": {
		HashMax: "6811098a50d231471461547e9f507ec084dfcaac",
		HashMin: "c5db4a62b7e57a176da7cf5f9107a711486e76b4",
	},
	"cmd/go/internal/work": {
		HashMax: "3263b832948709221aa7824709aa37b7e26bb4af",
		HashMin: "a0373ff89473e02f927fe8ca85eb5e35ae41f548",
	},
	"cmd/gofmt": {
		HashMax: "353c998a81a8973b45f6a6f5dd59a47bced81262",
		HashMin: "e4128bab4039e99bf4e3d99618755999785a4d89",
	},
	"cmd/internal/bio": {
		HashMax: "594f85c81864b355476b35d63447f7e547cdd7e0",
		HashMin: "4ae41aea93d7b46c777c752e3602b6b8d5bc34b1",
	},
	"cmd/internal/browser": {
		HashMax: "8f402df0c84ad41a7890f4c228fe349e5350d08b",
		HashMin: "b7e5f4e4156c400a9141ae3df5d5aaac093a503d",
	},
	"cmd/internal/dwarf": {
		HashMax: "db38a9c9e563fc151b92fd084b41092feb7bbfc4",
		HashMin: "a340b6e3ff3eaa1e06063cdf8dc71dd5bfc7160a",
	},
	"cmd/internal/gcprog": {
		HashMax: "debfa7b4217a9d59964c1fee4609a5e1fd1d38d5",
		HashMin: "b8a3fa62d1fe42e11ad388c1dd31ecdfd62e84f6",
	},
	"cmd/internal/goobj": {
		HashMax: "889248589946626b24af6e75c350a147464ce9a3",
		HashMin: "4fc29ece4afeb8633018e6f5f06c68bdb496bf91",
	},
	"cmd/internal/obj": {
		HashMax: "eb9bf1c4086db674f8cfb240c418b29dd866f78a",
		HashMin: "19d7d6e62afebb4b14bff497dc003911ade3a778",
	},
	"cmd/internal/obj/arm": {
		HashMax: "0721ff1c2e2f75d51ffef61ac33ee0d33272fd2a",
		HashMin: "4b0dad0b41be21a18191cf2a7d8d93604eef5354",
	},
	"cmd/internal/obj/arm64": {
		HashMax: "4ca4e182b31086e79767f216f93fedde8a4b93e9",
		HashMin: "78a77350eef48e9aff9a359ff4a789fe4b80981d",
	},
	"cmd/internal/obj/mips": {
		HashMax: "ad9adf13c9153b162afebdd1f59e6c23a5b7bf11",
		HashMin: "c41f138690a64c7f368bfb942decb21f14adc299",
	},
	"cmd/internal/obj/ppc64": {
		HashMax: "703eaa122a47c142a82a201119edec78479370b6",
		HashMin: "d04ad1f9357e1eb459d6f4109041c8bbaac2aa74",
	},
	"cmd/internal/obj/s390x": {
		HashMax: "0a728b67ce80c36653ccbca079299567a90df96d",
		HashMin: "1500ef4480cc9e6e048aef71d9fa808097bf9d97",
	},
	"cmd/internal/obj/x86": {
		HashMax: "d5c5cb8b2164169765301be808f909b70d737374",
		HashMin: "aa190ce67ac622571ed0323b3f90780e02690909",
	},
	"cmd/internal/objabi": {
		HashMax: "8a82b616dbab70958e058415f1db4cceb66ca224",
		HashMin: "2a653e7ad952ed09f0b9fefd44ac2a06d13496a0",
	},
	"cmd/internal/objfile": {
		HashMax: "4d36d61f883710bd0db48c22b339b6737fa4577c",
		HashMin: "e973b264bd2bba3f81e43cf4bff3ced5293ee0e1",
	},
	"cmd/internal/src": {
		HashMax: "3e55fb8612569aa08ba9c80b3786d2075077c8f6",
		HashMin: "4f8c8dc22ed92abfcd0467b6d3c3ae3d305c9465",
	},
	"cmd/internal/sys": {
		HashMax: "e81e37782d6b3dbb2e8cc895d0a76b5ebfea5e07",
		HashMin: "3314c68cfafd44efbe802195ec734a2e14c53209",
	},
	"cmd/link": {
		HashMax: "db15e39c79b7d5f4bbc284e0733822c031a68777",
		HashMin: "d9e3734409d95b09bdd5fce47129da874746e348",
	},
	"cmd/link/internal/amd64": {
		HashMax: "e93d923dca125ed7561bd73f096cbd52df338ec8",
		HashMin: "9b086aecb06a5cfcd154fd903989c59bdccd76c4",
	},
	"cmd/link/internal/arm": {
		HashMax: "ccfacc000134293254b19b727323b56023bb7dc6",
		HashMin: "d8a6054b0e9c64a6fdf4ba6194b7d8bc2e116b72",
	},
	"cmd/link/internal/arm64": {
		HashMax: "13996e9986a25113cb131e859e6cdb6517741d5d",
		HashMin: "956f3562aa60ccc7344165c4f2ea6083e7888e91",
	},
	"cmd/link/internal/ld": {
		HashMax: "56002a77d0ceffaa9cb896e5c936e9413e6bc1bc",
		HashMin: "efe37338f65001c5ab309080ecb97e5fd8bdc4ef",
	},
	"cmd/link/internal/mips": {
		HashMax: "917f188854a29dc12a37e54a16c55aff1f11f621",
		HashMin: "fd319c064e736024da6b9eea468aa950f3c3d3b6",
	},
	"cmd/link/internal/mips64": {
		HashMax: "2bac433e1d2daef84b2d2a8e0131c13b9d00129b",
		HashMin: "7e30452d84addfe70dd8612c6fceca10d2dec917",
	},
	"cmd/link/internal/ppc64": {
		HashMax: "587c499c348f9ddc55ae8e8761328d9aceadefb2",
		HashMin: "0a91cb4cff7fe9928485f61a5ec5a66eb310999f",
	},
	"cmd/link/internal/s390x": {
		HashMax: "fef47a7828849089d0e69b66101205c61bf2edf9",
		HashMin: "7ef71215b6443a6e074d6eac91eb21880a850689",
	},
	"cmd/link/internal/x86": {
		HashMax: "d594d4b6d3247e308df4f7f0f88cc56041ae485f",
		HashMin: "6020f44e5cf027d8a09a6e52883f6d5fd2cbc9e5",
	},
	"cmd/nm": {
		HashMax: "d557064c6a419310b61072de62f6a86fa1a02453",
		HashMin: "6cf2f1462d51507aebeb695aa69656023401b394",
	},
	"cmd/objdump": {
		HashMax: "9b74ce3dcb0d5eae8eae6307287e032a2287876f",
		HashMin: "758ccb1e37ad9bcccd340fe571f4710576a2d3d3",
	},
	"cmd/pack": {
		HashMax: "eeb85194ddf1ea8378646c3aa4f50fbb11ac6b8e",
		HashMin: "78264315280e5b2ac28c55b39aea2079bba635fa",
	},
	"cmd/pprof": {
		HashMax: "f1cb224509665e920cefaec0df16619b4ae2f7b7",
		HashMin: "50dec933c340d5196e47464d5536247f25587c68",
	},
	"cmd/trace": {
		HashMax: "a5c256d7ee9d6f67fabaf44cc852bd8430d6db7f",
		HashMin: "73d18ff48d403b77b4601056a5d56807b58dcd5c",
	},
	"cmd/vet": {
		HashMax: "89863a75e0bdff3743a2dd1781c76f7f028c280a",
		HashMin: "0b50298546051030e3f8c19b9f9d42f29d57e7a3",
	},
	"cmd/vet/internal/cfg": {
		HashMax: "58b7394b24dbe43f05eee9acb896d7a0e9c261a8",
		HashMin: "354eeefd5dfc1cd20fa7a19a7ac532a501811191",
	},
	"cmd/vet/internal/whitelist": {
		HashMax: "4705ecda11b5f48c5f77a7b51587b61f0e60410a",
		HashMin: "f2f1f4b5e360e69632c549922ee08af6f35fa77f",
	},
	"compress/bzip2": {
		HashMax: "7e1abd521637daf213f3667900e8af90771cad8d",
		HashMin: "dea9c0cfa042ac1210277be745c70689966ffaa4",
	},
	"compress/flate": {
		HashMax: "d586296e160bdc264a9a7edf478f5c9b1d810f12",
		HashMin: "4d7449750fe5ea7c5897470c1f7953979e3f0da3",
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
		HashMax: "6354895b96a73f3919506ea43088b518c2d5d191",
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
		HashMax: "9520a4b3dbb10ccda66983a00b1b332f96f53218",
		HashMin: "aba6a2750cd44b8beccfe772a7d4608600f8c732",
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
		HashMax: "dbe97ed37f47a029dc4c770903d5f4d2b6229796",
		HashMin: "6cbd92f57c8806a994c68f78972a383b149a84f9",
	},
	"crypto/elliptic": {
		HashMax: "fa876c25517a6b701529dfe671c1a8cc7b92d8e0",
		HashMin: "bb4321084eda30b487a3350aa40411f936121e5f",
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
		HashMax: "80114807307c37fc32705f45d82db0a191eea748",
		HashMin: "8674bf2a3c87e7d48c925f632fa833065714ab7e",
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
		HashMax: "bc89a9d05ca06ce47a545806de57c7e75e90c3c6",
		HashMin: "9f9a8410ef1a10d3922ca3d090ca74d6024a01b4",
	},
	"crypto/sha1": {
		HashMax: "bd55613d0981cc07c07248daf2e997743e724268",
		HashMin: "b1d92e4ac379ed875e57ae9a77976f719fbd000b",
	},
	"crypto/sha256": {
		HashMax: "96187694ba4d9ea5cecdbba60708cc73253b404b",
		HashMin: "ba8c205388d70bbbb3860cbaf879a5273baf8fe3",
	},
	"crypto/sha512": {
		HashMax: "ab38787e36b963627e2fe194fac8bf180a4f6e0c",
		HashMin: "6aebc88cd3624d121d87c2896ae08bb669c4cdc6",
	},
	"crypto/subtle": {
		HashMax: "ed49b3db5cf3f0fa9a2b2a74c5d86478c288b4db",
		HashMin: "20dacf9e39e8e538911ed08aaef990e789afc709",
	},
	"crypto/tls": {
		HashMax: "f0d769149e338d78a56a6a46f21ee19e326af983",
		HashMin: "4923e94da133b1cfdda077af58ab3ab2613d31ef",
	},
	"crypto/x509": {
		HashMax: "a765e3ee6bd78d8a6781b57c158ebe2d7f1856e0",
		HashMin: "947b608e982bc773ad355127657bfa26fac22f06",
	},
	"crypto/x509/pkix": {
		HashMax: "1ecce3059f721505380075bfc411902138cc6fd6",
		HashMin: "1d51f9b914e86629a9e9f71d88185c21e1d5420c",
	},
	"database/sql": {
		HashMax: "7017312fc63d0d47a38f7d9b5efd1f52251be947",
		HashMin: "053f693b4394c1d61be4f18926fd69b9fc349d6e",
	},
	"database/sql/driver": {
		HashMax: "74024871b753aea4cfcd56f8f69d8193da5c68b0",
		HashMin: "a014b4d6a1d4559a1dbc8951364d115cda81f13b",
	},
	"debug/dwarf": {
		HashMax: "37b2e83139055dc609fa7df40aa0b1efdb70ca97",
		HashMin: "865c54e749b676dfd1b42e3d883b612825b7b326",
	},
	"debug/elf": {
		HashMax: "6069b79b8ef7476c16dc0c6317bd30f60bffc1bb",
		HashMin: "f08d03599c5fbf649aa322c6104d26041023e453",
	},
	"debug/gosym": {
		HashMax: "b12a5ebbec1c8803978460685fbececa175f1b2a",
		HashMin: "a52c0f6ab730d3c4f281f11d469cb862d7ade39a",
	},
	"debug/macho": {
		HashMax: "03ff5ad066a88884e569ed0746db3042b8229681",
		HashMin: "970c2455da2babb84908ad5678bd1b2c24955684",
	},
	"debug/pe": {
		HashMax: "4f7d3d6e1f30547ac5c603997f445b98ecf1e9bc",
		HashMin: "fbc700c3d3acbd947395e1b20097e2e2b09ca16d",
	},
	"debug/plan9obj": {
		HashMax: "3f8aa95c6ccb860761f1c590734c22af014762ef",
		HashMin: "e00f83283090ba67da165f66ea0d078c98ef0eb7",
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
		HashMax: "3092e92c4bb384dc34569684c6e657389d40cd24",
		HashMin: "dfa7400f12f8152d7764e38d5ea2242e36686810",
	},
	"encoding/base32": {
		HashMax: "7c51d22cc2ccbd6a06d9c898e79ae862be15506c",
		HashMin: "34dea370bc28e7f35acd293fc2e4b20888bdfdce",
	},
	"encoding/base64": {
		HashMax: "ed66ee4582f27b07c6a6d8df0ac93ee63b4f18ea",
		HashMin: "75ef58eae242e4deacbdfa7948cc68741829e323",
	},
	"encoding/binary": {
		HashMax: "0d837f81c0699751ccbaa2bceb001bffa6327f05",
		HashMin: "6847fdc47acbe42028f1aad6e3e0e59b9eb623d6",
	},
	"encoding/csv": {
		HashMax: "f5d079bdf38f477a13a487972b654cd77628345d",
		HashMin: "4428696cdaa8252a5e91ee80aa5feb4cc9f65546",
	},
	"encoding/gob": {
		HashMax: "016e97c939a6c93a28704cdc1516b190b3ef23b6",
		HashMin: "70038663c154bf894527fc333cebf25ff0c91945",
	},
	"encoding/hex": {
		HashMax: "0dd3d0cce86a2dd7d4010aa9f8e17e93602bbf9a",
		HashMin: "b52381799eef11031acb452384951109033d7f27",
	},
	"encoding/json": {
		HashMax: "b4c4fd931f704247ea9a7829ee169a47ac633073",
		HashMin: "423532cf62677962f729bb09d4ac32367380b12b",
	},
	"encoding/pem": {
		HashMax: "46e7313ecb92cad5ccd66dcffda32b25deea3008",
		HashMin: "31a7e505dccaf9fda35b1da40b61a8075c359328",
	},
	"encoding/xml": {
		HashMax: "890596cb4c76b128098493997668ec46c2e4697a",
		HashMin: "36d74a577e7c001706bcb0acaab9ff66cebd17e0",
	},
	"errors": {
		HashMax: "886e1794220ce12c736a2353097f1ef03d402fc5",
		HashMin: "fcf1373327065f532a6495922919487ba0f6b76a",
	},
	"expvar": {
		HashMax: "cf777ebf7fcd1b3014f50767fdaf09d2b4d05227",
		HashMin: "5b19778a07ab569f4cae4c0ab2f3d0b9376342f3",
	},
	"flag": {
		HashMax: "ca8babce67003633764159cc186943b69f567216",
		HashMin: "cb0a5d8e3640f987f9ea53d2b5ceb493bf080ac3",
	},
	"fmt": {
		HashMax: "1174e78667f3ea29e9368b60cec28d98210f25fd",
		HashMin: "f33f5232c6705001c6d88be48cb828f5247868ad",
	},
	"github.com/google/pprof/driver": {
		HashMax: "15fe67485d256ea8aad46ca0197e71766750cbbd",
		HashMin: "3b493a3eacdc6d6d8bf0881deb926bc70a04b872",
	},
	"github.com/google/pprof/internal/binutils": {
		HashMax: "393ecdbe7036d028c29b5858139af87a09ef2105",
		HashMin: "8b2b51e6b5177a26b97652e873be11cf9c90af16",
	},
	"github.com/google/pprof/internal/driver": {
		HashMax: "1ce2b29f7be3219690717ab07e89887db6041861",
		HashMin: "65835ab639d4aa1082191039d9e88ab3ae44756f",
	},
	"github.com/google/pprof/internal/elfexec": {
		HashMax: "3c239b3fe0297cdcd737e8785822654fc1d2e4dd",
		HashMin: "a38d3ea2aaab1025f6a55ef615169c8c3b5c8ed5",
	},
	"github.com/google/pprof/internal/graph": {
		HashMax: "375b31ad303824e27766866329f4894adfaa52bd",
		HashMin: "d0d1b5404e3ef9a091c423a42a538a1e1ac86efb",
	},
	"github.com/google/pprof/internal/measurement": {
		HashMax: "72b570e242e09d1787a78a5c299c8adbce48a8a1",
		HashMin: "117106a1c7581a8930234a0a617b44ac4d2f219a",
	},
	"github.com/google/pprof/internal/plugin": {
		HashMax: "2c1ed90fe2e4a54b20efe4a668e6a02281d7c8d0",
		HashMin: "f03a10372949e10d80dfb58f3ea10f1692cc2553",
	},
	"github.com/google/pprof/internal/report": {
		HashMax: "08edbee975a68cb5489b1d6fad0a4a7fd45fd7c8",
		HashMin: "e6327e2a9d39b06d5f9c8163ee7429f2c4930a07",
	},
	"github.com/google/pprof/internal/symbolizer": {
		HashMax: "17ca1af50f76907bc155c4b077ea86327038d6fe",
		HashMin: "0c5f871c7046ff1b344f76206d4f9f448dc7c8ab",
	},
	"github.com/google/pprof/internal/symbolz": {
		HashMax: "835fd1d4243553fcbb2febf21fc0edc527250214",
		HashMin: "076dbf2581da7950fbeefb874f19273b60bb5666",
	},
	"github.com/google/pprof/profile": {
		HashMax: "228dbdfbdc5779447d33d7be5f9050f27139986d",
		HashMin: "4255600bf8798a0df0d536a472103c63bf3e792d",
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
		HashMax: "5162848a89ba420cc20a4747e789df1c38ed44fa",
		HashMin: "344d082fb13cc87dd506fbda84d27dd06ed7dc4e",
	},
	"go/ast": {
		HashMax: "be71f297707061251d8e0e7e3467a67be7c7419b",
		HashMin: "59e36c3473d163e90a2627c90e01c0cc6eac1075",
	},
	"go/build": {
		HashMax: "968bdfebaa18159e372a3b58d17bbe51ddfb6531",
		HashMin: "e79567f9631fef740c204a64e58840307637a707",
	},
	"go/constant": {
		HashMax: "ab196fe5e8c3e6ab7c4cd9ee1b30717ba9d58420",
		HashMin: "5d2672daf134b5eb9e2181b31e2d0069a93a83f2",
	},
	"go/doc": {
		HashMax: "03cec3448710082178b067abe49b901f9d3bead2",
		HashMin: "a3b704e72d68e7822b8adfaa348052671f19ac3c",
	},
	"go/format": {
		HashMax: "a7f398008d84d3e9711c908c4d1ccdc585b7ca79",
		HashMin: "403f0d1a96527311c63b714bb5e06bb66c5d848c",
	},
	"go/importer": {
		HashMax: "121ab94048d14d37ef5cc2155dcedb7fee3e2eda",
		HashMin: "395cd4b4f347c34cb82ae4f311aa839c0a2eb20c",
	},
	"go/internal/gccgoimporter": {
		HashMax: "23aa239bec3cc0b1874c8a4baa6c22241e5f20ed",
		HashMin: "039c8339e61ffd31534ddecd0e1bc89c3cb60285",
	},
	"go/internal/gcimporter": {
		HashMax: "cfb588b51e65327c60f14f554b357681a099ae67",
		HashMin: "8861bcdabf5dcb9b9572cf7722361286a5a04b28",
	},
	"go/internal/srcimporter": {
		HashMax: "1380a020bc9305675e5481dcee4e1633ba35adee",
		HashMin: "86476be9e5820be56c7f505bdbbace727f8b9f70",
	},
	"go/parser": {
		HashMax: "18cea8427d2ee72655bc09a04c7b68c5b44acf55",
		HashMin: "d496771da7964a1b65bd4d3cf5df6e1194d8334d",
	},
	"go/printer": {
		HashMax: "8d7fe791f91451e2ed29851cb81a1cbd9580e1f8",
		HashMin: "035009542cb4c96a2de753b0177de03c318f0cae",
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
		HashMax: "c5b6374b464612a1450f81caf8733b4fa8033410",
		HashMin: "9587b3ec6b5804c4e8ae27e62e4c8825d22cde71",
	},
	"golang.org/x/arch/arm/armasm": {
		HashMax: "02654bcb84200d91c3de5cb02331b352da2036f9",
		HashMin: "9dfc5f5f2fddc25a8f5f4ac4397ff330ec765520",
	},
	"golang.org/x/arch/ppc64/ppc64asm": {
		HashMax: "133e823237e62498de49970f5a139cb75598d22b",
		HashMin: "69db7411831e35c2324d6410e9a5cff14f482ebb",
	},
	"golang.org/x/arch/x86/x86asm": {
		HashMax: "d9bec60c60c3904e34c4df4cfd3f30b46651294c",
		HashMin: "aa252c7e82c012176c19d6b10aaa1195f0a0bf55",
	},
	"golang_org/x/crypto/chacha20poly1305": {
		HashMax: "158a3be7ad003b3fbac39f8991ddbac41d526255",
		HashMin: "bbd59c273b1e35c522b7c702d82643de01757dc1",
	},
	"golang_org/x/crypto/chacha20poly1305/internal/chacha20": {
		HashMax: "e260b9ad4eb715cb385af0afd7a13fd599a2a23d",
		HashMin: "2254ddc8e0334badd2ccf67f52f01354be288e84",
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
		HashMax: "ee9464f7a84f78cdde55f230db8ef865de6da861",
		HashMin: "438054f8c7c80cdd05dd1c1e52c76399dda05742",
	},
	"golang_org/x/net/idna": {
		HashMax: "e7fdedf7805f8fc63c662420ee859bb74bcebe81",
		HashMin: "b2fffa3968e9cc981447507d7609470fcf8a5dbb",
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
		HashMax: "2b1d04d821976a66a463414ea6ae4af097fe8104",
		HashMin: "9d65a1e2d3f786fa2c46339055abbebb78b7d706",
	},
	"golang_org/x/text/secure/bidirule": {
		HashMax: "a11094eb47c5d1f1a0154c632133cdcd2a4d353b",
		HashMin: "e581ac695f673b0a4d36f55dc35699e9d358159b",
	},
	"golang_org/x/text/transform": {
		HashMax: "ca7ffeb6f50548142374341e241870a2df2e89d5",
		HashMin: "d39ed0bbdccf845450b73dd5825814c6bca62787",
	},
	"golang_org/x/text/unicode/bidi": {
		HashMax: "aa75e64103f5b1ed94e4bc13c602f63b7b97a57a",
		HashMin: "33d058ddb4ed1d48064ad93e491b7db6df9ead5f",
	},
	"golang_org/x/text/unicode/norm": {
		HashMax: "fa53c7dec5b899ced07f16a19af2a0ba01aa0ac6",
		HashMin: "2149535a7fed54fa57eb538d150eb89a21020cdd",
	},
	"hash": {
		HashMax: "e329c0e33783da782d2383ec4d0d16d30f67892b",
		HashMin: "089192c7e86e9e29a8b8250f1d82d364695f2929",
	},
	"hash/adler32": {
		HashMax: "328f7451afcd4893f7b1db9c604c31909be93bc6",
		HashMin: "19cf57c83ac115f21f09370f24495c7daeff5a96",
	},
	"hash/crc32": {
		HashMax: "8a30d743ed99b27fda312e8189406622604749ae",
		HashMin: "9d04e701eb071d828c0943cd2f7869cb4d6898c1",
	},
	"hash/crc64": {
		HashMax: "9af7f2e3069630a5ac6ba06a7c18e0975d5796a4",
		HashMin: "ebd550ff7a03c9e77f0bc6a4c66e6d4149a6c502",
	},
	"hash/fnv": {
		HashMax: "63b8c817673eccb45546ab4b6d2320cf41c3a85e",
		HashMin: "f78c8f71dd5fbc329fd566503bf3b339fc6aa184",
	},
	"html": {
		HashMax: "719f600b345be5c2c4be25b349cc1d847a431fc8",
		HashMin: "fa0c758d50e488338e7506fdb5f706d2aa4055ff",
	},
	"html/template": {
		HashMax: "914493b1a2f5e9bf4289acffe25faecbdeac7fed",
		HashMin: "8eccf7490411cc9403310bcca6b73392a7901d74",
	},
	"image": {
		HashMax: "34a465ef7e86aae83ba17c3a5c377bffe7bc386f",
		HashMin: "9eb25575a4477d9c727af69ce6abab046356f9f1",
	},
	"image/color": {
		HashMax: "09529ff5efbf5e29cfdb46022d770998c3cebcbd",
		HashMin: "170741da655663f518ce1b8f94e0deaf62865da9",
	},
	"image/color/palette": {
		HashMax: "0de8f292d07e6094d805ec810bdef501778cf1fe",
		HashMin: "c636e93134517345f43e5fa90a2012c490e8afed",
	},
	"image/draw": {
		HashMax: "103ee7a9b377e6d7a4cde3f07c00850cfd2a2afc",
		HashMin: "59031b93c94aabcde7c3b21dc7c51fa4cd3b664f",
	},
	"image/gif": {
		HashMax: "17386b084c0dff97fc63635b70291bac778cb496",
		HashMin: "80c7d9a80a7cd0862591848f5761ab8b1a65e30c",
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
		HashMax: "c9c77f4bfe6a7b7d32dc186aa94b4fad7569f26c",
		HashMin: "9a6183b11c8838321bdbcc638c659b4be6dc0a91",
	},
	"index/suffixarray": {
		HashMax: "890ed1296740e98a0e2303c543676fa3ce90f926",
		HashMin: "7ed99d428c4d5feb70344af4ed03180f624fff89",
	},
	"internal/nettrace": {
		HashMax: "2fd8d800a101b64c845d0e151128823ac50ce685",
		HashMin: "e9c84746c6725f43758190d87b82b9740a228a9a",
	},
	"internal/poll": {
		HashMax: "41491b4b53111445aa5b91a3c9db563ccd6a59e2",
		HashMin: "8a4fcd512fdeaf17bab5ae033c8e170765c434d9",
	},
	"internal/race": {
		HashMax: "4bb49c054eb6aa21f9aa217c6b57e666c7681bd7",
		HashMin: "4b904d32201027c0e8aacbbf5be4cf9c0f597478",
	},
	"internal/singleflight": {
		HashMax: "c0f441c2c6168f89f11ef67a90cf33f74fb0e466",
		HashMin: "e455d1b06dbd1a304ac55d211a9809b162865a8b",
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
		HashMax: "dc6bd93ab55f52368470406e44f380f25f70c159",
		HashMin: "b7185cae2baff028cc2e54e9dac7b1bba820e399",
	},
	"internal/trace": {
		HashMax: "0b19f3e8a5a82de962952d7e7cbd3aab0eb84cd9",
		HashMin: "090af7ea04be92963117f7cbc3ee81544c8774e1",
	},
	"io": {
		HashMax: "8cf7fd5d8f383567f48853ab73c58d38e47f4f4d",
		HashMin: "6497ae654fa768e97244d0cf9f66ce9defda6ae3",
	},
	"io/ioutil": {
		HashMax: "954ec4af37ef76acb3ab86b2081b2d9376355e97",
		HashMin: "e93ace49c72f093d218ead1e9ad23411c98bdb7f",
	},
	"log": {
		HashMax: "d80734d26169f6668eaefedda718afbea69eadf2",
		HashMin: "4fc2b334b97a1d1b9d22c259640438a961810bbd",
	},
	"log/syslog": {
		HashMax: "4228a2f966d375bc0e034a4390c9549ae6a2b61e",
		HashMin: "cc63ad34c6c592528609cae63796b00df82493de",
	},
	"math": {
		HashMax: "839ea0c7ae167176c0a0cd39425b573a3d594e7f",
		HashMin: "0c9ea5e01080951eed635a3c2a2d9e4774f48c05",
	},
	"math/big": {
		HashMax: "05abf3c095619e22756cb495479c7df19057431c",
		HashMin: "ed05eec93683c016811c7b478018beea95d17c8f",
	},
	"math/bits": {
		HashMax: "bc287d1385da9be1725cb05c9431887d23639461",
		HashMin: "f601cad97a89015871e3ee769167c8dec5605432",
	},
	"math/cmplx": {
		HashMax: "4a1ae9aa573863b5c16890b4426d49eed330adc6",
		HashMin: "0353c92904e02e18a28d8b7853a6baeed1a67f60",
	},
	"math/rand": {
		HashMax: "a2fd01333ab1025ce6f758c2389f4215745984fa",
		HashMin: "8cfdab3a4041cf4101b779ba613ad2f254d33503",
	},
	"mime": {
		HashMax: "3d5327bc32621f67f180911fb3a8779040d7f520",
		HashMin: "7565f5260f800f2e08f484a003108f1b2c746378",
	},
	"mime/multipart": {
		HashMax: "bddca2aec6ae1070ee267aaae437146396654621",
		HashMin: "4c2fb3b46748f07e03c6657d150cc6de7ea99753",
	},
	"mime/quotedprintable": {
		HashMax: "caf88113afe6f49ff622ed21e8ff275cb2d9e6ba",
		HashMin: "7b236f837ecd01f6c62e6cd70e73ddb3bb3f84f6",
	},
	"net": {
		HashMax: "e2d9d853007b50209a7fa7b37fd8ca271c46ebf8",
		HashMin: "4363a3010920806c8afb91506737847968cfd6fa",
	},
	"net/http": {
		HashMax: "e3ab36259964dce6e92b92382cff975fd65ebf46",
		HashMin: "b7a80ed03d6912fe4143585c28a42edc22bc2f76",
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
		HashMax: "39c0bf6d0ac27a65ddb16ef0d225b780a4acdcb4",
		HashMin: "db9a9419d8856c6ac8c7fbc06e37f2c225d6b3b6",
	},
	"net/http/httptest": {
		HashMax: "ca7c74feb4b3fb6cca46e05b1eaa72c65df4604f",
		HashMin: "ae1234c28d700d875aa8267f700b171ae6d90a9a",
	},
	"net/http/httptrace": {
		HashMax: "81bf3216a56b0e510af7f376885624d2b9998b39",
		HashMin: "08ea793e2d06ccf3283ceadccec9796fe478834f",
	},
	"net/http/httputil": {
		HashMax: "54d8cbe51d7fa18648e5df97425547dca952713a",
		HashMin: "bc7fc02002a7b3851077cdbed0fb11375d3a2035",
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
		HashMax: "ff02dfbda1c2f3114f4d9376b2f33131134b3a29",
		HashMin: "19d0b6c25aab6556b8a4fa4ba5b8430c240206b6",
	},
	"net/rpc": {
		HashMax: "7f7f23c4e4dbd8d7de0494d6fa368c2eeae6d296",
		HashMin: "f9b18e1f19da84befb199033b05f0f1a423df419",
	},
	"net/rpc/jsonrpc": {
		HashMax: "7a9f2dd0f2d6c4d2ae6fde3a96352a66fef8ffa3",
		HashMin: "fca77e7cf35126b3e60cf2e9e6c98dbbe038ba7d",
	},
	"net/smtp": {
		HashMax: "b86f33ba909111e541e7edff5175dfe35c81a602",
		HashMin: "3758db47d03a599ee8ca7ea2a7fa86812bd69cba",
	},
	"net/textproto": {
		HashMax: "30e8ad26a35f759f118877056df36f433e80a23d",
		HashMin: "8af25838b513e83b883e9a4121ccaff5c983f33a",
	},
	"net/url": {
		HashMax: "fc11049ca55bfe08272d8c3441823469a5278f95",
		HashMin: "a42759d0dec88030217d5d39c3fc77432eb59b42",
	},
	"os": {
		HashMax: "9479af62c63d1d98b8f886a8e63fbd9704655257",
		HashMin: "b7d73162cb7198e6def755a065401cdd60d056b2",
	},
	"os/exec": {
		HashMax: "131a1aa45b97d0c35c6bc258e30ea402cbad39df",
		HashMin: "f6bf2cf319feb4f6e4617999b151ad028b0f4916",
	},
	"os/signal": {
		HashMax: "bb3c78be12271c70fe1c873857b8aac5d85b3d07",
		HashMin: "4a10f569ddc81a9520e94f12c4924246181f7785",
	},
	"os/user": {
		HashMax: "6aa259f6f37549366bc538d20de9f6103af10b65",
		HashMin: "9fbc106bf1e3f59ad5cd2712a1b38ddaa606b370",
	},
	"path": {
		HashMax: "07831caed0377467fea67da31731303a38960bc9",
		HashMin: "549632e142893a702850e27669501a110c13293f",
	},
	"path/filepath": {
		HashMax: "b7865a17abe4be920689f9abecc248bf882d218d",
		HashMin: "c34f5cd4d2e5e1b658a441e6b00fa5c9d10a7830",
	},
	"reflect": {
		HashMax: "420dcef512badada25816cc4f7087c8cb9961e6d",
		HashMin: "23fc1c4bc9a5581b2bdd3afa4365e2a066c39c2e",
	},
	"regexp": {
		HashMax: "361f6d8f0521c0a008b4d5a703c35473c129cd67",
		HashMin: "d45a342c3ce48406127779e68593e8ae4506cc99",
	},
	"regexp/syntax": {
		HashMax: "3c5ff9ea5ce19505f6953da7f6e260561f06287c",
		HashMin: "e1f4a439f087a93f7dc944153554fb3dfaf25ab7",
	},
	"runtime": {
		HashMax: "0248fe4510192108b45bd22e971b76e63a89b690",
		HashMin: "7cd5351b71dc57a4f5cb23afe4e60850b0c73ef8",
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
		HashMax: "a322ea41bd11f899f81865886dbb1dbebd4a78da",
		HashMin: "a1540c7e6eed490905543931b7f4e3355a04c113",
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
		HashMax: "b13ed7cb35eda6d2d062a4c3b181e621663e33d6",
		HashMin: "395fe046062eb0d41c1c90eeaf95db3a50579a53",
	},
	"strconv": {
		HashMax: "5763a223c6208dedd367f9b185136630ae5357af",
		HashMin: "636bccc08791c861c77f3490804d7440ceb6bea5",
	},
	"strings": {
		HashMax: "6793f328fc233911c6e465d84d58e066acc7466a",
		HashMin: "af83ab1eb65383d9ffa4adf84fd1018fbf9c582f",
	},
	"sync": {
		HashMax: "bb98fdf8d5e59e04bcd62c548ccf89a5c0cc1d15",
		HashMin: "ef9c33df5e752e90209d822f04f299ceb19f0bf5",
	},
	"sync/atomic": {
		HashMax: "eb4603bf01c9a808563562e2c66acbff685df271",
		HashMin: "05111322aaeadbc937d992c8a44c83dc3d01aa3e",
	},
	"syscall": {
		HashMax: "6d3d4b69d810d40861a10c2f92886b1e408c75eb",
		HashMin: "ec028e4bce355d5db9e2f22f7b21a8c6fc2efa2b",
	},
	"testing": {
		HashMax: "b15365bf32abde9a42625e66bf8717253702a544",
		HashMin: "301aff3d56d37e8bd38ca6dd1264627ee11352cf",
	},
	"testing/internal/testdeps": {
		HashMax: "d52ece2e32f643c1f790ca86bf9d2e7f18f9f03a",
		HashMin: "fc1e4e921b9020445d6e93e560fa463461c315de",
	},
	"testing/iotest": {
		HashMax: "83ea335c9825c934ff86a644d3073f01f8c9647b",
		HashMin: "e7c860e87d61c9aede1fa65b384a26246dea6af5",
	},
	"testing/quick": {
		HashMax: "65556aca96ee272dd3d4009507e4ad277f2535e7",
		HashMin: "51a7d39cde9bc6614332b2550ca3e7bc6a0ed469",
	},
	"text/scanner": {
		HashMax: "76777659db086b1e397aa717e36634f5e9d54ece",
		HashMin: "99f3d62d5ec5dd50d8a8e52bcdc03c0671e6532e",
	},
	"text/tabwriter": {
		HashMax: "589d99babe408e0759ac063a05813d996b42eddb",
		HashMin: "9e5d9097ad06b08b852f931c84a7980a4b4e5d77",
	},
	"text/template": {
		HashMax: "8af52c3c6c407b39d0856df8424020186bb12b29",
		HashMin: "727479f8e1da243533396eba4a0ea93bed6d46b9",
	},
	"text/template/parse": {
		HashMax: "246fd1dcde3905c39ebbbab34b46f35346cf40d9",
		HashMin: "05a66b0f4e440f6eb8fe7a1501638a8de864e347",
	},
	"time": {
		HashMax: "a10ca8e16a977ec3f363fc5362329699881c3289",
		HashMin: "e385c21d018cbdb36ac4faf18471ec690b905e65",
	},
	"unicode": {
		HashMax: "2afa7bde16a2e0bbc84eb46ee5ecfc2adeccebbc",
		HashMin: "2e52c54bcc8e9c3cf29a5a20029f00c1de1ef498",
	},
	"unicode/utf16": {
		HashMax: "cad2c09cbfef4b755810abc40b6c58bb4eaaa7af",
		HashMin: "40cde41561d9378c3c58c8b1831a188250a2b778",
	},
	"unicode/utf8": {
		HashMax: "4315f7f70f323bf23d63192036c6117ee86c47cc",
		HashMin: "74a321a1effd55c9ff40bd9af5869bc88e79d6e3",
	},
	"unsafe": {
		HashMax: "5a535455f5561432f06987f3ed4c103353691c20",
		HashMin: "f6cbd69d1a94141ebde424922312b4f98cdc46dd",
	},
}
