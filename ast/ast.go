package ast

import (
	"strconv"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	// Lexer lexes HLB into tokens for the parser.
	Lexer = lexer.MustStateful(lexer.Rules{
		"Root": {
			{"whitespace", `[\r\t ]+`, nil},
			{"Keyword", `\b(if|else|for|in|match|with|as)\b`, nil},
			{"Numeric", `\b(0(b|B|o|O|x|X)[a-fA-F0-9]+)\b`, nil},
			{"Decimal", `\b(0|[1-9][0-9]*)\b`, nil},
			{"Bool", `\b(true|false)\b`, nil},
			{"String", `"`, lexer.Push("String")},
			{"RawString", "`", lexer.Push("RawString")},
			{"Heredoc", `<<[-~]?(\w+)`, lexer.Push("Heredoc")},
			{"RawHeredoc", "<<[-~]?`(\\w+)`", lexer.Push("RawHeredoc")},
			{"Brace", `{`, lexer.Push("Brace")},
			{"Paren", `\(`, lexer.Push("Paren")},
			{"Bracket", `\[`, lexer.Push("Bracket")},
			{"Ident", `\b([[:alpha:]_]\w*)\b`, nil},
			{"Assignment", `(\^=|\+=|-=|\*=|/=|\|=|&=|%=|=)`, nil},
			{"Operator", `(>=|<=|&&|\|\||==|!=|[-~+*/%<>^!|&])`, nil},
			{"Punct", `[@:;?.,]`, nil},
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
			{"HeredocEnd", `\b\1\b`, lexer.Pop()},
			{"Spaces", `\s+`, nil},
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
		"Bracket": {
			{"BracketEnd", `\]`, lexer.Pop()},
			lexer.Include("Root"),
		},
		"Comment": {
			{"CommentEnd", `\n`, lexer.Pop()},
			{"CommentText", `[^\n]`, nil},
		},
	})

	Parser = participle.MustBuild(
		&Module{},
		participle.Lexer(Lexer),
	)
)

type Module struct {
	Attrs []*AttrStmt `parser:"@@*"`
}

type AttrStmt struct {
	Newline  *Newline  `parser:"( @@"`
	Comments *Comments `parser:"| @@"`
	Attr     *Attr     `parser:"| @@ )"`
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

type Attr struct {
	Keys     []*Ident  `parser:"( (@@ ':')*"`
	Destruct *Destruct `parser:"(@@ ':')? )!"`
	Expr     *Expr     `parser:"@@"`
}

type Destruct struct {
	Start  *OpenBrace  `parser:"@@"`
	Idents []*Ident    `parser:"@@ (',' @@)*"`
	Fin    *CloseBrace `parser:"@@"`
}

type Ident struct {
	Text string `parser:"@Ident"`
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
	Func  *Func  `parser:"( @@"`
	Group *Group `parser:"| @@"`
	Lit   *Lit   `parser:"| @@ )"`
}

type Func struct {
	Params *FieldList `parser:"@@"`
	At     *AtList    `parser:"@@?"`
	With   *WithList  `parser:"@@?"`
	Type   *Type      `parser:"'-' '>' @@"`
	Body   *StmtList  `parser:"@@?"`
}

type FieldList struct {
	Start  *OpenParen   `parser:"@@"`
	Fields []*FieldStmt `parser:"@@*"`
	Fin    *CloseParen  `parser:"@@"`
}

type FieldStmt struct {
	Newline  *Newline  `parser:"( @@"`
	Comments *Comments `parser:"| @@"`
	Field    *Field    `parser:"| @@ ','? )"`
}

type Field struct {
	Name *Ident `parser:"@@ ':'"`
	Type *Type  `parser:"@@"`
}

type AtList struct {
	At   *At        `parser:"@@"`
	List *FieldList `parser:"@@"`
}

type WithList struct {
	With *With     `parser:"@@"`
	List *AttrList `parser:"@@"`
}

type AttrList struct {
	Start *OpenParen  `parser:"@@"`
	Attrs []*AttrStmt `parser:"@@*"`
	Fin   *CloseParen `parser:"@@"`
}

type Type struct {
	Scalar  *Ident   `parser:"( @@"`
	Array   *Type    `parser:"| '[' ']' @@ )"`
	Subtype *Subtype `parser:"@@?"`
}

type Subtype struct {
	OpenAngle  string `parser:"@'<'"`
	Type       *Type  `parser:"@@"`
	CloseAngle string `parser:"@'>'"`
}

type StmtList struct {
	Start *OpenBrace  `parser:"@@"`
	Stmts []*Stmt     `parser:"@@*"`
	Fin   *CloseBrace `parser:"@@"`
}

type Stmt struct {
	Newline  *Newline   `parser:"( @@"`
	Comments *Comments  `parser:"| @@"`
	If       *IfStmt    `parser:"| @@ ';'?"`
	For      *ForStmt   `parser:"| @@ ';'?"`
	Match    *MatchStmt `parser:"| @@ ';'?"`
	Attr     *Attr      `parser:"| @@ ';'?"`
	Expr     *ExprStmt  `parser:"| @@ ';'? )"`
}

type IfStmt struct {
	If      *If           `parser:"@@"`
	Cond    *Group        `parser:"@@"`
	Body    *StmtList     `parser:"@@"`
	ElseIfs []*ElseIfStmt `parser:"@@*"`
	Else    *ElseStmt     `parser:"@@?"`
}

type If struct {
	Text string `parser:"@'if'"`
}

type Group struct {
	Start *OpenParen  `parser:"@@"`
	Expr  *Expr       `parser:"@@"`
	Fin   *CloseParen `parser:"@@"`
}

type ElseIfStmt struct {
	Else *Else     `parser:"@@"`
	If   *If       `parser:"@@"`
	Cond *Group    `parser:"@@"`
	Body *StmtList `parser:"@@"`
}

type Else struct {
	Text string `parser:"@'else'"`
}

type ElseStmt struct {
	Else *Else     `parser:"@@"`
	Body *StmtList `parser:"@@"`
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
	Start    *OpenParen  `parser:"@@"`
	Counter  *Ident      `parser:"( @@ ',' )?"`
	Var      *Ident      `parser:"@@"`
	In       *In         `parser:"@@"`
	Iterable *Expr       `parser:"@@"`
	Fin      *CloseParen `parser:"@@"`
}

type In struct {
	Text string `parser:"@'in'"`
}

type MatchStmt struct {
	Match *MatchKW `parser:"@@"`
	Group *Group   `parser:"@@"`
	Body  *ArmList `parser:"@@"`
}

type MatchKW struct {
	Text string `parser:"@'match'"`
}

type ArmList struct {
	Start *OpenBrace  `parser:"@@"`
	Arms  []*ArmStmt  `parser:"@@*"`
	Fin   *CloseBrace `parser:"@@"`
}

type ArmStmt struct {
	Newline  *Newline  `parser:"( @@"`
	Comments *Comments `parser:"| @@"`
	Arm      *Arm      `parser:"| @@ ';'? )"`
}

type Arm struct {
	Pattern *Expr `parser:"@@ ':'"`
	Expr    *Expr `parser:"@@"`
}

type ExprStmt struct {
	LHS *Expr `parser:"@@"`
	Op  Op    `parser:"( @Assignment"`
	RHS *Expr `parser:"@@ )?"`
}

type Lit struct {
	Value      *Type         `parser:"( ( @@?"`
	Block      *StmtList     `parser:"@@? )!"`
	Decimal    *int          `parser:"| @Decimal"`
	Numeric    *NumericLit   `parser:"| @Numeric"`
	Bool       *string       `parser:"| @Bool"`
	Str        *StringLit    `parser:"| @@"`
	RawStr     *RawStringLit `parser:"| @@"`
	Heredoc    *Heredoc      `parser:"| @@"`
	RawHeredoc *RawHeredoc   `parser:"| @@ )"`
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
	Fin       *Quote            `parser:"@@"`
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
	Fin   *Backtick `parser:"@@"`
}

type Backtick struct {
	Text string `parser:"@(RawString | RawStringEnd)"`
}

type Heredoc struct {
	Value     string
	Start     string             `parser:"@Heredoc"`
	Fragments []*HeredocFragment `parser:"@@*"`
	Fin       *HeredocEnd        `parser:"@@"`
}

type HeredocFragment struct {
	Spaces       *string       `parser:"( @Spaces"`
	Escaped      *string       `parser:"| @Escaped"`
	Interpolated *Interpolated `parser:"| @@"`
	Text         *string       `parser:"| @(Text | RawText) )"`
}

type HeredocEnd struct {
	Text string `parser:"@HeredocEnd"`
}

type RawHeredoc struct {
	Start     string             `parser:"@RawHeredoc"`
	Fragments []*HeredocFragment `parser:"@@*"`
	Fin       *HeredocEnd        `parser:"@@"`
}

type Interpolated struct {
	Start *OpenInterpolated `parser:"@@"`
	Expr  *Expr             `parser:"@@?"`
	Fin   *CloseBrace       `parser:"@@"`
}

type OpenInterpolated struct {
	Text string `parser:"@Interpolated"`
}

type RefNext struct {
	Subscript *Subscript `parser:"( @@"`
	Selector  *Selector  `parser:"| @@"`
	Splat     *Splat     `parser:"| @@"`
	Call      *Call      `parser:"| @@"`
	As        *AsClause  `parser:"| @@ )"`
	Next      *RefNext   `@@?`
}

type Subscript struct {
	Start *OpenBracket  `parser:"@@"`
	Left  *Expr         `parser:"( @@?"`
	Colon *string       `parser:"@':'?"`
	Right *Expr         `parser:"@@? )!"`
	Fin   *CloseBracket `parser:"@@"`
}

type Selector struct {
	Dot   string `parser:"@'.'"`
	Ident *Ident `parser:"@@"`
}

type Splat struct {
	Text string `parser:"@('.' '.' '.')"`
}

type Call struct {
	Args *ArgList    `parser:"@@"`
	At   *AtClause   `parser:"@@?"`
	With *WithClause `parser:"@@?"`
}

type ArgList struct {
	Start *OpenParen  `parser:"@@"`
	Args  []*ArgStmt  `parser:"@@*"`
	Fin   *CloseParen `parser:"@@"`
}

type ArgStmt struct {
	Newline  *Newline  `parser:"( @@"`
	Comments *Comments `parser:"| @@"`
	Arg      *Arg      `parser:"| @@ ','? )"`
}

type Arg struct {
	Field *Ident `parser:"( @@ ':' )?"`
	Expr  *Expr  `parser:"@@"`
}

type AtClause struct {
	At     *At    `parser:"@@"`
	Effect *Ident `parser:"@@"`
}

type At struct {
	Text string `parser:"@'@'"`
}

type WithClause struct {
	With *With `parser:"@@"`
	Expr *Expr `parser:"@@"`
}

type With struct {
	Text string `parser:"@'with'"`
}

type AsClause struct {
	As    *As         `parser:"@@"`
	Ident *Ident      `parser:"( @@"`
	List  *AssignList `parser:"| @@ )"`
}

type As struct {
	Text string `parser:"@'as'"`
}

type AssignList struct {
	Start   *OpenParen    `parser:"@@"`
	Assigns []*AssignStmt `parser:"@@*"`
	Fin     *CloseParen   `parser:"@@"`
}

type AssignStmt struct {
	Newline  *Newline  `parser:"( @@"`
	Comments *Comments `parser:"| @@"`
	Assign   *Assign   `parser:"| @@ ';'? )"`
}

type Assign struct {
	Name   *Ident `parser:"@@ ':'"`
	Effect *Ident `parser:"'@' @@"`
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
	Text string `parser:"@Bracket"`
}

type CloseBracket struct {
	Text string `parser:"@BracketEnd"`
}
