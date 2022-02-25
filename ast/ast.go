package ast

import (
	"strconv"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	Lexer = lexer.MustStateful(lexer.Rules{
		"Root": {
			{"whitespace", `[\r\t ]+`, nil},
			{"Modifier", `\b(pub)\b`, nil},
			{"Keyword", `\b(if|else|for|in|with|as|import|fun)\b`, nil},
			{"Numeric", `\b(0(b|B|o|O|x|X)[a-fA-F0-9]+)\b`, nil},
			{"Decimal", `\b(0|[1-9][0-9]*)\b`, nil},
			{"Bool", `\b(true|false)\b`, nil},
			{"String", `"`, lexer.Push("String")},
			{"RawString", "`", lexer.Push("RawString")},
			{"Heredoc", `<<[-~]?(\w+)`, lexer.Push("Heredoc")},
			{"RawHeredoc", "<<[-~]?`(\\w+)`", lexer.Push("RawHeredoc")},
			{"Brace", `{`, lexer.Push("Brace")},
			{"Paren", `\(`, lexer.Push("Paren")},
			{"Ident", `\b([[:alpha:]_]\w*)\b`, nil},
			{"Operator", `(>=|<=|&&|\|\||==|!=|[-+*/%^!|&])`, nil},
			{"Punct", `[][@:;?.,]`, nil},
			{"Newline", `\n`, nil},
			{"Comment", `#`, lexer.Push("Comment")},
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
			{"BraceEnd", `}`, lexer.Pop()},
			lexer.Include("Root"),
		},
		"Brace": {
			{"BraceEnd", `}`, lexer.Pop()},
			lexer.Include("Root"),
		},
		"Paren": {
			{"ParenEnd", `\)`, lexer.Pop()},
			lexer.Include("Root"),
		},
		"Comment": {
			{"CommentEnd", `\n`, lexer.Pop()},
			{"CommentText", `[^\n]`, nil},
		},
	})

	Parser = participle.MustBuild(
		&Module{},
		participle.Lexer(&semicolonLexerDefinition{}),
	)
)

type Module struct {
	Comments *Comments `parser:"@@?"`

	Decls []*Decl `parser:"@@*"`
}

type Decl struct {
	Import   *ImportDecl `parser:"( @@ (?= ';' | Comment) ';'?"`
	Func     *FuncDecl   `parser:"| @@ (?= ';' | Comment) ';'?"`
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

type FuncDecl struct {
	Modifiers []*Modifier `parser:"@@*"`
	Func      *Func       `parser:"@@"`
	Name      *Ident      `parser:"@@"`
	Params    *FieldList  `parser:"@@"`
	Type      *Type       `parser:"@@"`
	Effects   *FieldList  `parser:"@@?"`
	Body      *StmtList   `parser:"@@?"`
}

type Modifier struct {
	Public *Public `@@`
}

type Public struct {
	Text string `parser:"@'pub'"`
}

type Func struct {
	Text string `parser:"@'fun'"`
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

type Type struct {
	Scalar      *Ident       `parser:"( @@"`
	Array       *Ident       `parser:"| '[' ']' @@ )"`
	Association *Association `parser:"@@?"`
}

type Association struct {
	Symbol string `parser:"@( ':' ':' )"`
	Ident  *Ident `parser:"@@"`
}

type StmtList struct {
	OpenBrace  *OpenBrace  `parser:"@@"`
	Stmts      []*Stmt     `parser:"@@*"`
	CloseBrace *CloseBrace `parser:"@@"`
}

type Stmt struct {
	If       *IfStmt   `parser:"( @@ (?= ';' | Comment) ';'?"`
	For      *ForStmt  `parser:"| @@ (?= ';' | Comment) ';'?"`
	Expr     *Expr     `parser:"| @@ (?= ';' | Comment) ';'?"`
	Newline  *Newline  `parser:"| @@"`
	Comments *Comments `parser:"| @@ )"`
}

type IfStmt struct {
	If        *If           `parser:"@@"`
	Condition *Condition    `parser:"@@"`
	Body      *StmtList     `parser:"@@"`
	ElseIfs   []*ElseIfStmt `parser:"@@*"`
	Else      *ElseStmt     `parser:"@@?"`
}

type Condition struct {
	OpenParen  *OpenParen  `parser:"@@"`
	Expr       *Expr       `parser:"@@"`
	CloseParen *CloseParen `parser:"@@"`
}

type ElseIfStmt struct {
	Else      *Else      `parser:"@@"`
	If        *If        `parser:"@@"`
	Condition *Condition `parser:"@@"`
	Body      *StmtList  `parser:"@@"`
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
	For    *For       `parser:"@@"`
	Header *ForHeader `parser:"@@"`
	Body   *StmtList  `parser:"@@"`
}

type For struct {
	Text string `parser:"@'for'"`
}

type ForHeader struct {
	OpenParen  *OpenParen  `parser:"@@"`
	Counter    *Ident      `parser:"( @@ ',' )?"`
	Var        *Ident      `parser:"@@"`
	In         *In         `parser:"@@"`
	Iterable   *Expr       `parser:"@@"`
	CloseParen *CloseParen `parser:"@@"`
}

type In struct {
	Text string `parser:"@'in'"`
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
	Ident *Ident `parser:"@@"`
}

type Literal struct {
	Block      *BlockLit     `parser:"( @@"`
	Decimal    *int          `parser:"| @Decimal"`
	Numeric    *NumericLit   `parser:"| @Numeric"`
	Bool       *bool         `parser:"| @Bool"`
	String     *StringLit    `parser:"| @@"`
	RawString  *RawStringLit `parser:"| @@"`
	Heredoc    *Heredoc      `parser:"| @@"`
	RawHeredoc *RawHeredoc   `parser:"| @@ )"`
}

type BlockLit struct {
	Type  *Type     `parser:"@@?"`
	Block *StmtList `parser:"@@"`
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
	End       *Quote            `parser:"@@"`
}

type Quote struct {
	Text string `parser:"@(String | StringEnd)"`
}

type StringFragment struct {
	Escaped      *string       `parser:"( @Escaped"`
	Interpolated *Interpolated `parser:"| @@"`
	Text         *string       `parser:"| @Char )"`
}

type RawStringLit struct {
	Start *Backtick `parser:"@@"`
	Text  string    `parser:"@RawChar"`
	End   *Backtick `parser:"@@"`
}

type Backtick struct {
	Text string `parser:"@(RawString | RawStringEnd)"`
}

type Heredoc struct {
	Value     string
	Start     string             `parser:"@Heredoc"`
	Fragments []*HeredocFragment `parser:"@@*"`
	End       *HeredocEnd        `parser:"@@"`
}

type HeredocFragment struct {
	Spaces       *string       `parser:"( @Spaces"`
	Escaped      *string       `parser:"| @Escaped"`
	Interpolated *Interpolated `parser:"| @@"`
	Text         *string       `parser:"| @(Text | RawText) )"`
}

type HeredocEnd struct {
	Text string `parser:"@(HeredocEnd | RawHeredocEnd)"`
}

type RawHeredoc struct {
	Start     string             `parser:"@RawHeredoc"`
	Fragments []*HeredocFragment `parser:"@@*"`
	End       *HeredocEnd        `parser:"@@"`
}

type Interpolated struct {
	Start *OpenInterpolated `parser:"@@"`
	Expr  *Expr             `parser:"@@?"`
	End   *CloseBrace       `parser:"@@"`
}

type OpenInterpolated struct {
	Text string `parser:"@Interpolated"`
}

type Ident struct {
	Text string `parser:"@Ident"`
}

type Newline struct {
	Text string `parser:"@Newline"`
}

type Comments struct {
	Comments []*Comment `parser:"@@+"`
}

type Comment struct {
	Text string `parser:"Comment @(CommentText*) CommentEnd"`
}

type OpenBrace struct {
	Text string `parser:"@Brace"`
}

type CloseBrace struct {
	Text string `parser:"@BraceEnd"`
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
