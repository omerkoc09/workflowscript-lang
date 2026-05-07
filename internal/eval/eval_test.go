package eval

import (
	"testing"
	"workflowscript/internal/lexer"
	"workflowscript/internal/parser"
)

func run(t *testing.T, src string) *Interpreter {
	t.Helper()
	l := lexer.New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("lexer: %v", err)
	}
	p := parser.New(tokens)
	prog, err := p.ParseProgram()
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	interp := New()
	if err := interp.Run(prog); err != nil {
		t.Fatalf("eval: %v", err)
	}
	return interp
}

func TestEval_VarDecl(t *testing.T) {
	interp := run(t, `var x: int = 3 + 4`)
	v, err := interp.env.Get("x")
	if err != nil {
		t.Fatal(err)
	}
	if v.(int64) != 7 {
		t.Errorf("expected 7, got %v", v)
	}
}

func TestEval_WhileLoop(t *testing.T) {
	src := `
var i: int = 0
while i < 3 {
    i = i + 1
}
`
	interp := run(t, src)
	v, _ := interp.env.Get("i")
	if v.(int64) != 3 {
		t.Errorf("expected 3, got %v", v)
	}
}

func TestEval_FunctionCall(t *testing.T) {
	src := `
fun double(x: int) -> int { return x * 2 }
var result: int = double(21)
`
	interp := run(t, src)
	v, _ := interp.env.Get("result")
	if v.(int64) != 42 {
		t.Errorf("expected 42, got %v", v)
	}
}

func TestEval_TaskFieldAccess(t *testing.T) {
	src := `
task build { timeout: 30 retries: 0 parallel: false }
var t: int = build.timeout
`
	interp := run(t, src)
	v, _ := interp.env.Get("t")
	if v.(int64) != 30 {
		t.Errorf("expected 30, got %v", v)
	}
}

func TestEval_DivisionByZero(t *testing.T) {
	src := `var x: int = 10 / 0`
	l := lexer.New(src)
	tokens, _ := l.Tokenize()
	p := parser.New(tokens)
	prog, _ := p.ParseProgram()
	interp := New()
	err := interp.Run(prog)
	if err == nil {
		t.Error("expected division by zero error")
	}
}

func TestEval_StaticScoping(t *testing.T) {
	src := `
var x: int = 10
fun get_x() -> int { return x }
var y: int = get_x()
`
	interp := run(t, src)
	v, _ := interp.env.Get("y")
	if v.(int64) != 10 {
		t.Errorf("expected 10 (static scoping), got %v", v)
	}
}
