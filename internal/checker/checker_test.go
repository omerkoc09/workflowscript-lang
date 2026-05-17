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

func TestChecker_ReturnTypeMismatch(t *testing.T) {
	err := check(t, `fun f() -> int { return true }`)
	if err == nil {
		t.Error("expected type error: returning bool from int function")
	}
}

func TestChecker_VoidFunctionReturnValue(t *testing.T) {
	src := `
task build { timeout: 10 retries: 0 parallel: false }
fun notify() -> void { return 42 }
`
	if err := check(t, src); err == nil {
		t.Error("expected type error: returning value from void function")
	}
}

func TestChecker_RunUndefinedTarget(t *testing.T) {
	if err := check(t, `run ghost`); err == nil {
		t.Error("expected compile-time error for undefined task or pipeline 'ghost'")
	}
}

func TestChecker_PrintBuiltin(t *testing.T) {
	if err := check(t, `print(42)`); err != nil {
		t.Errorf("print should be recognized as a built-in: %v", err)
	}
}

func TestChecker_ExhaustiveReturn(t *testing.T) {
	// if without else — false branch falls off the end
	err := check(t, `fun f(x: int) -> int { if x > 0 { return x } }`)
	if err == nil {
		t.Error("expected error: function does not return on all paths")
	}
}

func TestChecker_ExhaustiveReturnIfElse(t *testing.T) {
	// if + else, both branches return — must be accepted
	src := `fun f(x: int) -> int { if x > 0 { return x } else { return 0 } }`
	if err := check(t, src); err != nil {
		t.Errorf("if/else with both branches returning should be valid: %v", err)
	}
}

func TestChecker_VoidVariable(t *testing.T) {
	src := `
task build { timeout: 10 retries: 0 parallel: false }
fun notify() -> void { run build }
var x: void = notify()
`
	if err := check(t, src); err == nil {
		t.Error("expected error: cannot declare variable of type void")
	}
}

func TestChecker_VoidComparison(t *testing.T) {
	src := `
task build { timeout: 10 retries: 0 parallel: false }
fun notify() -> void { run build }
var eq: bool = notify() == notify()
`
	if err := check(t, src); err == nil {
		t.Error("expected error: void is not comparable")
	}
}
