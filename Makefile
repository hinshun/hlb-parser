wasm:
	@GOOS=js GOARCH=wasm go build -o public/parser.wasm
