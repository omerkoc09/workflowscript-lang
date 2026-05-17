package eval

import (
	"fmt"
	"workflowscript/internal/ast"
	"workflowscript/internal/token"
)

type returnSignal struct{ value interface{} }

type TaskValue struct {
	Timeout  int64
	Retries  int64
	Parallel bool
}

type Interpreter struct {
	env       *Env
	funs      map[string]*ast.FunDecl
	tasks     map[string]*TaskValue
	pipelines map[string]*ast.PipelineDecl
	runLog    []string // tracks executed task/stage names (for testing)
}

func New() *Interpreter {
	return &Interpreter{
		env:       NewEnv(nil),
		funs:      map[string]*ast.FunDecl{},
		tasks:     map[string]*TaskValue{},
		pipelines: map[string]*ast.PipelineDecl{},
	}
}

func (interp *Interpreter) Run(prog *ast.Program) error {
	// First pass: register tasks and functions
	for _, d := range prog.Decls {
		switch n := d.(type) {
		case *ast.TaskDecl:
			var timeout, retries int64
			fmt.Sscanf(n.Timeout.Lexeme, "%d", &timeout)
			fmt.Sscanf(n.Retries.Lexeme, "%d", &retries)
			tv := &TaskValue{
				Timeout:  timeout,
				Retries:  retries,
				Parallel: n.Parallel.Lexeme == "true",
			}
			interp.tasks[n.Name.Lexeme] = tv
			interp.env.Set(n.Name.Lexeme, tv)
		case *ast.PipelineDecl:
			interp.pipelines[n.Name.Lexeme] = n
		case *ast.FunDecl:
			interp.funs[n.Name.Lexeme] = n
		}
	}
	// Second pass: execute statements
	for _, d := range prog.Decls {
		switch d.(type) {
		case *ast.TaskDecl, *ast.FunDecl, *ast.PipelineDecl:
			continue
		}
		if _, err := interp.evalStmt(d, interp.env); err != nil {
			return err
		}
	}
	return nil
}

func (interp *Interpreter) evalStmt(s ast.Stmt, env *Env) (interface{}, error) {
	switch n := s.(type) {
	case *ast.VarDecl:
		val, err := interp.evalExpr(n.Value, env)
		if err != nil {
			return nil, err
		}
		env.Set(n.Name.Lexeme, val)
		return nil, nil
	case *ast.AssignStmt:
		val, err := interp.evalExpr(n.Value, env)
		if err != nil {
			return nil, err
		}
		if err := env.Assign(n.Name.Lexeme, val); err != nil {
			return nil, fmt.Errorf("line %d: %v", n.Name.Line, err)
		}
		return nil, nil
	case *ast.IfStmt:
		cond, err := interp.evalExpr(n.Condition, env)
		if err != nil {
			return nil, err
		}
		if cond.(bool) {
			return interp.evalBlock(n.Then, NewEnv(env))
		} else if n.Else != nil {
			return interp.evalBlock(n.Else, NewEnv(env))
		}
		return nil, nil
	case *ast.WhileStmt:
		for {
			cond, err := interp.evalExpr(n.Condition, env)
			if err != nil {
				return nil, err
			}
			if !cond.(bool) {
				break
			}
			ret, err := interp.evalBlock(n.Body, NewEnv(env))
			if err != nil {
				return nil, err
			}
			if ret != nil {
				return ret, nil
			}
		}
		return nil, nil
	case *ast.RunStmt:
		name := n.Target.Lexeme
		if _, ok := interp.tasks[name]; ok {
			fmt.Printf("[run] %s\n", name)
			interp.runLog = append(interp.runLog, name)
			return nil, nil
		}
		if pipeline, ok := interp.pipelines[name]; ok {
			for _, stage := range pipeline.Stages {
				fmt.Printf("[run] %s\n", stage.Lexeme)
				interp.runLog = append(interp.runLog, stage.Lexeme)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("line %d: undefined task or pipeline %q", n.Target.Line, name)
	case *ast.ReturnStmt:
		val, err := interp.evalExpr(n.Value, env)
		if err != nil {
			return nil, err
		}
		return &returnSignal{value: val}, nil
	case *ast.ExprStmt:
		_, err := interp.evalExpr(n.Expr, env)
		return nil, err
	case *ast.PipelineDecl, *ast.TaskDecl, *ast.FunDecl:
		return nil, nil
	}
	return nil, fmt.Errorf("unknown statement %T", s)
}

func (interp *Interpreter) evalBlock(b *ast.Block, env *Env) (interface{}, error) {
	for _, s := range b.Stmts {
		ret, err := interp.evalStmt(s, env)
		if err != nil {
			return nil, err
		}
		if ret != nil {
			return ret, nil
		}
	}
	return nil, nil
}

func (interp *Interpreter) evalExpr(e ast.Expr, env *Env) (interface{}, error) {
	switch n := e.(type) {
	case *ast.IntLit:
		return n.Value, nil
	case *ast.FloatLit:
		return n.Value, nil
	case *ast.BoolLit:
		return n.Value, nil
	case *ast.Ident:
		val, err := env.Get(n.Token.Lexeme)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", n.Token.Line, err)
		}
		return val, nil
	case *ast.FieldExpr:
		val, err := env.Get(n.Object.Lexeme)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", n.Object.Line, err)
		}
		tv, ok := val.(*TaskValue)
		if !ok {
			return nil, fmt.Errorf("line %d: %q is not a task", n.Object.Line, n.Object.Lexeme)
		}
		switch n.Field.Lexeme {
		case "timeout":
			return tv.Timeout, nil
		case "retries":
			return tv.Retries, nil
		case "parallel":
			return tv.Parallel, nil
		}
		return nil, fmt.Errorf("line %d: task has no field %q", n.Field.Line, n.Field.Lexeme)
	case *ast.CallExpr:
		return interp.evalCall(n, env)
	case *ast.BinaryExpr:
		if n.Op.Type == token.AND || n.Op.Type == token.OR {
			return interp.evalLogical(n, env)
		}
		return interp.evalBinary(n, env)
	case *ast.UnaryExpr:
		return interp.evalUnary(n, env)
	}
	return nil, fmt.Errorf("unknown expression %T", e)
}

func (interp *Interpreter) evalLogical(n *ast.BinaryExpr, env *Env) (interface{}, error) {
	left, err := interp.evalExpr(n.Left, env)
	if err != nil {
		return nil, err
	}
	if n.Op.Type == token.AND && !left.(bool) {
		return false, nil
	}
	if n.Op.Type == token.OR && left.(bool) {
		return true, nil
	}
	right, err := interp.evalExpr(n.Right, env)
	if err != nil {
		return nil, err
	}
	return right.(bool), nil
}

func (interp *Interpreter) evalCall(n *ast.CallExpr, env *Env) (interface{}, error) {
	if n.Name.Lexeme == "print" {
		val, err := interp.evalExpr(n.Args[0], env)
		if err != nil {
			return nil, err
		}
		fmt.Println(val)
		return nil, nil
	}
	fn, ok := interp.funs[n.Name.Lexeme]
	if !ok {
		return nil, fmt.Errorf("line %d: undeclared function %q", n.Name.Line, n.Name.Lexeme)
	}
	// Static scoping: close over global env, not caller's env
	callEnv := NewEnv(interp.env)
	for i, param := range fn.Params {
		val, err := interp.evalExpr(n.Args[i], env)
		if err != nil {
			return nil, err
		}
		callEnv.Set(param.Name.Lexeme, val)
	}
	ret, err := interp.evalBlock(fn.Body, callEnv)
	if err != nil {
		return nil, err
	}
	if sig, ok := ret.(*returnSignal); ok {
		return sig.value, nil
	}
	return nil, nil
}

func (interp *Interpreter) evalBinary(n *ast.BinaryExpr, env *Env) (interface{}, error) {
	left, err := interp.evalExpr(n.Left, env)
	if err != nil {
		return nil, err
	}
	right, err := interp.evalExpr(n.Right, env)
	if err != nil {
		return nil, err
	}

	li, lIsInt := left.(int64)
	ri, rIsInt := right.(int64)
	lf, _ := toFloat(left)
	rf, _ := toFloat(right)
	mixed := (lIsInt && !rIsInt) || (!lIsInt && rIsInt)

	switch n.Op.Type {
	case token.PLUS:
		if mixed || (!lIsInt && !rIsInt) {
			return lf + rf, nil
		}
		return li + ri, nil
	case token.MINUS:
		if mixed || (!lIsInt && !rIsInt) {
			return lf - rf, nil
		}
		return li - ri, nil
	case token.STAR:
		if mixed || (!lIsInt && !rIsInt) {
			return lf * rf, nil
		}
		return li * ri, nil
	case token.SLASH:
		if lIsInt && rIsInt {
			if ri == 0 {
				return nil, fmt.Errorf("line %d: division by zero", n.Op.Line)
			}
			return li / ri, nil
		}
		if rf == 0 {
			return nil, fmt.Errorf("line %d: division by zero", n.Op.Line)
		}
		return lf / rf, nil
	case token.LT:
		if mixed || (!lIsInt && !rIsInt) {
			return lf < rf, nil
		}
		return li < ri, nil
	case token.LTE:
		if mixed || (!lIsInt && !rIsInt) {
			return lf <= rf, nil
		}
		return li <= ri, nil
	case token.GT:
		if mixed || (!lIsInt && !rIsInt) {
			return lf > rf, nil
		}
		return li > ri, nil
	case token.GTE:
		if mixed || (!lIsInt && !rIsInt) {
			return lf >= rf, nil
		}
		return li >= ri, nil
	case token.EQ:
		return left == right, nil
	case token.NEQ:
		return left != right, nil
	}
	return nil, fmt.Errorf("unknown operator %q", n.Op.Lexeme)
}

func (interp *Interpreter) evalUnary(n *ast.UnaryExpr, env *Env) (interface{}, error) {
	val, err := interp.evalExpr(n.Operand, env)
	if err != nil {
		return nil, err
	}
	switch n.Op.Type {
	case token.MINUS:
		if i, ok := val.(int64); ok {
			return -i, nil
		}
		return -val.(float64), nil
	case token.BANG:
		return !val.(bool), nil
	}
	return nil, fmt.Errorf("unknown unary op %q", n.Op.Lexeme)
}

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int64:
		return float64(n), true
	}
	return 0, false
}
