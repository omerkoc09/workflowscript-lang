package checker

import (
	"testing"
	"workflowscript/internal/lexer"
	"workflowscript/internal/parser"
)

func check(t *testing.T, src string) error {
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
	c := New()
	return c.Check(prog)
}

func TestChecker_ValidVarDecl(t *testing.T) {
	if err := check(t, `var x: int = 42`); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChecker_TypeMismatch(t *testing.T) {
	err := check(t, `var x: int = true`)
	if err == nil {
		t.Error("expected type mismatch error")
	}
}

func TestChecker_IntToFloatCoercion(t *testing.T) {
	if err := check(t, `var x: float = 1 + 2.5`); err != nil {
		t.Errorf("unexpected error for int+float: %v", err)
	}
}

func TestChecker_UndeclaredVariable(t *testing.T) {
	err := check(t, `var x: int = y + 1`)
	if err == nil {
		t.Error("expected undeclared variable error")
	}
}

func TestChecker_VoidReturnAsValue(t *testing.T) {
	src := `
task build { timeout: 10 retries: 0 parallel: false }
fun notify() -> void { run build }
var x: int = notify()
`
	if err := check(t, src); err == nil {
		t.Error("expected error: cannot use void as value")
	}
}

func TestChecker_UnknownTaskInPipeline(t *testing.T) {
	src := `pipeline ci { stage ghost }`
	if err := check(t, src); err == nil {
		t.Error("expected error for unknown task in pipeline")
	}
}

func TestChecker_TaskFieldAccess(t *testing.T) {
	src := `
task build { timeout: 30 retries: 0 parallel: false }
var t: int = build.timeout
`
	if err := check(t, src); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChecker_InvalidTaskField(t *testing.T) {
	src := `
task build { timeout: 30 retries: 0 parallel: false }
var t: int = build.nonexistent
`
	if err := check(t, src); err == nil {
		t.Error("expected error for invalid task field")
	}
}
