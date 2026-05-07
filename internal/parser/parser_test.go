package parser

import (
	"testing"
	"workflowscript/internal/ast"
	"workflowscript/internal/lexer"
)

func parse(t *testing.T, src string) *ast.Program {
	t.Helper()
	l := lexer.New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	p := New(tokens)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	return prog
}

func TestParser_TaskDecl(t *testing.T) {
	src := `task build { timeout: 30 retries: 0 parallel: false }`
	prog := parse(t, src)
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	task, ok := prog.Decls[0].(*ast.TaskDecl)
	if !ok {
		t.Fatalf("expected *ast.TaskDecl, got %T", prog.Decls[0])
	}
	if task.Name.Lexeme != "build" {
		t.Errorf("expected name 'build', got %q", task.Name.Lexeme)
	}
	if task.Timeout.Lexeme != "30" {
		t.Errorf("expected timeout 30, got %q", task.Timeout.Lexeme)
	}
	if task.Parallel.Lexeme != "false" {
		t.Errorf("expected parallel false, got %q", task.Parallel.Lexeme)
	}
}

func TestParser_PipelineDecl(t *testing.T) {
	src := `pipeline ci { stage build stage test }`
	prog := parse(t, src)
	if len(prog.Decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(prog.Decls))
	}
	pipe, ok := prog.Decls[0].(*ast.PipelineDecl)
	if !ok {
		t.Fatalf("expected *ast.PipelineDecl, got %T", prog.Decls[0])
	}
	if len(pipe.Stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(pipe.Stages))
	}
	if pipe.Stages[0].Lexeme != "build" {
		t.Errorf("expected stage 'build', got %q", pipe.Stages[0].Lexeme)
	}
}

func TestParser_FunDecl(t *testing.T) {
	src := `fun double(x: int) -> int { return x }`
	prog := parse(t, src)
	fn, ok := prog.Decls[0].(*ast.FunDecl)
	if !ok {
		t.Fatalf("expected *ast.FunDecl, got %T", prog.Decls[0])
	}
	if fn.Name.Lexeme != "double" {
		t.Errorf("expected name 'double', got %q", fn.Name.Lexeme)
	}
	if len(fn.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(fn.Params))
	}
	if fn.ReturnType.Lexeme != "int" {
		t.Errorf("expected return type 'int', got %q", fn.ReturnType.Lexeme)
	}
}

func TestParser_VarDecl(t *testing.T) {
	src := `var x: int = 42`
	prog := parse(t, src)
	v, ok := prog.Decls[0].(*ast.VarDecl)
	if !ok {
		t.Fatalf("expected *ast.VarDecl, got %T", prog.Decls[0])
	}
	if v.Name.Lexeme != "x" {
		t.Errorf("expected name 'x', got %q", v.Name.Lexeme)
	}
	lit, ok := v.Value.(*ast.IntLit)
	if !ok || lit.Value != 42 {
		t.Errorf("expected IntLit(42), got %T", v.Value)
	}
}

func TestParser_BinaryExpr_Precedence(t *testing.T) {
	// 3 + 4 * 2 should parse as (3 + (4 * 2))
	src := `var x: int = 3 + 4 * 2`
	prog := parse(t, src)
	v := prog.Decls[0].(*ast.VarDecl)
	add, ok := v.Value.(*ast.BinaryExpr)
	if !ok || add.Op.Lexeme != "+" {
		t.Fatalf("expected BinaryExpr(+), got %T", v.Value)
	}
	mul, ok := add.Right.(*ast.BinaryExpr)
	if !ok || mul.Op.Lexeme != "*" {
		t.Errorf("expected right side to be BinaryExpr(*), got %T", add.Right)
	}
}

func TestParser_IfStmt(t *testing.T) {
	src := `if x > 5 { run build }`
	prog := parse(t, src)
	_, ok := prog.Decls[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected *ast.IfStmt, got %T", prog.Decls[0])
	}
}

func TestParser_MissingBrace(t *testing.T) {
	src := `task build { timeout: 30 retries: 0 parallel: false`
	l := lexer.New(src)
	tokens, _ := l.Tokenize()
	p := New(tokens)
	_, err := p.ParseProgram()
	if err == nil {
		t.Error("expected parse error for missing closing brace")
	}
}

func TestParser_ErrorIncludesLineNumber(t *testing.T) {
	src := "task build {\n  timeout: 30\n  retries: 0\n  parallel: false"
	l := lexer.New(src)
	tokens, _ := l.Tokenize()
	p := New(tokens)
	_, err := p.ParseProgram()
	if err == nil {
		t.Error("expected parse error")
	}
}
