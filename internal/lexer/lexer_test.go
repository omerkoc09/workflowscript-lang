package lexer

import (
	"testing"
	"workflowscript/internal/token"
)

func TestLexer_BasicTokens(t *testing.T) {
	src := `task build { timeout: 30 retries: 0 parallel: false }`
	l := New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []token.Type{
		token.TASK, token.IDENT, token.LBRACE,
		token.KW_TIMEOUT, token.COLON, token.INT_LIT,
		token.KW_RETRIES, token.COLON, token.INT_LIT,
		token.KW_PARALLEL, token.COLON, token.BOOL_LIT,
		token.RBRACE, token.EOF,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tt := range expected {
		if tokens[i].Type != tt {
			t.Errorf("token[%d]: expected %s, got %s (%q)", i, tt, tokens[i].Type, tokens[i].Lexeme)
		}
	}
}

func TestLexer_FloatLiteral(t *testing.T) {
	src := `3.14`
	l := New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens[0].Type != token.FLOAT_LIT || tokens[0].Lexeme != "3.14" {
		t.Errorf("expected FLOAT_LIT 3.14, got %s %q", tokens[0].Type, tokens[0].Lexeme)
	}
}

func TestLexer_LineNumbers(t *testing.T) {
	src := "task build {\n  timeout: 30\n}"
	l := New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tok := range tokens {
		if tok.Type == token.KW_TIMEOUT && tok.Line != 2 {
			t.Errorf("timeout should be on line 2, got line %d", tok.Line)
		}
	}
}

func TestLexer_IllegalChar(t *testing.T) {
	src := `task @ build`
	l := New(src)
	_, err := l.Tokenize()
	if err == nil {
		t.Error("expected error for illegal character @")
	}
}

func TestLexer_Arrow(t *testing.T) {
	src := `-> -`
	l := New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens[0].Type != token.ARROW {
		t.Errorf("expected ARROW, got %s", tokens[0].Type)
	}
	if tokens[1].Type != token.MINUS {
		t.Errorf("expected MINUS, got %s", tokens[1].Type)
	}
}

func TestLexer_Comment(t *testing.T) {
	src := "task // this is a comment\nbuild"
	l := New(src)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens[0].Type != token.TASK {
		t.Errorf("expected TASK, got %s", tokens[0].Type)
	}
	if tokens[1].Type != token.IDENT || tokens[1].Lexeme != "build" {
		t.Errorf("expected IDENT build, got %s %q", tokens[1].Type, tokens[1].Lexeme)
	}
}
