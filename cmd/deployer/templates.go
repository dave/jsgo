package deployer

import "text/template"

var defaultIndexTemplate = template.Must(template.New("main").Parse(`<html>
<head><meta charset="utf-8"></head>
<body>
	<script src="{{ .Script }}"></script>
	<script src="{{ .Loader }}"></script>
</body>
</html>`))

var loaderTemplate = template.Must(template.New("main").Parse(`if (!WebAssembly.instantiateStreaming) {
	WebAssembly.instantiateStreaming = async (resp, importObject) => {
		const source = await (await resp).arrayBuffer();
		return await WebAssembly.instantiate(source, importObject);
	};
}
const go = new Go();
WebAssembly.instantiateStreaming(fetch("{{ .Binary }}"), go.importObject).then(result => {
	go.run(result.instance);
});`))

var loaderTemplateMin = template.Must(template.New("main").Parse(`WebAssembly.instantiateStreaming||(WebAssembly.instantiateStreaming=(async(t,a)=>{const e=await(await t).arrayBuffer();return await WebAssembly.instantiate(e,a)}));const go=new Go;WebAssembly.instantiateStreaming(fetch("{{ .Binary }}"),go.importObject).then(t=>{go.run(t.instance)});`))
