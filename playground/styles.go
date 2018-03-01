package main

func applyStyles() {
	css := `
		html, body {
			height: 100%;
		}
		.split, .editor {
			height: 100%;
			width: 100%;
		}
		.gutter {
			height: 100%;
			background-color: #eee;
			background-repeat: no-repeat;
			background-position: 50%;
		}
		.gutter.gutter-horizontal {
			cursor: col-resize;
			background-image:  url('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAUAAAAeCAYAAADkftS9AAAAIklEQVQoU2M4c+bMfxAGAgYYmwGrIIiDjrELjpo5aiZeMwF+yNnOs5KSvgAAAABJRU5ErkJggg==')
		}
		.split {
			-webkit-box-sizing: border-box;
			-moz-box-sizing: border-box;
			box-sizing: border-box;
		}
		.split, .gutter.gutter-horizontal {
			float: left;
		}
		.split {
			overflow-y: auto;
			overflow-x: hidden;
		}
		.header {
			height: 20px;
			padding-left: 45px; /* margin: 41px, padding: 4px */
			padding-top: 4px;
			padding-bottom: 4px;
			background-color: #eee;
		}
		.preview {
			border: 0;
			height: 100%;
			width: 100%;
		}
	`
	s := document.CreateElement("style")
	s.SetInnerHTML(css)
	document.GetElementsByTagName("head")[0].AppendChild(s)
}
