package ast

import (
	"fmt"
	"strings"
	"workflowscript/internal/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Stmt interface {
	Node
	stmtNode()
}

type Expr interface {
	Node
	exprNode()
}

// Program is the root node
type Program struct {
	Decls []Stmt
}

func (p *Program) TokenLiteral() string { return "" }
func (p *Program) String() string {
	var sb strings.Builder
	for _, d := range p.Decls {
		sb.WriteString(d.String())
		sb.WriteString("\n")
	}
	return sb.String()
}

// TaskDecl: task build { timeout: 30 retries: 0 parallel: false }
type TaskDecl struct {
	Name     token.Token
	Timeout  token.Token // INT_LIT
	Retries  token.Token // INT_LIT
	Parallel token.Token // BOOL_LIT
}

func (t *TaskDecl) TokenLiteral() string { return t.Name.Lexeme }
func (t *TaskDecl) String() string {
	return fmt.Sprintf("task %s { timeout: %s retries: %s parallel: %s }",
		t.Name.Lexeme, t.Timeout.Lexeme, t.Retries.Lexeme, t.Parallel.Lexeme)
}
func (t *TaskDecl) stmtNode() {}

// PipelineDecl: pipeline ci { stage build stage test on_failure { ... } }
type PipelineDecl struct {
	Name      token.Token
	Stages    []token.Token
	OnFailure []Stmt // nil if no on_failure
}

func (p *PipelineDecl) TokenLiteral() string { return p.Name.Lexeme }
func (p *PipelineDecl) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("pipeline %s {\n", p.Name.Lexeme))
	for _, s := range p.Stages {
		sb.WriteString(fmt.Sprintf("  stage %s\n", s.Lexeme))
	}
	if len(p.OnFailure) > 0 {
		sb.WriteString("  on_failure {\n")
		for _, s := range p.OnFailure {
			sb.WriteString(fmt.Sprintf("    %s\n", s.String()))
		}
		sb.WriteString("  }\n")
	}
	sb.WriteString("}")
	return sb.String()
}
func (p *PipelineDecl) stmtNode() {}

// FunDecl: fun name(p: type) -> type { body }
type FunDecl struct {
	Name       token.Token
	Params     []Param
	ReturnType token.Token
	Body       *Block
}

type Param struct {
	Name token.Token
	Type token.Token
}

func (f *FunDecl) TokenLiteral() string { return f.Name.Lexeme }
func (f *FunDecl) String() string {
	var params []string
	for _, p := range f.Params {
		params = append(params, fmt.Sprintf("%s: %s", p.Name.Lexeme, p.Type.Lexeme))
	}
	return fmt.Sprintf("fun %s(%s) -> %s %s",
		f.Name.Lexeme, strings.Join(params, ", "), f.ReturnType.Lexeme, f.Body.String())
}
func (f *FunDecl) stmtNode() {}

// Block: { stmt* }
type Block struct {
	Stmts []Stmt
}

func (b *Block) TokenLiteral() string { return "{" }
func (b *Block) String() string {
	var sb strings.Builder
	sb.WriteString("{\n")
	for _, s := range b.Stmts {
		sb.WriteString("  " + s.String() + "\n")
	}
	sb.WriteString("}")
	return sb.String()
}

// VarDecl: var x: int = expr
type VarDecl struct {
	Name  token.Token
	Type  token.Token
	Value Expr
}

func (v *VarDecl) TokenLiteral() string { return v.Name.Lexeme }
func (v *VarDecl) String() string {
	return fmt.Sprintf("var %s: %s = %s", v.Name.Lexeme, v.Type.Lexeme, v.Value.String())
}
func (v *VarDecl) stmtNode() {}

// AssignStmt: x = expr
type AssignStmt struct {
	Name  token.Token
	Value Expr
}

func (a *AssignStmt) TokenLiteral() string { return a.Name.Lexeme }
func (a *AssignStmt) String() string {
	return fmt.Sprintf("%s = %s", a.Name.Lexeme, a.Value.String())
}
func (a *AssignStmt) stmtNode() {}

// IfStmt: if expr { } else { }
type IfStmt struct {
	Condition Expr
	Then      *Block
	Else      *Block // nil if no else
}

func (i *IfStmt) TokenLiteral() string { return "if" }
func (i *IfStmt) String() string {
	s := fmt.Sprintf("if %s %s", i.Condition.String(), i.Then.String())
	if i.Else != nil {
		s += " else " + i.Else.String()
	}
	return s
}
func (i *IfStmt) stmtNode() {}

// WhileStmt: while expr { }
type WhileStmt struct {
	Condition Expr
	Body      *Block
}

func (w *WhileStmt) TokenLiteral() string { return "while" }
func (w *WhileStmt) String() string {
	return fmt.Sprintf("while %s %s", w.Condition.String(), w.Body.String())
}
func (w *WhileStmt) stmtNode() {}

// RunStmt: run build
type RunStmt struct {
	Target token.Token
}

func (r *RunStmt) TokenLiteral() string { return "run" }
func (r *RunStmt) String() string       { return fmt.Sprintf("run %s", r.Target.Lexeme) }
func (r *RunStmt) stmtNode()            {}

// ReturnStmt: return expr
type ReturnStmt struct {
	Value Expr
}

func (r *ReturnStmt) TokenLiteral() string { return "return" }
func (r *ReturnStmt) String() string       { return fmt.Sprintf("return %s", r.Value.String()) }
func (r *ReturnStmt) stmtNode()            {}

// ExprStmt: a call used as a statement
type ExprStmt struct {
	Expr Expr
}

func (e *ExprStmt) TokenLiteral() string { return e.Expr.TokenLiteral() }
func (e *ExprStmt) String() string       { return e.Expr.String() }
func (e *ExprStmt) stmtNode()            {}

// BinaryExpr: left op right
type BinaryExpr struct {
	Op    token.Token
	Left  Expr
	Right Expr
}

func (b *BinaryExpr) TokenLiteral() string { return b.Op.Lexeme }
func (b *BinaryExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Op.Lexeme, b.Right.String())
}
func (b *BinaryExpr) exprNode() {}

// UnaryExpr: op operand
type UnaryExpr struct {
	Op      token.Token
	Operand Expr
}

func (u *UnaryExpr) TokenLiteral() string { return u.Op.Lexeme }
func (u *UnaryExpr) String() string {
	return fmt.Sprintf("(%s%s)", u.Op.Lexeme, u.Operand.String())
}
func (u *UnaryExpr) exprNode() {}

// CallExpr: name(arg, ...)
type CallExpr struct {
	Name token.Token
	Args []Expr
}

func (c *CallExpr) TokenLiteral() string { return c.Name.Lexeme }
func (c *CallExpr) String() string {
	var args []string
	for _, a := range c.Args {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%s(%s)", c.Name.Lexeme, strings.Join(args, ", "))
}
func (c *CallExpr) exprNode() {}

// FieldExpr: task.field
type FieldExpr struct {
	Object token.Token
	Field  token.Token
}

func (f *FieldExpr) TokenLiteral() string { return f.Object.Lexeme }
func (f *FieldExpr) String() string {
	return fmt.Sprintf("%s.%s", f.Object.Lexeme, f.Field.Lexeme)
}
func (f *FieldExpr) exprNode() {}

// Ident: variable reference
type Ident struct {
	Token token.Token
}

func (i *Ident) TokenLiteral() string { return i.Token.Lexeme }
func (i *Ident) String() string       { return i.Token.Lexeme }
func (i *Ident) exprNode()            {}

// IntLit: 42
type IntLit struct {
	Token token.Token
	Value int64
}

func (i *IntLit) TokenLiteral() string { return i.Token.Lexeme }
func (i *IntLit) String() string       { return i.Token.Lexeme }
func (i *IntLit) exprNode()            {}

// FloatLit: 3.14
type FloatLit struct {
	Token token.Token
	Value float64
}

func (f *FloatLit) TokenLiteral() string { return f.Token.Lexeme }
func (f *FloatLit) String() string       { return f.Token.Lexeme }
func (f *FloatLit) exprNode()            {}

// BoolLit: true / false
type BoolLit struct {
	Token token.Token
	Value bool
}

func (b *BoolLit) TokenLiteral() string { return b.Token.Lexeme }
func (b *BoolLit) String() string       { return b.Token.Lexeme }
func (b *BoolLit) exprNode()            {}
