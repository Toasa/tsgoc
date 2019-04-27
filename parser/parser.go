package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"tgoc/ast"
	"tgoc/token"
	"tgoc/utils"
)

type Parser struct {
	Tokens []token.Token
	Pos    int
	VarMap map[string]*ast.Ident
	Stmts  []ast.Stmt
}

func New(t []token.Token) *Parser {
	return &Parser{Tokens: t, Pos: 0, VarMap: map[string]*ast.Ident{}, Stmts: []ast.Stmt{}}
}

func (p *Parser) parseTerm() ast.Expr {
	utils.Assert(p.curTokenIs(token.INT) || p.curTokenIs(token.LPAREN), fmt.Sprintf("invalid token: %s", p.curToken().Literal))

	if p.curTokenIs(token.INT) {
		n, _ := strconv.Atoi(p.Tokens[p.Pos].Literal)
		p.nextToken()
		return &ast.IntLit{Val: n}
	}
	if p.curTokenIs(token.LPAREN) {
		p.nextToken()
		node := p.parseExpr()
		utils.Assert(p.curTokenIs(token.RPAREN), fmt.Sprintf("expected RPAREN, but got %s", p.curToken().Literal))
		p.nextToken()
		return node
	}

	return nil
}

func (p *Parser) parseIdent() ast.Expr {
	if p.curTokenIs(token.IDENT) {
		ident, ok := p.VarMap[p.curToken().Literal]
		utils.Assert(ok, fmt.Sprintf("undeclared identifier: %s", p.curToken().Literal))
		p.nextToken()
		return ident
	} else if p.curTokenIs(token.TRUE) {
		p.nextToken()
		return &ast.Boolean{Val: true}
	} else if p.curTokenIs(token.FALSE) {
		p.nextToken()
		return &ast.Boolean{Val: false}
	} else {
		return p.parseTerm()
	}
}

func (p *Parser) parseUnary() ast.Expr {
	var lhs ast.Expr
	if p.curTokenIs(token.SUB) {
		p.nextToken()
		lhs = &ast.UnaryExpr{Op: "-", Expr: p.parseIdent()}
	} else if p.curTokenIs(token.NOT) {
		p.nextToken()
		lhs = &ast.UnaryExpr{Op: "!", Expr: p.parseIdent()}
	} else {
		if p.curTokenIs(token.ADD) {
			p.nextToken()
		}
		lhs = p.parseIdent()
	}
	return lhs
}

func (p *Parser) parseMul() ast.Expr {
	lhs := p.parseUnary()

	for p.curTokenIs(token.MUL) || p.curTokenIs(token.DIV) ||
		p.curTokenIs(token.REM) || p.curTokenIs(token.LSHIFT) ||
		p.curTokenIs(token.RSHIFT) || p.curTokenIs(token.BAND) ||
		p.curTokenIs(token.BCLR) {

		op := p.curToken().Literal
		p.nextToken()
		rhs := p.parseUnary()
		lhs = &ast.BinaryExpr{Op: op, Lhs: lhs, Rhs: rhs}
	}

	return lhs
}

func (p *Parser) parseAdd() ast.Expr {
	lhs := p.parseMul()

	for p.curTokenIs(token.ADD) || p.curTokenIs(token.SUB) ||
		p.curTokenIs(token.BOR) || p.curTokenIs(token.BXOR) {

		op := p.curToken().Literal
		p.nextToken()
		rhs := p.parseMul()
		lhs = &ast.BinaryExpr{Op: op, Lhs: lhs, Rhs: rhs}
	}
	return lhs
}

func (p *Parser) parseComparison() ast.Expr {
	lhs := p.parseAdd()

	for p.curTokenIs(token.EQ) || p.curTokenIs(token.NQ) ||
		p.curTokenIs(token.LT) || p.curTokenIs(token.GT) ||
		p.curTokenIs(token.LTE) || p.curTokenIs(token.GTE) {

		op := p.curToken().Literal
		p.nextToken()
		rhs := p.parseAdd()
		lhs = &ast.LogicalExpr{Op: op, Lhs: lhs, Rhs: rhs}
	}
	return lhs
}

func (p *Parser) parseCAnd() ast.Expr {
	lhs := p.parseComparison()
	for p.curTokenIs(token.CAND) {
		p.nextToken()
		rhs := p.parseComparison()
		lhs = &ast.LogicalExpr{Op: "&&", Lhs: lhs, Rhs: rhs}
	}
	return lhs
}

func (p *Parser) parseCOr() ast.Expr {
	lhs := p.parseCAnd()
	for p.curTokenIs(token.COR) {
		p.nextToken()
		rhs := p.parseCAnd()
		lhs = &ast.LogicalExpr{Op: "||", Lhs: lhs, Rhs: rhs}
	}
	return lhs
}

func (p *Parser) parseExpr() ast.Expr {
	lhs := p.parseCOr()
	//printTree(lhs, 0)
	return lhs
}

func (p *Parser) parseExprStmt() ast.Stmt {
	expr := p.parseExpr()
	return &ast.ExprStmt{Expr: expr}
}

func (p *Parser) parseDeclStmt() ast.Stmt {
	utils.Assert(p.curTokenIs(token.IDENT), "identifier needed")
	name := p.Tokens[p.Pos].Literal
	p.nextToken()
	p.nextToken()
	val := p.parseExpr()

	p.assignVal(name, val)
	decl := &ast.SVDecl{Name: name, Val: val}
	return &ast.DeclStmt{Decl: decl}
}

func (p *Parser) parseAssignStmt() ast.Stmt {
	name := p.Tokens[p.Pos].Literal
	if _, ok := p.VarMap[name]; !ok {
		fmt.Printf("Undeclared identifier: %s", name)
		os.Exit(1)
	}
	p.nextToken()
	p.nextToken()
	val := p.parseExpr()

	p.assignVal(name, val)

	return &ast.AssignStmt{Name: name, Val: val}
}

func (p *Parser) parseReturnStmt() ast.Stmt {
	p.nextToken()
	return &ast.ReturnStmt{Expr: p.parseExpr()}
}

func (p *Parser) parseBlockStmt() []ast.Stmt {
	p.expectToken(token.LBRACE)

	bs := []ast.Stmt{}
	for !p.curTokenIs(token.RBRACE) {
		bs = append(bs, p.parseStmt())
	}
	p.nextToken()
	return bs
}

func (p *Parser) parseIfStmt() ast.Stmt {
	p.nextToken()
	cond := p.parseExpr()
	cons := p.parseBlockStmt()
	var alt []ast.Stmt = nil
	if p.curTokenIs(token.ELSE) {
		p.nextToken()
		alt = p.parseBlockStmt()
	}
	return &ast.IfStmt{Cond: cond, Cons: cons, Alt: alt}
}

func (p *Parser) parseForSingleStmt() ast.Stmt {
	// skip the `for` token
	p.nextToken()
	cond := p.parseExpr()
	stmts := p.parseBlockStmt()
	return &ast.ForSingleStmt{Cond: cond, Stmts: stmts}
}

func (p *Parser) parseForStmt() ast.Stmt {
	// ForClauseとForRangeも追加する
	return p.parseForSingleStmt()
}

func (p *Parser) parseStmt() ast.Stmt {
	var stmt ast.Stmt

	if p.curTokenIs(token.IDENT) && p.peepTokenIs(token.SVDECL) {
		stmt = p.parseDeclStmt()
	} else if p.curTokenIs(token.IDENT) && p.peepTokenIs(token.ASSIGN) {
		stmt = p.parseAssignStmt()
	} else if p.curTokenIs(token.RETURN) {
		stmt = p.parseReturnStmt()
	} else if p.curTokenIs(token.IF) {
		stmt = p.parseIfStmt()
	} else if p.curTokenIs(token.FOR) {
		stmt = p.parseForStmt()
	} else {
		stmt = p.parseExprStmt()
	}

	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) Parse() []ast.Stmt {
	for !p.curTokenIs(token.EOF) {
		p.Stmts = append(p.Stmts, p.parseStmt())
	}
	return p.Stmts
}

func (p *Parser) curTokenIs(tt token.TokenType) bool {
	return tt == p.curToken().Type
}

func (p *Parser) peepTokenIs(tt token.TokenType) bool {
	return tt == p.peepToken().Type
}

func (p *Parser) curToken() token.Token {
	return p.Tokens[p.Pos]
}

func (p *Parser) peepToken() token.Token {
	return p.Tokens[p.Pos+1]
}

func (p *Parser) nextToken() {
	if p.curTokenIs(token.EOF) {
		return
	}
	p.Pos++
}

func (p *Parser) expectToken(tt token.TokenType) {
	if p.curTokenIs(tt) {
		p.nextToken()
		return
	}
	panic(fmt.Sprintf("expected %s, but got %s", tt, p.curToken().Type))
}

func printTree(node ast.Expr, tab int) {
	be, ok := node.(*ast.BinaryExpr)
	if ok {
		printTree(be.Lhs, tab+4)
		fmt.Println(strings.Repeat(" ", tab), be.Op)
		printTree(be.Rhs, tab+4)
		return
	}

	il, ok := node.(*ast.IntLit)
	if ok {
		fmt.Println(strings.Repeat(" ", tab), il.Val)
	}
	return
}

func (p *Parser) assignVal(name string, expr ast.Expr) {
	p.VarMap[name] = &ast.Ident{Name: name, Val: expr}
}
