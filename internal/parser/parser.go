package parser

import (
	"fmt"
	"strconv"
	"workflowscript/internal/ast"
	"workflowscript/internal/token"
)

type Parser struct {
	tokens []token.Token
	pos    int
}

func New(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) peek() token.Token {
	if p.pos >= len(p.tokens) {
		return token.Token{Type: token.EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() token.Token {
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *Parser) check(tt token.Type) bool {
	return p.peek().Type == tt
}

func (p *Parser) expect(tt token.Type) (token.Token, error) {
	tok := p.peek()
	if tok.Type != tt {
		return tok, fmt.Errorf("line %d: expected %s, got %q", tok.Line, tt, tok.Lexeme)
	}
	return p.advance(), nil
}

func (p *Parser) ParseProgram() (*ast.Program, error) {
	prog := &ast.Program{}
	for !p.check(token.EOF) {
		stmt, err := p.parseTopDecl()
		if err != nil {
			return nil, err
		}
		prog.Decls = append(prog.Decls, stmt)
	}
	return prog, nil
}

func (p *Parser) parseTopDecl() (ast.Stmt, error) {
	switch p.peek().Type {
	case token.TASK:
		return p.parseTaskDecl()
	case token.PIPELINE:
		return p.parsePipelineDecl()
	case token.FUN:
		return p.parseFunDecl()
	default:
		return p.parseStmt()
	}
}

func (p *Parser) parseTaskDecl() (*ast.TaskDecl, error) {
	p.advance() // consume "task"
	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.KW_TIMEOUT); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	timeout, err := p.expect(token.INT_LIT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.KW_RETRIES); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	retries, err := p.expect(token.INT_LIT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.KW_PARALLEL); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	parallel, err := p.expect(token.BOOL_LIT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}
	return &ast.TaskDecl{Name: name, Timeout: timeout, Retries: retries, Parallel: parallel}, nil
}

func (p *Parser) parsePipelineDecl() (*ast.PipelineDecl, error) {
	p.advance() // consume "pipeline"
	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}
	var stages []token.Token
	for p.check(token.STAGE) {
		p.advance()
		stage, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		stages = append(stages, stage)
	}
	if len(stages) == 0 {
		tok := p.peek()
		return nil, fmt.Errorf("line %d: pipeline must have at least one stage", tok.Line)
	}
	var onFailure []ast.Stmt
	if p.check(token.ON_FAILURE) {
		p.advance()
		if _, err := p.expect(token.LBRACE); err != nil {
			return nil, err
		}
		for !p.check(token.RBRACE) && !p.check(token.EOF) {
			s, err := p.parseStmt()
			if err != nil {
				return nil, err
			}
			onFailure = append(onFailure, s)
		}
		if _, err := p.expect(token.RBRACE); err != nil {
			return nil, err
		}
	}
	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}
	return &ast.PipelineDecl{Name: name, Stages: stages, OnFailure: onFailure}, nil
}

func (p *Parser) parseFunDecl() (*ast.FunDecl, error) {
	p.advance() // consume "fun"
	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}
	var params []ast.Param
	for !p.check(token.RPAREN) && !p.check(token.EOF) {
		pname, err := p.expect(token.IDENT)
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.COLON); err != nil {
			return nil, err
		}
		ptype, err := p.parseType()
		if err != nil {
			return nil, err
		}
		params = append(params, ast.Param{Name: pname, Type: ptype})
		if p.check(token.COMMA) {
			p.advance()
		}
	}
	if _, err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	if _, err := p.expect(token.ARROW); err != nil {
		return nil, err
	}
	retType, err := p.parseType()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.FunDecl{Name: name, Params: params, ReturnType: retType, Body: body}, nil
}

func (p *Parser) parseType() (token.Token, error) {
	tok := p.peek()
	switch tok.Type {
	case token.KW_INT, token.KW_FLOAT, token.KW_BOOL, token.KW_VOID, token.TASK:
		return p.advance(), nil
	default:
		return tok, fmt.Errorf("line %d: expected type, got %q", tok.Line, tok.Lexeme)
	}
}

func (p *Parser) parseBlock() (*ast.Block, error) {
	if _, err := p.expect(token.LBRACE); err != nil {
		return nil, err
	}
	var stmts []ast.Stmt
	for !p.check(token.RBRACE) && !p.check(token.EOF) {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
	}
	if _, err := p.expect(token.RBRACE); err != nil {
		return nil, err
	}
	return &ast.Block{Stmts: stmts}, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	switch p.peek().Type {
	case token.VAR:
		return p.parseVarDecl()
	case token.IF:
		return p.parseIfStmt()
	case token.WHILE:
		return p.parseWhileStmt()
	case token.RETURN:
		return p.parseReturnStmt()
	case token.RUN:
		return p.parseRunStmt()
	case token.IDENT:
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == token.ASSIGN {
			return p.parseAssignStmt()
		}
		return p.parseExprStmt()
	default:
		tok := p.peek()
		return nil, fmt.Errorf("line %d: unexpected token %q in statement", tok.Line, tok.Lexeme)
	}
}

func (p *Parser) parseVarDecl() (*ast.VarDecl, error) {
	p.advance() // consume "var"
	name, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.COLON); err != nil {
		return nil, err
	}
	typ, err := p.parseType()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(token.ASSIGN); err != nil {
		return nil, err
	}
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.VarDecl{Name: name, Type: typ, Value: val}, nil
}

func (p *Parser) parseAssignStmt() (*ast.AssignStmt, error) {
	name := p.advance()
	p.advance() // consume "="
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.AssignStmt{Name: name, Value: val}, nil
}

func (p *Parser) parseIfStmt() (*ast.IfStmt, error) {
	p.advance() // consume "if"
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	var elseBranch *ast.Block
	if p.check(token.ELSE) {
		p.advance()
		elseBranch, err = p.parseBlock()
		if err != nil {
			return nil, err
		}
	}
	return &ast.IfStmt{Condition: cond, Then: then, Else: elseBranch}, nil
}

func (p *Parser) parseWhileStmt() (*ast.WhileStmt, error) {
	p.advance() // consume "while"
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &ast.WhileStmt{Condition: cond, Body: body}, nil
}

func (p *Parser) parseReturnStmt() (*ast.ReturnStmt, error) {
	kw := p.advance() // consume "return"
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ast.ReturnStmt{Keyword: kw, Value: val}, nil
}

func (p *Parser) parseRunStmt() (*ast.RunStmt, error) {
	p.advance() // consume "run"
	target, err := p.expect(token.IDENT)
	if err != nil {
		return nil, err
	}
	return &ast.RunStmt{Target: target}, nil
}

func (p *Parser) parseExprStmt() (*ast.ExprStmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil, fmt.Errorf("line %d: expression statement must be a function call", p.peek().Line)
	}
	return &ast.ExprStmt{Expr: call}, nil
}

func (p *Parser) parseExpr() (ast.Expr, error)  { return p.parseOr() }

func (p *Parser) parseOr() (ast.Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.check(token.OR) {
		op := p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAnd() (ast.Expr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.check(token.AND) {
		op := p.advance()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseEquality() (ast.Expr, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.check(token.EQ) || p.check(token.NEQ) {
		op := p.advance()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseComparison() (ast.Expr, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}
	for p.check(token.LT) || p.check(token.LTE) || p.check(token.GT) || p.check(token.GTE) {
		op := p.advance()
		right, err := p.parseAddSub()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAddSub() (ast.Expr, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}
	for p.check(token.PLUS) || p.check(token.MINUS) {
		op := p.advance()
		right, err := p.parseMulDiv()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseMulDiv() (ast.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.check(token.STAR) || p.check(token.SLASH) {
		op := p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (ast.Expr, error) {
	if p.check(token.MINUS) || p.check(token.BANG) {
		op := p.advance()
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Op: op, Operand: operand}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	tok := p.peek()
	switch tok.Type {
	case token.INT_LIT:
		p.advance()
		val, _ := strconv.ParseInt(tok.Lexeme, 10, 64)
		return &ast.IntLit{Token: tok, Value: val}, nil
	case token.FLOAT_LIT:
		p.advance()
		val, _ := strconv.ParseFloat(tok.Lexeme, 64)
		return &ast.FloatLit{Token: tok, Value: val}, nil
	case token.BOOL_LIT:
		p.advance()
		return &ast.BoolLit{Token: tok, Value: tok.Lexeme == "true"}, nil
	case token.LPAREN:
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(token.RPAREN); err != nil {
			return nil, err
		}
		return expr, nil
	case token.IDENT:
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == token.LPAREN {
			return p.parseCallExpr()
		}
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == token.DOT {
			obj := p.advance()
			p.advance() // consume "."
			field := p.peek()
			// Allow keywords as field names (e.g., obj.timeout)
			if field.Type != token.IDENT && field.Type != token.KW_TIMEOUT && field.Type != token.KW_RETRIES && field.Type != token.KW_PARALLEL {
				return nil, fmt.Errorf("line %d: expected field name, got %q", field.Line, field.Lexeme)
			}
			p.advance()
			return &ast.FieldExpr{Object: obj, Field: field}, nil
		}
		return &ast.Ident{Token: p.advance()}, nil
	default:
		return nil, fmt.Errorf("line %d: unexpected token %q in expression", tok.Line, tok.Lexeme)
	}
}

func (p *Parser) parseCallExpr() (*ast.CallExpr, error) {
	name := p.advance()
	p.advance() // consume "("
	var args []ast.Expr
	for !p.check(token.RPAREN) && !p.check(token.EOF) {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.check(token.COMMA) {
			p.advance()
		}
	}
	if _, err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	return &ast.CallExpr{Name: name, Args: args}, nil
}
