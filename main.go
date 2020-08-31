package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
	"github.com/alecthomas/participle/lexer/stateful"
	"github.com/alecthomas/repr"
)

var (
	Lexer = lexer.Must(stateful.New(stateful.Rules{
		"Root": {
			{"Keyword", `\b(import|export|with|as)\b`, nil},
			{"Numeric", `\b(0(b|B|o|O|x|X)[a-fA-F0-9]+)\b`, nil},
			{"Decimal", `\b(0|[1-9][0-9]*)\b`, nil},
			{"Bool", `\b(true|false)\b`, nil},
			{"String", `"`, stateful.Push("String")},
			{"RawString", "`", stateful.Push("RawString")},
			{"Heredoc", `<<[-~]?(\w+)`, stateful.Push("Heredoc")},
			{"RawHeredoc", "<<[-~]?`(\\w+)`", stateful.Push("RawHeredoc")},
			{"Block", `{`, stateful.Push("Block")},
			{"Paren", `\(`, stateful.Push("Paren")},
			stateful.Include("Common"),
		},
		"Common": {
			{"Ident", `[\w:]+`, stateful.Push("Reference")},
			{"Operator", `;`, nil},
			{"Newline", `\n`, nil},
			{"Comment", `#[^\n]*\n`, nil},
			{"whitespace", `[\r\t ]+`, nil},
		},
		"Reference": {
			{"Dot", `\.`, nil},
			{"Ident", `[\w:]+`, nil},
			{"pop", ``, stateful.PopIfEmpty()},
		},
		"String": {
			{"StringEnd", `"`, stateful.Pop()},
			{"Escaped", `\\.`, nil},
			{"Interpolated", `\${`, stateful.Push("Interpolated")},
			{"Char", `[^"$\\]+`, nil},
		},
		"RawString": {
			{"RawStringEnd", "`", stateful.Pop()},
			{"RawChar", "[^`]+", nil},
		},
		"Heredoc": {
			{"HeredocEnd", `\b\1\b`, stateful.Pop()},
			{"Whitespace", `\s+`, nil},
			{"Interpolated", `\${`, stateful.Push("Interpolated")},
			{"Text", `[^\s$]+`, nil},
		},
		"RawHeredoc": {
			{"RawHeredocEnd", `\b\1\b`, stateful.Pop()},
			{"Whitespace", `\s+`, nil},
			{"RawText", `[^\s]+`, nil},
		},
		"Interpolated": {
			{"BlockEnd", `}`, stateful.Pop()},
			stateful.Include("Root"),
		},
		"Block": {
			{"BlockEnd", `}`, stateful.Pop()},
			stateful.Include("Root"),
		},
		"Paren": {
			{"ParenEnd", `\)`, stateful.Pop()},
			{"Delimit", `,`, nil},
			stateful.Include("Root"),
		},
	}))

	Parser = participle.MustBuild(
		&Module{},
		participle.Lexer(Lexer),
	)
)

type Module struct {
	Pos   lexer.Position
	Decls []*Decl `parser:"@@*"`
}

type Decl struct {
	Pos              lexer.Position
	Import           *ImportDecl           `parser:"( @@"`
	Export           *ExportDecl           `parser:"| @@"`
	Func             *FuncDecl             `parser:"| @@"`
	Newline          *Newline              `parser:"| @@"`
	CommentGroup     *CommentGroup         `parser:"| @@ )"`
}

type ImportDecl struct {
	Pos    lexer.Position
	Import *Import `parser:"@@"`
	Expr   *Expr   `parser:"@@"`
	As     *As     `parser:"@@"`
	Name   *Ident  `parser:"@@"`
}

type Import struct {
	Pos  lexer.Position
	Text string `parser:"@'import'"`
}

type ExportDecl struct {
	Pos    lexer.Position
	Export *Export `parser:"@@"`
	Name   *Ident  `parser:"@@"`
}

type Export struct {
	Pos  lexer.Position
	Text string `parser:"@'export'"`
}

type FuncDecl struct {
	Pos     lexer.Position
	Type    *Ident         `parser:"@@"`
	Name    *Ident         `parser:"@@"`
	Params  *FieldList     `parser:"@@"`
	Effects *EffectsClause `parser:"( @@ )?"`
	Body    *BlockStmt     `parser:"@@"`
}

type EffectsClause struct {
	Pos     lexer.Position
	Binds   *Binds     `parser:"@@"`
	Effects *FieldList `parser:"@@"`
}

type Binds struct {
	Pos  lexer.Position
	Text string `parser:"@'binds'"`
}

type BlockStmt struct {
	Pos        lexer.Position
	OpenBrace  *OpenBrace  `parser:"@@"`
	Stmts      []*Stmt     `parser:"@@*"`
	CloseBrace *CloseBrace `parser:"@@"`
}

type OpenBrace struct {
	Pos  lexer.Position
	Text string `parser:"@Block"`
}

type CloseBrace struct {
	Pos  lexer.Position
	Text string `parser:"@BlockEnd"`
}

type FieldList struct {
	Pos        lexer.Position
	OpenParen  *OpenParen   `parser:"@@"`
	Fields     []*FieldStmt `parser:"@@*"`
	CloseParen *CloseParen  `parser:"@@"`
}

type FieldStmt struct {
	Pos     lexer.Position
	Field   *Field   `parser:"( @@ Delimit?"`
	Newline *Newline `parser:"| @@"`
	Comment *Comment `parser:"| @@ )"`
}

type Field struct {
	Pos      lexer.Position
	Modifier *Modifier `parser:"@@?"`
	Type     *Ident    `parser:"@@"`
	Name     *Ident    `parser:"@@"`
}

type Modifier struct {
	Pos      lexer.Position
	Variadic *Variadic `parser:"@@"`
}

type Variadic struct {
	Pos  lexer.Position
	Text string `parser:"@'variadic'"`
}

type OpenParen struct {
	Pos  lexer.Position
	Text string `parser:"@Paren"`
}

type CloseParen struct {
	Pos  lexer.Position
	Text string `parser:"@ParenEnd"`
}

type Stmt struct {
	Call    *CallStmt `parser:"( @@"`
	Expr    *ExprStmt `parser:"| @@"`
	Newline *Newline  `parser:"| @@"`
	Comment *Comment  `parser:"| @@ )"`
}

type CallStmt struct {
	Pos        lexer.Position
	Name       *IdentExpr  `parser:"@@"`
	Args       []*Expr     `parser:"@@*"`
	WithClause *WithClause `parser:"@@?"`
	Binds      *BindClause `parser:"@@?"`
	StmtEnd    *StmtEnd    `parser:"@@?"`
}

type WithClause struct {
	Pos     lexer.Position
	With    *With `parser:"@@"`
	Expr    *Expr `parser:"@@"`
	Closure *FuncDecl
}

type With struct {
	Pos  lexer.Position
	Text string `parser:"@'with'"`
}

type BindClause struct {
	Pos   lexer.Position
	As    *As        `parser:"@@"`
	Ident *Ident     `parser:"( @@"`
	Binds *FieldList `parser:"| @@ )"`
}

type As struct {
	Pos  lexer.Position
	Text string `parser:"@'as'"`
}

type ExprStmt struct {
	Pos     lexer.Position
	Expr    *Expr    `parser:"@@"`
	StmtEnd *StmtEnd `parser:"@@?"`
}

type StmtEnd struct {
	Pos       lexer.Position
	Semicolon *string  `parser:"( @';'"`
	Newline   *Newline `parser:"| @@"`
	Comment   *Comment `parser:"| @@ )"`
}

type Expr struct {
	Pos       lexer.Position
	FuncLit   *FuncLit   `parser:"( @@"`
	BasicLit  *BasicLit  `parser:"| @@"`
	CallExpr  *CallExpr  `parser:"| @@"`
	IdentExpr *IdentExpr `parser:"| @@ )"`
}

type FuncLit struct {
	Pos  lexer.Position
	Type *Ident     `parser:"@@"`
	Body *BlockStmt `parser:"@@"`
}

type BasicLit struct {
	Pos        lexer.Position
	Decimal    *int          `parser:"( @Decimal"`
	Numeric    *NumericLit   `parser:"| @Numeric"`
	Bool       *bool         `parser:"| @Bool"`
	String     *StringLit    `parser:"| @@"`
	RawString  *RawStringLit `parser:"| @@"`
	Heredoc    *Heredoc      `parser:"| @@"`
	RawHeredoc *RawHeredoc   `parser:"| @@ )"`
}

type NumericLit struct {
	Pos   lexer.Position
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
	Pos       lexer.Position
	Start     *Quote            `parser:"@@"`
	Fragments []*StringFragment `parser:"@@*"`
	Terminate *Quote            `parser:"@@"`
}

type Quote struct {
	Pos  lexer.Position
	Text string `parser:"@(String | StringEnd)"`
}

type StringFragment struct {
	Pos          lexer.Position
	Escaped      *string       `parser:"( @Escaped"`
	Interpolated *Interpolated `parser:"| @@"`
	Text         *string       `parser:"| @Char )"`
}

type Interpolated struct {
	Pos       lexer.Position
	Start     string `parser:"@Interpolated"`
	Expr      *Expr  `parser:"@@?"`
	Terminate string `parser:"@BlockEnd"`
}

type RawStringLit struct {
	Pos       lexer.Position
	Start     *Backtick `parser:"@@"`
	Text      string    `parser:"@RawChar"`
	Terminate *Backtick `parser:"@@"`
}

type Backtick struct {
	Pos  lexer.Position
	Text string `parser:"@(RawString | RawStringEnd)"`
}

type Heredoc struct {
	Pos       lexer.Position
	Start     string             `parser:"@Heredoc"`
	Body      []*HeredocFragment `parser:"@@*"`
	Terminate *HeredocEnd        `parser:"@@"`
}

type HeredocFragment struct {
	Pos          lexer.Position
	Whitespace   *string       `parser:"( @Whitespace"`
	Interpolated *Interpolated `parser:"| @@"`
	Text         *string       `parser:"| @(Text | RawText) )"`
}

type HeredocEnd struct {
	Pos lexer.Position
	EOF string `parser:"@(HeredocEnd | RawHeredocEnd)"`
}

type RawHeredoc struct {
	Pos       lexer.Position
	Start     string             `parser:"@RawHeredoc"`
	Body      []*HeredocFragment `parser:"@@*"`
	Terminate *HeredocEnd        `parser:"@@"`
}

type CallExpr struct {
	Pos  lexer.Position
	Name *Ident    `parser:"@@"`
	Args *ExprList `parser:"@@"`
}

type ExprList struct {
	Pos        lexer.Position
	OpenParen  *OpenParen   `parser:"@@"`
	Fields     []*ExprField `parser:"@@*"`
	CloseParen *CloseParen  `parser:"@@"`
}

type ExprField struct {
	Pos     lexer.Position
	Expr    *Expr    `parser:"( @@ Delimit?"`
	Newline *Newline `parser:"| @@"`
	Comment *Comment `parser:"| @@ )"`
}

type IdentExpr struct {
	Pos       lexer.Position
	Ident     *Ident `parser:"@@"`
	Reference *Ident `parser:"(Dot @@)?"`
}

type Ident struct {
	Pos  lexer.Position
	Text string `parser:"@Ident"`
}

type Newline struct {
	Pos  lexer.Position
	Text string `parser:"@Newline"`
}

type CommentGroup struct {
	Pos      lexer.Position
	Comments []*Comment `parser:"@@+"`
}

type Comment struct {
	Pos  lexer.Position
	Text string `parser:"@Comment"`
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

	err = Parser.Parse(f, mod)
	repr.Println(mod)
	return err
}
