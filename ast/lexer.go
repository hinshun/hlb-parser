package ast

import (
	"io"

	"github.com/alecthomas/participle/v2/lexer"
)

var (
	parenToken      = Lexer.Symbols()["Paren"]
	braceToken      = Lexer.Symbols()["Brace"]
	newlineToken    = Lexer.Symbols()["Newline"]
	commentEndToken = Lexer.Symbols()["CommentEnd"]
)

// A Lexer that inserts semi-colons.
type semicolonLexerDefinition struct{}

func (l *semicolonLexerDefinition) Lex(path string, r io.Reader) (lexer.Lexer, error) {
	ll, err := Lexer.Lex(path, r)
	if err != nil {
		return nil, err
	}
	return &semicolonLexer{lexer: ll}, nil
}

func (l *semicolonLexerDefinition) Symbols() map[string]lexer.TokenType {
	return Lexer.Symbols()
}

type semicolonLexer struct {
	lexer lexer.Lexer
	last  lexer.Token
}

func (l *semicolonLexer) Next() (lexer.Token, error) {
	for {
		token, err := l.lexer.Next()
		if err != nil {
			return token, err
		}
		if token.Type != newlineToken {
			l.last = token
			return token, nil
		}

		// Do we need to insert a semi-colon?
		switch l.last.Value {
		case ";", ",":
			l.last = token
			continue
		case "}":
			token.Value = ";"
			token.Type = ';'
		default:
			switch l.last.Type {
			case parenToken, braceToken, newlineToken, commentEndToken:
				l.last = token
				continue
			default:
				token.Value = ";"
				token.Type = ';'
			}
		}
		l.last = token
		return token, nil
	}
}
