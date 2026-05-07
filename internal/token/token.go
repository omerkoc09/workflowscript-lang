package token

type Type int

const (
	ILLEGAL Type = iota
	EOF

	// Literals
	INT_LIT   // 42
	FLOAT_LIT // 3.14
	BOOL_LIT  // true, false
	IDENT     // my_task

	// Declaration keywords
	TASK       // task
	PIPELINE   // pipeline
	STAGE      // stage
	ON_FAILURE // on_failure
	FUN        // fun
	VAR        // var
	RETURN     // return
	IF         // if
	ELSE       // else
	WHILE      // while
	RUN        // run

	// Type keywords
	KW_INT   // int
	KW_FLOAT // float
	KW_BOOL  // bool
	KW_VOID  // void

	// Task field keywords
	KW_TIMEOUT  // timeout
	KW_RETRIES  // retries
	KW_PARALLEL // parallel

	// Operators
	PLUS   // +
	MINUS  // -
	STAR   // *
	SLASH  // /
	EQ     // ==
	NEQ    // !=
	LT     // <
	LTE    // <=
	GT     // >
	GTE    // >=
	AND    // &&
	OR     // ||
	BANG   // !
	ASSIGN // =
	ARROW  // ->
	DOT    // .

	// Separators
	LPAREN // (
	RPAREN // )
	LBRACE // {
	RBRACE // }
	COLON  // :
	COMMA  // ,
)

type Token struct {
	Type   Type
	Lexeme string
	Line   int
}

var keywords = map[string]Type{
	"task":       TASK,
	"pipeline":   PIPELINE,
	"stage":      STAGE,
	"on_failure": ON_FAILURE,
	"fun":        FUN,
	"var":        VAR,
	"return":     RETURN,
	"if":         IF,
	"else":       ELSE,
	"while":      WHILE,
	"run":        RUN,
	"int":        KW_INT,
	"float":      KW_FLOAT,
	"bool":       KW_BOOL,
	"void":       KW_VOID,
	"timeout":    KW_TIMEOUT,
	"retries":    KW_RETRIES,
	"parallel":   KW_PARALLEL,
	"true":       BOOL_LIT,
	"false":      BOOL_LIT,
}

func LookupIdent(s string) Type {
	if t, ok := keywords[s]; ok {
		return t
	}
	return IDENT
}

var names = map[Type]string{
	ILLEGAL: "ILLEGAL", EOF: "EOF",
	INT_LIT: "INT_LIT", FLOAT_LIT: "FLOAT_LIT", BOOL_LIT: "BOOL_LIT", IDENT: "IDENT",
	TASK: "TASK", PIPELINE: "PIPELINE", STAGE: "STAGE", ON_FAILURE: "ON_FAILURE",
	FUN: "FUN", VAR: "VAR", RETURN: "RETURN", IF: "IF", ELSE: "ELSE",
	WHILE: "WHILE", RUN: "RUN",
	KW_INT: "KW_INT", KW_FLOAT: "KW_FLOAT", KW_BOOL: "KW_BOOL", KW_VOID: "KW_VOID",
	KW_TIMEOUT: "KW_TIMEOUT", KW_RETRIES: "KW_RETRIES", KW_PARALLEL: "KW_PARALLEL",
	PLUS: "PLUS", MINUS: "MINUS", STAR: "STAR", SLASH: "SLASH",
	EQ: "EQ", NEQ: "NEQ", LT: "LT", LTE: "LTE", GT: "GT", GTE: "GTE",
	AND: "AND", OR: "OR", BANG: "BANG", ASSIGN: "ASSIGN", ARROW: "ARROW", DOT: "DOT",
	LPAREN: "LPAREN", RPAREN: "RPAREN", LBRACE: "LBRACE", RBRACE: "RBRACE",
	COLON: "COLON", COMMA: "COMMA",
}

func (t Type) String() string {
	if s, ok := names[t]; ok {
		return s
	}
	return "UNKNOWN"
}
