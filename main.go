package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/repr"
	"github.com/hinshun/hlb-parser/ast"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %s", err)
		os.Exit(1)
	}
}

func run() error {
	f, err := os.Open("./baz.hlb")
	if err != nil {
		return err
	}
	defer f.Close()

	mod := &ast.Module{}
	err = ast.Parser.Parse(f.Name(), f, mod)
	repr.Println(mod)
	return err
}
