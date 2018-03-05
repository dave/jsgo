package old

func applyStyles() {
	css := `
		html, body {
			height: 100%;
		}
		.editor {
			height: calc(100% - 65px);
			width: 100%;
		}
		.split {
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
