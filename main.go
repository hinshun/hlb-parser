package main

import (
	"fmt"
	"os"
	"strconv"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"
)

var (
	Lexer = lexer.MustStateful(lexer.Rules{
		"Root": {
			{"Keyword", `\b(if|else|for|in|with|as|import|export)\b`, nil},
			{"Numeric", `\b(0(b|B|o|O|x|X)[a-fA-F0-9]+)\b`, nil},
			{"Decimal", `\b(0|[1-9][0-9]*)\b`, nil},
			{"Bool", `\b(true|false)\b`, nil},
			{"String", `"`, lexer.Push("String")},
			{"RawString", "`", lexer.Push("RawString")},
			{"Heredoc", `<<[-~]?(\w+)`, lexer.Push("Heredoc")},
			{"RawHeredoc", "<<[-~]?`(\\w+)`", lexer.Push("RawHeredoc")},
			{"Block", `{`, lexer.Push("Block")},
			{"Paren", `\(`, lexer.Push("Paren")},
			{"Ident", `[\w:]+`, lexer.Push("Reference")},
			{"Condition", `>=|<=|&&|\|\||==|!=`, nil},
			{"Operator", `[-+*/<>%^!|&]`, nil},
			{"Punct", `[][@:;.,?]`, nil},
			{"Newline", `\n`, nil},
			{"Comment", `#[^\n]*\n`, nil},
			{"whitespace", `[\r\t ]+`, nil},
		},
		"Reference": {
			{"Ident", `[\w:]+`, nil},
			lexer.Return(),
		},
		"String": {
			{"StringEnd", `"`, lexer.Pop()},
			{"Escaped", `\\.`, nil},
			{"Interpolated", `\${`, lexer.Push("Interpolated")},
			{"Char", `\$|[^"$\\]+`, nil},
		},
		"RawString": {
			{"RawStringEnd", "`", lexer.Pop()},
			{"RawChar", "[^`]+", nil},
		},
		"Heredoc": {
			{"HeredocEnd", `\b\1\b`, lexer.Pop()},
			{"Spaces", `\s+`, nil},
			{"Escaped", `\\.`, nil},
			{"Interpolated", `\${`, lexer.Push("Interpolated")},
			{"Text", `\$|[^\s$]+`, nil},
		},
		"RawHeredoc": {
			{"RawHeredocEnd", `\b\1\b`, lexer.Pop()},
			{"Whitespace", `\s+`, nil},
			{"RawText", `[^\s]+`, nil},
		},
		"Interpolated": {
			{"BlockEnd", `}`, lexer.Pop()},
			lexer.Include("Root"),
		},
		"Block": {
			{"BlockEnd", `}`, lexer.Pop()},
			lexer.Include("Root"),
		},
		"Paren": {
			{"ParenEnd", `\)`, lexer.Pop()},
			lexer.Include("Root"),
		},
	})

	Parser = participle.MustBuild(
		&Module{},
		participle.Lexer(Lexer),
	)
)

type Module struct {
	Comments *Comments `parser:"@@?"`

	Decls []*Decl `parser:"@@*"`
}

type Comments struct {
	Comments []*Comment `parser:"@@+"`
}

type Comment struct {
	Text string `parser:"@Comment"`
}

type Decl struct {
	Import   *ImportDecl `parser:"( @@"`
	Export   *ExportDecl `parser:"| @@"`
	Func     *FuncDecl   `parser:"| @@"`
	Newline  *Newline    `parser:"| @@"`
	Comments *Comments   `parser:"| @@ )"`
}

type ImportDecl struct {
	Import *Import `parser:"@@"`
	Name   *Ident  `parser:"@@"`
	From   *From   `parser:"@@"`
	Expr   *Expr   `parser:"@@"`
}

type Import struct {
	Text string `parser:"@'import'"`
}

type From struct {
	Text string `parser:"@'from'"`
}

type ExportDecl struct {
	Export *Export `parser:"@@"`
	Name   *Ident  `parser:"@@"`
}

type Export struct {
	Text string `parser:"@'export'"`
}

type FuncDecl struct {
	Func    *Func      `parser:"@@"`
	Name    *Ident     `parser:"@@"`
	Params  *FieldList `parser:"@@"`
	Type    *Type      `parser:"@@"`
	Effects *FieldList `parser:"@@?"`
	Body    *StmtList  `parser:"@@"`
}

type Func struct {
	Text string `parser:"@'func'"`
}

type FieldList struct {
	OpenParen  *OpenParen   `parser:"@@"`
	Fields     []*FieldStmt `parser:"@@*"`
	CloseParen *CloseParen  `parser:"@@"`
}

type FieldStmt struct {
	Field    *Field    `parser:"( @@ ','?"`
	Newline  *Newline  `parser:"| @@"`
	Comments *Comments `parser:"| @@ )"`
}

type Field struct {
	Type     *Type   `parser:"@@"`
	Variadic *string `parser:"@( '.' '.' '.' )?"`
	Name     *Ident  `parser:"@@"`
}

type StmtList struct {
	OpenBrace  *OpenBrace  `parser:"@@"`
	Stmts      []*Stmt     `parser:"@@*"`
	CloseBrace *CloseBrace `parser:"@@"`
}

type OpenBrace struct {
	Text string `parser:"@Block"`
}

type CloseBrace struct {
	Text string `parser:"@BlockEnd"`
}

type OpenParen struct {
	Text string `parser:"@Paren"`
}

type CloseParen struct {
	Text string `parser:"@ParenEnd"`
}

type Stmt struct {
	If       *IfStmt   `parser:"( @@"`
	For      *ForStmt  `parser:"| @@"`
	Expr     *Expr     `parser:"| @@"`
	Newline  *Newline  `parser:"| @@"`
	Comments *Comments `parser:"| @@ )"`
}

type IfStmt struct {
	If        *If           `parser:"@@"`
	Condition *Expr         `parser:"@@"`
	Body      *StmtList     `parser:"@@"`
	ElseIfs   []*ElseIfStmt `parser:"@@*"`
	Else      *ElseStmt     `parser:"@@?"`
}

type ElseIfStmt struct {
	Else      *Else     `parser:"@@"`
	If        *If       `parser:"@@"`
	Condition *Expr     `parser:"@@"`
	Body      *StmtList `parser:"@@"`
}

type ElseStmt struct {
	Else *Else     `parser:"@@"`
	Body *StmtList `parser:"@@"`
}

type If struct {
	Text string `parser:"@'if'"`
}

type Else struct {
	Text string `parser:"@'else'"`
}

type ForStmt struct {
	For      *For      `parser:"@@"`
	Counter  *Ident    `parser:"( @@ ',' )?"`
	Var      *Ident    `parser:"@@"`
	In       *In       `parser:"@@"`
	Iterable *Expr     `parser:"@@"`
	Body     *StmtList `parser:"@@"`
}

type For struct {
	Text string `parser:"@'for'"`
}

type In struct {
	Text string `parser:"@'in'"`
}

type Expr struct {
	Block      *StmtList     `parser:"( @@"`
	Decimal    *int          `parser:"| @Decimal"`
	Numeric    *NumericLit   `parser:"| @Numeric"`
	Bool       *bool         `parser:"| @Bool"`
	String     *StringLit    `parser:"| @@"`
	RawString  *RawStringLit `parser:"| @@"`
	Heredoc    *Heredoc      `parser:"| @@"`
	RawHeredoc *RawHeredoc   `parser:"| @@"`
	Call       *CallExpr     `parser:"| @@ )"`
	Splat      *string       `parser:"@('.' '.' '.')?"`
}

type CallExpr struct {
	Func *IdentExpr  `parser:"@@"`
	Args *ExprList   `parser:"@@?"`
	At   *AtClause   `parser:"@@?"`
	With *WithClause `parser:"@@?"`
	As   *AsClause   `parser:"@@?"`
}

type IdentExpr struct {
	Name      *Ident     `parser:"@@"`
	Reference *Reference `parser:"@@?"`
}

type Reference struct {
	Dot   string `parser:"@'.'"`
	Field *Ident `parser:"@@"`
}

type ExprList struct {
	OpenParen  *OpenParen  `parser:"@@"`
	Exprs      []*ExprStmt `parser:"@@*"`
	CloseParen *CloseParen `parser:"@@"`
}

type ExprStmt struct {
	Expr     *Expr     `parser:"( @@ ','?"`
	Newline  *Newline  `parser:"| @@"`
	Comments *Comments `parser:"| @@ )"`
}

type AtClause struct {
	At     *At    `parser:"@@"`
	Effect *Ident `parser:"@@"`
}

type At struct {
	Text string `parser:"@'@'"`
}

type WithClause struct {
	With    *With `parser:"@@"`
	Expr    *Expr `parser:"@@"`
	Closure *FuncDecl
}

type With struct {
	Text string `parser:"@'with'"`
}

type AsClause struct {
	As     *As    `parser:"@@"`
	Effect *Ident `parser:"@@"`
}

type As struct {
	Text string `parser:"@'as'"`
}

type NumericLit struct {
	Value int64
	Base  int
}

func (l *NumericLit) Capture(tokens []string) error {
	base := 10
	n := tokens[0]
	if len(n) >= 2 {
		switch n[1] {
		case 'b', 'B':
			base = 2
		case 'o', 'O':
			base = 8
		case 'x', 'X':
			base = 16
		}
		n = n[2:]
	}
	var err error
	num, err := strconv.ParseInt(n, base, 64)
	l.Value = num
	l.Base = base
	return err
}

type StringLit struct {
	Start     *Quote            `parser:"@@"`
	Fragments []*StringFragment `parser:"@@*"`
	Terminate *Quote            `parser:"@@"`
}

type Quote struct {
	Text string `parser:"@(String | StringEnd)"`
}

type StringFragment struct {
	Escaped      *string       `parser:"( @Escaped"`
	Interpolated *Interpolated `parser:"| @@"`
	Text         *string       `parser:"| @Char )"`
}

type Interpolated struct {
	Start     string `parser:"@Interpolated"`
	Expr      *Expr  `parser:"@@?"`
	Terminate string `parser:"@BlockEnd"`
}

type RawStringLit struct {
	Start     *Backtick `parser:"@@"`
	Text      string    `parser:"@RawChar"`
	Terminate *Backtick `parser:"@@"`
}

type Backtick struct {
	Text string `parser:"@(RawString | RawStringEnd)"`
}

type Heredoc struct {
	Start     string             `parser:"@Heredoc"`
	Body      []*HeredocFragment `parser:"@@*"`
	Terminate *HeredocEnd        `parser:"@@"`
}

type HeredocFragment struct {
	Whitespace   *string       `parser:"( @Whitespace"`
	Interpolated *Interpolated `parser:"| @@"`
	Text         *string       `parser:"| @(Text | RawText) )"`
}

type HeredocEnd struct {
	EOF string `parser:"@(HeredocEnd | RawHeredocEnd)"`
}

type RawHeredoc struct {
	Start     string             `parser:"@RawHeredoc"`
	Body      []*HeredocFragment `parser:"@@*"`
	Terminate *HeredocEnd        `parser:"@@"`
}

type Type struct {
	Scalar *Ident `parser:"( @@"`
	Array  *Ident `parser:"| '[' ']' @@ )"`
}

type Ident struct {
	Text string `parser:"@Ident"`
}

type Newline struct {
	Text string `parser:"@Newline"`
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %s", err)
		os.Exit(1)
	}
}

func run() error {
	mod := &Module{}

	f, err := os.Open("./build.hlb")
	if err != nil {
		return err
	}
	defer f.Close()

	err = Parser.Parse(f.Name(), f, mod)
	fmt.Println(mod)
	repr.Println(mod)
	return err
}
