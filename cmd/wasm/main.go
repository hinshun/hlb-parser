package main

import (
	"bytes"
	"strings"
	"syscall/js"

	"github.com/alecthomas/repr"
	"github.com/hinshun/hlb-parser/ast"
)

func main() {
	js.Global().Set("parseHLB", parseWrapper())
	<-make(chan struct{})
}

func parseWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 1 {
			return "must have exactly 1 arg"
		}

		input := args[0].String()
		mod, err := parse(input)
		if err != nil {
			return err.Error()
		}
		return mod
	})
}

func parse(input string) (string, error) {
	mod := &ast.Module{}
	err := ast.Parser.Parse("build.hlb", strings.NewReader(input), mod)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	repr.New(buf).Println(mod)
	return buf.String(), nil
}
