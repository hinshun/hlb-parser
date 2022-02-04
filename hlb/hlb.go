package hlb

import (
	"fmt"
	"strings"

	participle "github.com/alecthomas/participle/v2"
	"github.com/hinshun/hlb-parser/ast"
)

type decl interface {
	toDecl() *ast.Decl
}

type expr interface {
	toExpr() *ast.Expr
	toExprStmt() *ast.ExprStmt
}

type stmt interface {
	toStmt() *ast.Stmt
}

func Module(nodes ...interface{}) *ast.Module {
	var (
		comments *ast.Comments
		decls    []*ast.Decl
	)
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.Comments:
			if comments == nil {
				comments = &ast.Comments{}
			}
			comments.Comments = append(comments.Comments, n.Comments...)
		case *ast.Comment:
			if comments == nil {
				comments = &ast.Comments{}
			}
			comments.Comments = append(comments.Comments, n)
		case *ast.Decl:
			decls = append(decls, n)
		case decl:
			decls = append(decls, n.toDecl())
		default:
			panic(fmt.Sprintf("unknown decl %v", n))
		}
	}

	return &ast.Module{
		Comments: comments,
		Decls:    decls,
	}
}

type importDecl struct {
	name string
	from expr
}

func Import(name string) *importDecl {
	return &importDecl{name: name}
}

func (id *importDecl) From(expr expr) *importDecl {
	id.from = expr
	return id
}

func (id *importDecl) toDecl() *ast.Decl {
	return &ast.Decl{
		Import: &ast.ImportDecl{
			Import: &ast.Import{Text: "import"},
			Name:   &ast.Ident{Text: id.name},
			From:   &ast.From{Text: "from"},
			Expr:   id.from.toExpr(),
		},
	}
}

type funcDecl struct {
	name    string
	export  bool
	params  []*ast.FieldStmt
	returns *ast.Type
	effects []*ast.FieldStmt
	body    []*ast.Stmt
}

func Func(name string) *funcDecl {
	return &funcDecl{name: name}
}

func (fd *funcDecl) Export() *funcDecl {
	fd.export = true
	return fd
}

func (fd *funcDecl) Params(nodes ...interface{}) *funcDecl {
	if len(nodes)%2 != 0 {
		panic(fmt.Sprintf("params must be given in pairs, got %d", len(nodes)))
	}

	var stmts []*ast.FieldStmt
	for i := 0; i < len(nodes); i += 2 {
		t, v := nodes[i], nodes[i+1]
		field := &ast.Field{}

		switch n := t.(type) {
		case string:
			field.Type = astType(n)
		case *ast.Field:
			field = n
		default:
			panic(fmt.Sprintf("unknown field type %v", n))
		}

		switch n := v.(type) {
		case string:
			field.Name = &ast.Ident{Text: n}
		default:
			panic(fmt.Sprintf("unknown field name %v", n))
		}

		stmts = append(stmts, &ast.FieldStmt{Field: field})
	}
	fd.params = stmts
	return fd
}

func (fd *funcDecl) Returns(t string) *funcDecl {
	fd.returns = &ast.Type{
		Scalar: &ast.Ident{Text: t},
	}
	return fd
}

func (fd *funcDecl) Body(nodes ...interface{}) *funcDecl {
	fd.body = nodesToStmts(nodes...)
	return fd
}

func (fd *funcDecl) toDecl() *ast.Decl {
	fun := &ast.FuncDecl{
		Func: &ast.Func{Text: "func"},
		Name: &ast.Ident{Text: fd.name},
		Type: fd.returns,
	}
	if fd.export {
		fun.Modifiers = append(fun.Modifiers, &ast.Modifier{
			Export: &ast.Export{Text: "export"},
		})
	}
	if len(fd.params) > 0 {
		fun.Params = &ast.FieldList{
			Fields: fd.params,
		}
	}
	if len(fd.effects) > 0 {
		fun.Effects = &ast.FieldList{
			Fields: fd.effects,
		}
	}
	if len(fd.body) > 0 {
		fun.Body = &ast.StmtList{
			Stmts: fd.body,
		}
	}
	return &ast.Decl{Func: fun}
}

type ident struct {
	text string
	args []*ast.ExprStmt
	with *ast.WithClause
	as   *ast.AsClause
	at   *ast.AtClause
}

func Ident(text string) *ident {
	return &ident{text: text}
}

func (i *ident) Call(args ...interface{}) *ident {
	var exprs []*ast.ExprStmt
	for _, arg := range args {
		switch a := arg.(type) {
		case string:
			exprs = append(exprs, &ast.ExprStmt{
				Expr: parseLiteral(a),
			})
		case expr:
			exprs = append(exprs, a.toExprStmt())
		}
	}
	i.args = exprs
	return i
}

func (i *ident) With(node interface{}) *ident {
	var with *ast.Expr
	switch n := node.(type) {
	case *ast.Expr:
		with = n
	case expr:
		with = n.toExpr()
	}
	i.with = &ast.WithClause{
		With: &ast.With{Text: "with"},
		Expr: with,
	}
	return i
}

func (i *ident) As(s string) *ident {
	i.as = &ast.AsClause{
		As:     &ast.As{Text: "as"},
		Effect: &ast.Ident{Text: s},
	}
	return i
}

func (i *ident) At(s string) *ident {
	i.at = &ast.AtClause{
		At:     &ast.At{Text: "@"},
		Effect: &ast.Ident{Text: s},
	}
	return i
}

func (i *ident) toExpr() *ast.Expr {
	ref := &ast.Ref{
		Terminal: &ast.Terminal{
			Ident: &ast.Ident{
				Text: i.text,
			},
		},
	}

	if len(i.args) > 0 || i.with != nil || i.as != nil || i.at != nil {
		call := &ast.Call{
			Args: &ast.ExprList{
				Exprs: i.args,
			},
			With: i.with,
			As:   i.as,
			At:   i.at,
		}
		ref.Next = &ast.RefNext{Call: call}
	}
	return refExpr(ref)
}

func (i *ident) toExprStmt() *ast.ExprStmt {
	return &ast.ExprStmt{
		Expr: i.toExpr(),
	}
}

func refExpr(ref *ast.Ref) *ast.Expr {
	return &ast.Expr{
		Unary: &ast.Unary{
			Ref: ref,
		},
	}
}

func terminalExpr(term *ast.Terminal) *ast.Expr {
	return refExpr(&ast.Ref{
		Terminal: term,
	})
}

func literalExpr(lit *ast.Literal) *ast.Expr {
	return terminalExpr(&ast.Terminal{
		Lit: lit,
	})
}

func parseLiteral(str string) *ast.Expr {
	expr, err := _parseLiteral(str)
	if err != nil {
		expr, err = _parseLiteral(fmt.Sprintf("%q", str))
		if err != nil {
			panic(err)
		}
	}
	return expr
}

func _parseLiteral(str string) (*ast.Expr, error) {
	parser := participle.MustBuild(
		&ast.Literal{},
		participle.Lexer(ast.Lexer),
	)

	lit := &ast.Literal{}
	err := parser.Parse("", strings.NewReader(str), lit)
	if err != nil {
		return nil, err
	}

	return literalExpr(lit), nil
}

func astType(str string) *ast.Type {
	parser := participle.MustBuild(
		&ast.Type{},
		participle.Lexer(ast.Lexer),
	)

	t := &ast.Type{}
	err := parser.Parse("", strings.NewReader(str), t)
	if err != nil {
		panic(err)
	}
	return t
}

func Variadic(t string) *ast.Field {
	variadic := "..."
	return &ast.Field{
		Type:     astType(t),
		Variadic: &variadic,
	}
}

func Comments(cs ...string) *ast.Stmt {
	var comments []*ast.Comment
	for _, c := range cs {
		comments = append(comments, &ast.Comment{
			Text: fmt.Sprintf("# %s\n", c),
		})
	}
	return &ast.Stmt{Comments: &ast.Comments{Comments: comments}}
}

type array struct {
	t     string
	items []*ast.Stmt
}

func Array(t string) *array {
	return &array{t: t}
}

func (a *array) Items(nodes ...interface{}) *array {
	a.items = nodesToStmts(nodes...)
	return a
}

func (a *array) toExpr() *ast.Expr {
	lit := &ast.ArrayLit{
		Type: &ast.Ident{Text: a.t},
	}
	if len(a.items) > 0 {
		lit.Block = &ast.StmtList{
			Stmts: a.items,
		}
	}
	return literalExpr(&ast.Literal{
		ArrayLit: lit,
	})
}

func (a *array) toExprStmt() *ast.ExprStmt {
	return &ast.ExprStmt{
		Expr: a.toExpr(),
	}
}

func nodesToStmts(nodes ...interface{}) []*ast.Stmt {
	var stmts []*ast.Stmt
	for _, node := range nodes {
		switch n := node.(type) {
		case *ast.Stmt:
			stmts = append(stmts, n)
		case stmt:
			stmts = append(stmts, n.toStmt())
		case expr:
			stmts = append(stmts, &ast.Stmt{
				Expr: n.toExpr(),
			})
		default:
			panic(fmt.Sprintf("unknown stmt %v", n))
		}
	}
	return stmts
}
