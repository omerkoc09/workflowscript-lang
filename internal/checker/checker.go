package checker

import (
	"fmt"
	"workflowscript/internal/ast"
	"workflowscript/internal/token"
)

type Checker struct {
	scopes     []map[string]Type
	funs       map[string]*ast.FunDecl
	tasks      map[string]*ast.TaskDecl
	pipelines  map[string]*ast.PipelineDecl
	returnType Type // expected return type of current function
}

func New() *Checker {
	return &Checker{
		scopes:    []map[string]Type{{}},
		funs:      map[string]*ast.FunDecl{},
		tasks:     map[string]*ast.TaskDecl{},
		pipelines: map[string]*ast.PipelineDecl{},
	}
}

func (c *Checker) define(name string, t Type) {
	c.scopes[len(c.scopes)-1][name] = t
}

func (c *Checker) lookup(name string) (Type, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if t, ok := c.scopes[i][name]; ok {
			return t, true
		}
	}
	return nil, false
}

func (c *Checker) push() { c.scopes = append(c.scopes, map[string]Type{}) }
func (c *Checker) pop()  { c.scopes = c.scopes[:len(c.scopes)-1] }

func (c *Checker) Check(prog *ast.Program) error {
	// First pass: register tasks, pipelines, and functions for forward references
	for _, d := range prog.Decls {
		switch n := d.(type) {
		case *ast.TaskDecl:
			c.tasks[n.Name.Lexeme] = n
			c.define(n.Name.Lexeme, TTask)
		case *ast.PipelineDecl:
			c.pipelines[n.Name.Lexeme] = n
		case *ast.FunDecl:
			c.funs[n.Name.Lexeme] = n
		}
	}
	for _, d := range prog.Decls {
		if err := c.checkStmt(d); err != nil {
			return err
		}
	}
	return nil
}

func (c *Checker) checkStmt(s ast.Stmt) error {
	switch n := s.(type) {
	case *ast.TaskDecl:
		return nil
	case *ast.PipelineDecl:
		for _, stage := range n.Stages {
			if _, ok := c.tasks[stage.Lexeme]; !ok {
				return fmt.Errorf("line %d: unknown task %q in pipeline", stage.Line, stage.Lexeme)
			}
		}
		for _, stmt := range n.OnFailure {
			if err := c.checkStmt(stmt); err != nil {
				return err
			}
		}
		return nil
	case *ast.FunDecl:
		c.push()
		for _, p := range n.Params {
			t, err := typeFromLexeme(p.Type.Lexeme)
			if err != nil {
				return fmt.Errorf("line %d: %v", p.Type.Line, err)
			}
			c.define(p.Name.Lexeme, t)
		}
		retType, err := typeFromLexeme(n.ReturnType.Lexeme)
		if err != nil {
			return fmt.Errorf("line %d: %v", n.ReturnType.Line, err)
		}
		prev := c.returnType
		c.returnType = retType
		if err := c.checkBlock(n.Body); err != nil {
			c.returnType = prev
			c.pop()
			return err
		}
		c.returnType = prev
		c.pop()
		if _, isVoid := retType.(VoidType); !isVoid {
			if !blockAlwaysReturns(n.Body) {
				return fmt.Errorf("line %d: function %q does not return on all paths",
					n.Name.Line, n.Name.Lexeme)
			}
		}
		return nil
	case *ast.VarDecl:
		declared, err := typeFromLexeme(n.Type.Lexeme)
		if err != nil {
			return fmt.Errorf("line %d: %v", n.Type.Line, err)
		}
		if _, ok := declared.(VoidType); ok {
			return fmt.Errorf("line %d: cannot declare variable of type void", n.Name.Line)
		}
		actual, err := c.checkExpr(n.Value)
		if err != nil {
			return err
		}
		if !typesCompatible(actual, declared) {
			return fmt.Errorf("line %d: cannot assign %s to variable of type %s",
				n.Name.Line, actual.typeString(), declared.typeString())
		}
		c.define(n.Name.Lexeme, declared)
		return nil
	case *ast.AssignStmt:
		declared, ok := c.lookup(n.Name.Lexeme)
		if !ok {
			return fmt.Errorf("line %d: undeclared variable %q", n.Name.Line, n.Name.Lexeme)
		}
		actual, err := c.checkExpr(n.Value)
		if err != nil {
			return err
		}
		if !typesCompatible(actual, declared) {
			return fmt.Errorf("line %d: cannot assign %s to %s",
				n.Name.Line, actual.typeString(), declared.typeString())
		}
		return nil
	case *ast.IfStmt:
		cond, err := c.checkExpr(n.Condition)
		if err != nil {
			return err
		}
		if _, ok := cond.(BoolType); !ok {
			return fmt.Errorf("if condition must be bool, got %s", cond.typeString())
		}
		c.push()
		if err := c.checkBlock(n.Then); err != nil {
			c.pop()
			return err
		}
		c.pop()
		if n.Else != nil {
			c.push()
			if err := c.checkBlock(n.Else); err != nil {
				c.pop()
				return err
			}
			c.pop()
		}
		return nil
	case *ast.WhileStmt:
		cond, err := c.checkExpr(n.Condition)
		if err != nil {
			return err
		}
		if _, ok := cond.(BoolType); !ok {
			return fmt.Errorf("while condition must be bool, got %s", cond.typeString())
		}
		c.push()
		if err := c.checkBlock(n.Body); err != nil {
			c.pop()
			return err
		}
		c.pop()
		return nil
	case *ast.RunStmt:
		name := n.Target.Lexeme
		_, isTask := c.tasks[name]
		_, isPipeline := c.pipelines[name]
		if !isTask && !isPipeline {
			return fmt.Errorf("line %d: undefined task or pipeline %q", n.Target.Line, name)
		}
		return nil
	case *ast.ReturnStmt:
		actual, err := c.checkExpr(n.Value)
		if err != nil {
			return err
		}
		if c.returnType != nil && !typesCompatible(actual, c.returnType) {
			return fmt.Errorf("line %d: cannot return %s from function declared to return %s",
				n.Keyword.Line, actual.typeString(), c.returnType.typeString())
		}
		return nil
	case *ast.ExprStmt:
		t, err := c.checkExpr(n.Expr)
		if err != nil {
			return err
		}
		_ = t
		return nil
	}
	return nil
}

func (c *Checker) checkBlock(b *ast.Block) error {
	for _, s := range b.Stmts {
		if err := c.checkStmt(s); err != nil {
			return err
		}
	}
	return nil
}

func (c *Checker) checkExpr(e ast.Expr) (Type, error) {
	switch n := e.(type) {
	case *ast.IntLit:
		return TInt, nil
	case *ast.FloatLit:
		return TFloat, nil
	case *ast.BoolLit:
		return TBool, nil
	case *ast.Ident:
		t, ok := c.lookup(n.Token.Lexeme)
		if !ok {
			return nil, fmt.Errorf("line %d: undeclared variable %q", n.Token.Line, n.Token.Lexeme)
		}
		return t, nil
	case *ast.FieldExpr:
		t, ok := c.lookup(n.Object.Lexeme)
		if !ok {
			return nil, fmt.Errorf("line %d: undeclared variable %q", n.Object.Line, n.Object.Lexeme)
		}
		if _, ok := t.(TaskType); !ok {
			return nil, fmt.Errorf("line %d: %q is not a task", n.Object.Line, n.Object.Lexeme)
		}
		switch n.Field.Lexeme {
		case "timeout", "retries":
			return TInt, nil
		case "parallel":
			return TBool, nil
		default:
			return nil, fmt.Errorf("line %d: task has no field %q", n.Field.Line, n.Field.Lexeme)
		}
	case *ast.CallExpr:
		return c.checkCall(n)
	case *ast.BinaryExpr:
		return c.checkBinary(n)
	case *ast.UnaryExpr:
		return c.checkUnary(n)
	}
	return nil, fmt.Errorf("unknown expression type %T", e)
}

func (c *Checker) checkCall(n *ast.CallExpr) (Type, error) {
	if n.Name.Lexeme == "print" {
		if len(n.Args) != 1 {
			return nil, fmt.Errorf("line %d: print expects 1 argument, got %d", n.Name.Line, len(n.Args))
		}
		if _, err := c.checkExpr(n.Args[0]); err != nil {
			return nil, err
		}
		return TVoid, nil
	}
	fn, ok := c.funs[n.Name.Lexeme]
	if !ok {
		return nil, fmt.Errorf("line %d: undeclared function %q", n.Name.Line, n.Name.Lexeme)
	}
	if len(n.Args) != len(fn.Params) {
		return nil, fmt.Errorf("line %d: %q expects %d args, got %d",
			n.Name.Line, n.Name.Lexeme, len(fn.Params), len(n.Args))
	}
	for i, arg := range n.Args {
		argType, err := c.checkExpr(arg)
		if err != nil {
			return nil, err
		}
		paramType, _ := typeFromLexeme(fn.Params[i].Type.Lexeme)
		if !typesCompatible(argType, paramType) {
			return nil, fmt.Errorf("line %d: arg %d: expected %s, got %s",
				n.Name.Line, i+1, paramType.typeString(), argType.typeString())
		}
	}
	retType, err := typeFromLexeme(fn.ReturnType.Lexeme)
	if err != nil {
		return nil, err
	}
	if _, isVoid := retType.(VoidType); isVoid {
		return TVoid, nil
	}
	return retType, nil
}

func (c *Checker) checkBinary(n *ast.BinaryExpr) (Type, error) {
	left, err := c.checkExpr(n.Left)
	if err != nil {
		return nil, err
	}
	right, err := c.checkExpr(n.Right)
	if err != nil {
		return nil, err
	}
	switch n.Op.Type {
	case token.AND, token.OR:
		if _, ok := left.(BoolType); !ok {
			return nil, fmt.Errorf("line %d: %s requires bool operands", n.Op.Line, n.Op.Lexeme)
		}
		if _, ok := right.(BoolType); !ok {
			return nil, fmt.Errorf("line %d: %s requires bool operands", n.Op.Line, n.Op.Lexeme)
		}
		return TBool, nil
	case token.EQ, token.NEQ:
		if _, ok := left.(VoidType); ok {
			return nil, fmt.Errorf("line %d: void is not comparable", n.Op.Line)
		}
		if _, ok := right.(VoidType); ok {
			return nil, fmt.Errorf("line %d: void is not comparable", n.Op.Line)
		}
		if !typesCompatible(left, right) && !typesCompatible(right, left) {
			return nil, fmt.Errorf("line %d: cannot compare %s with %s",
				n.Op.Line, left.typeString(), right.typeString())
		}
		return TBool, nil
	case token.LT, token.LTE, token.GT, token.GTE:
		if !isNumeric(left) || !isNumeric(right) {
			return nil, fmt.Errorf("line %d: %s requires numeric operands", n.Op.Line, n.Op.Lexeme)
		}
		return TBool, nil
	case token.PLUS, token.MINUS, token.STAR, token.SLASH:
		if !isNumeric(left) || !isNumeric(right) {
			return nil, fmt.Errorf("line %d: %s requires numeric operands", n.Op.Line, n.Op.Lexeme)
		}
		if _, ok := left.(FloatType); ok {
			return TFloat, nil
		}
		if _, ok := right.(FloatType); ok {
			return TFloat, nil
		}
		return TInt, nil
	}
	return nil, fmt.Errorf("line %d: unknown operator %q", n.Op.Line, n.Op.Lexeme)
}

func (c *Checker) checkUnary(n *ast.UnaryExpr) (Type, error) {
	t, err := c.checkExpr(n.Operand)
	if err != nil {
		return nil, err
	}
	switch n.Op.Type {
	case token.MINUS:
		if !isNumeric(t) {
			return nil, fmt.Errorf("line %d: unary - requires numeric operand", n.Op.Line)
		}
		return t, nil
	case token.BANG:
		if _, ok := t.(BoolType); !ok {
			return nil, fmt.Errorf("line %d: ! requires bool operand", n.Op.Line)
		}
		return TBool, nil
	}
	return nil, fmt.Errorf("unknown unary op %q", n.Op.Lexeme)
}

func blockAlwaysReturns(b *ast.Block) bool {
	for _, s := range b.Stmts {
		if stmtAlwaysReturns(s) {
			return true
		}
	}
	return false
}

func stmtAlwaysReturns(s ast.Stmt) bool {
	switch n := s.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.IfStmt:
		return n.Else != nil && blockAlwaysReturns(n.Then) && blockAlwaysReturns(n.Else)
	}
	return false
}

func isNumeric(t Type) bool {
	switch t.(type) {
	case IntType, FloatType:
		return true
	}
	return false
}
