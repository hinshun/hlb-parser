package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/repr"
	x "github.com/hinshun/hlb-parser/hlb"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	mod := x.Module(
		x.Import("go").From(
			x.Ident("image").Call("openllb/go.hlb"),
		),

		x.Func("node").Returns("fs").Body(
			x.Ident("image").Call("node:alpine"),
		),

		x.Func("run").Params(x.Variadic("string"), "args").Returns("fs"),

		x.Func("nodeModules").Export().Returns("fs").Body(
			x.Comments("Optional parens for no argument functions"),
			x.Ident("node"),
			x.Ident("run").Call("npm install").With(x.Array("[]option::run").Items(
				x.Ident("dir").Call("/in"),
				x.Ident("mount").Call(x.Ident("src"), "/in"),
				x.Ident("mount").Call(x.Ident("scratch"), "/in/node_modules").As("return"),
			)),
		),

		x.Func("publishDigest").Returns("string").Body(
			x.Ident("nodeModules"),
			x.Ident("dockerPush").At("digest"),
		),

		x.Func("props").Returns("fs").Body(
			x.Ident("mkfile").Call("node_modules.props", "0o644", `<<~EOF
		digest=${publishDigest}
	EOF`),
		),
	)
	repr.Println(mod)
	return nil
}
