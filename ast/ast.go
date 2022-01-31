package ast

import (
	"fmt"
	"strconv"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	Lexer = lexer.MustStateful(lexer.Rules{
		"Root": {
			{"whitespace", `[\r\t ]+`, nil},
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
			{"Ident", `\b([[:alpha:]_]\w*)\b`, nil},
			{"Operator", `(>=|<=|&&|\|\||==|!=|[-+*/<>%^!|&])`, nil},
			{"Punct", `[][@:;?.,]`, nil},
			{"Newline", `\n`, nil},
			{"Comment", `#[^\n]*\n`, nil},
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

	unaryParser = participle.MustBuild(
		&Unary{},
		participle.Lexer(Lexer),
	)

	operatorToken = Lexer.Symbols()["Operator"]
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

type OpenBracket struct {
	Text string `parser:"@'['"`
}

type CloseBracket struct {
	Text string `parser:"@']'"`
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
	Unary *Unary

	Left  *Expr
	Op    Op
	Right *Expr
}

// Parse expressions with a custom precedence climbing implementation.
func (e *Expr) Parse(lex *lexer.PeekingLexer) error {
	ex, err := parseExpr(lex, 0)
	if err != nil {
		return err
	}
	*e = *ex
	return nil
}

type Op int

const (
	OpNone   Op = iota //
	OpGe               // >=
	OpLe               // <=
	OpAnd              // &&
	OpOr               // ||
	OpEq               // ==
	OpNe               // !=
	OpSub              // -
	OpAdd              // +
	OpMul              // *
	OpDiv              // /
	OpLt               // <
	OpGt               // >
	OpMod              // %
	OpPow              // ^
	OpNot              // !
	OpBitOr            // |
	OpBitAnd           // &
)

func (o *Op) Capture(values []string) error {
	switch values[0] {
	case ">=":
		*o = OpGe
	case "<=":
		*o = OpLe
	case "&&":
		*o = OpAnd
	case "||":
		*o = OpOr
	case "==":
		*o = OpEq
	case "!=":
		*o = OpNe
	case "-":
		*o = OpSub
	case "+":
		*o = OpAdd
	case "*":
		*o = OpMul
	case "/":
		*o = OpDiv
	case "%":
		*o = OpMod
	case "<":
		*o = OpLt
	case ">":
		*o = OpGt
	case "^":
		*o = OpPow
	case "!":
		*o = OpNot
	case "|":
		*o = OpBitOr
	case "&":
		*o = OpBitAnd
	default:
		return fmt.Errorf("invalid expression operator %q", values[0])
	}
	return nil
}

type opInfo struct {
	RightAssociative bool
	Priority         int
}

var opTable = map[Op]opInfo{
	OpAdd:    {Priority: 1},
	OpSub:    {Priority: 1},
	OpMul:    {Priority: 2},
	OpDiv:    {Priority: 2},
	OpMod:    {Priority: 2},
	OpPow:    {RightAssociative: true, Priority: 3},
	OpBitOr:  {Priority: 4},
	OpBitAnd: {Priority: 4},
}

// Precedence climbing implementation based on
// https://eli.thegreenplace.net/2012/08/02/parsing-expressions-by-precedence-climbing
func parseExpr(lex *lexer.PeekingLexer, minPrec int) (*Expr, error) {
	lhs, err := parseOperand(lex)
	if err != nil {
		return nil, err
	}

	for {
		token, err := lex.Peek(0)
		if err != nil {
			return nil, err
		}
		if token.Type != operatorToken {
			break
		}

		expr := &Expr{}
		err = expr.Op.Capture([]string{token.Value})
		if err != nil {
			return lhs, nil
		}
		if opTable[expr.Op].Priority < minPrec {
			break
		}

		_, _ = lex.Next()
		nextMinPrec := opTable[expr.Op].Priority
		if !opTable[expr.Op].RightAssociative {
			nextMinPrec++
		}

		rhs, err := parseExpr(lex, nextMinPrec)
		if err != nil {
			return nil, err
		}

		expr.Left = lhs
		expr.Right = rhs
		lhs = expr
	}

	return lhs, nil
}

func parseOperand(lex *lexer.PeekingLexer) (*Expr, error) {
	// tok, _ := lex.Peek(0)

	u := &Unary{}
	err := unaryParser.ParseFromLexer(lex, u, participle.AllowTrailing(true))
	if err != nil {
		return nil, err
	}
	return &Expr{Unary: u}, nil
}

type Unary struct {
	Op  Op   `parser:"@( '!' | '-' )?"`
	Ref *Ref `parser:"@@"`
}

type Ref struct {
	Terminal *Terminal `parser:"@@"`
	Next     *RefNext  `parser:"@@?"`
}

type Terminal struct {
	Lit   *Literal `parser:"( @@"`
	Ident *Ident   `parser:"| @@ )"`
}

type RefNext struct {
	Splat     *Splat     `parser:"( @@"`
	Subscript *Subscript `parser:"| @@"`
	Selector  *Selector  `parser:"| @@"`
	Call      *Call      `parser:"| @@ )"`
	Next      *RefNext   `@@?`
}

type Splat struct {
	Text string `parser:"@('.' '.' '.')"`
}

type Subscript struct {
	OpenBracket  *OpenBracket  `parser:"@@"`
	Expr         *Expr         `parser:"@@"`
	CloseBracket *CloseBracket `parser:"@@"`
}

type Selector struct {
	Dot   string `parser:"@'.'"`
	Field *Ident `parser:"@@"`
}

type Literal struct {
	Block      *StmtList     `parser:"( @@"`
	Decimal    *int          `parser:"| @Decimal"`
	Numeric    *NumericLit   `parser:"| @Numeric"`
	Bool       *bool         `parser:"| @Bool"`
	String     *StringLit    `parser:"| @@"`
	RawString  *RawStringLit `parser:"| @@"`
	Heredoc    *Heredoc      `parser:"| @@"`
	RawHeredoc *RawHeredoc   `parser:"| @@ )"`
}

type Call struct {
	Args *ExprList   `parser:"@@?"`
	At   *AtClause   `parser:"@@?"`
	With *WithClause `parser:"@@?"`
	As   *AsClause   `parser:"@@?"`
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
	Scalar      *Ident       `parser:"( @@"`
	Array       *Ident       `parser:"| '[' ']' @@ )"`
	Association *Association `parser:"@@?"`
}

type Association struct {
	Symbol string `parser:"@( ':' ':' )"`
	Field  *Ident `parser:"@@"`
}

type Ident struct {
	Text string `parser:"@Ident"`
}

type Newline struct {
	Text string `parser:"@Newline"`
}
