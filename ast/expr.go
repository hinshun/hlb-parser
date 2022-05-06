package ast

import (
	"fmt"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var (
	unaryParser = participle.MustBuild(
		&Unary{},
		participle.Lexer(Lexer),
	)

	operatorToken = Lexer.Symbols()["Operator"]
)

type Expr struct {
	Unary *Unary

	Left  *Expr
	Op    Op
	Right *Expr
}

// Parse expressions with a precedence climbing implementation.
func (e *Expr) Parse(lex *lexer.PeekingLexer) error {
	ex, err := parseExpr(lex, 0)
	if err != nil {
		return err
	}
	*e = *ex
	return nil
}

type Op string

const (
	OpNone    Op = ""
	OpAsgn    Op = "="
	OpAddAsgn Op = "+="
	OpSubAsgn Op = "-="
	OpMulAsgn Op = "*="
	OpDivAsgn Op = "/="
	OpModAsgn Op = "%="
	OpPowAsgn Op = "^="
	OpGe      Op = ">="
	OpLe      Op = "<="
	OpAnd     Op = "&&"
	OpOr      Op = "||"
	OpEq      Op = "=="
	OpNe      Op = "!="
	OpSub     Op = "-"
	OpAdd     Op = "+"
	OpMul     Op = "*"
	OpDiv     Op = "/"
	OpLt      Op = "<"
	OpGt      Op = ">"
	OpMod     Op = "%"
	OpPow     Op = "^"
	OpNot     Op = "!"
	OpMrg     Op = "&"
	OpUn      Op = "|"
)

func (o *Op) Capture(values []string) error {
	switch values[0] {
	case "%=":
		*o = OpModAsgn
	case "+=":
		*o = OpAddAsgn
	case "-=":
		*o = OpSubAsgn
	case "/=":
		*o = OpDivAsgn
	case "*=":
		*o = OpMulAsgn
	case "^=":
		*o = OpPowAsgn
	case "=":
		*o = OpAsgn
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
	case "&":
		*o = OpMrg
	case "|":
		*o = OpUn
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
	OpAdd: {Priority: 1},
	OpSub: {Priority: 1},
	OpMul: {Priority: 2},
	OpDiv: {Priority: 2},
	OpMod: {Priority: 2},
	OpPow: {RightAssociative: true, Priority: 3},
	OpMrg: {Priority: 4},
	OpUn:  {Priority: 4},
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
	u := &Unary{}
	err := unaryParser.ParseFromLexer(lex, u, participle.AllowTrailing(true))
	if err != nil {
		return nil, err
	}
	return &Expr{Unary: u}, nil
}
