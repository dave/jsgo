package deployer

const index = `<!doctype html>
<html>
<head><meta charset="utf-8"></head>
<body>
	<script src="{{ .Script }}"></script>
	<script>
		if (!WebAssembly.instantiateStreaming) {
			WebAssembly.instantiateStreaming = async (resp, importObject) => {
				const source = await (await resp).arrayBuffer();
				return await WebAssembly.instantiate(source, importObject);
			};
		}
		const go = new Go();
		WebAssembly.instantiateStreaming(fetch("{{ .Binary }}"), go.importObject).then(result => {
			go.run(result.instance);
		});
	</script>
</body>
</html>`
