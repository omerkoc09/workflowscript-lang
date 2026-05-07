package lexer

import (
	"fmt"
	"unicode"
	"workflowscript/internal/token"
)

type Lexer struct {
	src  []rune
	pos  int
	line int
}

func New(src string) *Lexer {
	return &Lexer{src: []rune(src), pos: 0, line: 1}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peekAt(offset int) rune {
	i := l.pos + offset
	if i >= len(l.src) {
		return 0
	}
	return l.src[i]
}

func (l *Lexer) advance() rune {
	ch := l.src[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
	}
	return ch
}

func (l *Lexer) skipWhitespaceAndComments() {
	for l.pos < len(l.src) {
		ch := l.peek()
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			l.advance()
		case ch == '/' && l.peekAt(1) == '/':
			for l.pos < len(l.src) && l.peek() != '\n' {
				l.advance()
			}
		default:
			return
		}
	}
}

func (l *Lexer) readIdent() token.Token {
	start := l.pos
	line := l.line
	for l.pos < len(l.src) && (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
		l.advance()
	}
	lexeme := string(l.src[start:l.pos])
	return token.Token{Type: token.LookupIdent(lexeme), Lexeme: lexeme, Line: line}
}

func (l *Lexer) readNumber() token.Token {
	start := l.pos
	line := l.line
	for l.pos < len(l.src) && unicode.IsDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && unicode.IsDigit(l.peekAt(1)) {
		l.advance()
		for l.pos < len(l.src) && unicode.IsDigit(l.peek()) {
			l.advance()
		}
		return token.Token{Type: token.FLOAT_LIT, Lexeme: string(l.src[start:l.pos]), Line: line}
	}
	return token.Token{Type: token.INT_LIT, Lexeme: string(l.src[start:l.pos]), Line: line}
}

func (l *Lexer) next() token.Token {
	l.skipWhitespaceAndComments()
	if l.pos >= len(l.src) {
		return token.Token{Type: token.EOF, Lexeme: "", Line: l.line}
	}
	line := l.line
	ch := l.peek()

	if unicode.IsLetter(ch) || ch == '_' {
		return l.readIdent()
	}
	if unicode.IsDigit(ch) {
		return l.readNumber()
	}

	l.advance()
	switch ch {
	case '+':
		return token.Token{Type: token.PLUS, Lexeme: "+", Line: line}
	case '-':
		if l.peek() == '>' {
			l.advance()
			return token.Token{Type: token.ARROW, Lexeme: "->", Line: line}
		}
		return token.Token{Type: token.MINUS, Lexeme: "-", Line: line}
	case '*':
		return token.Token{Type: token.STAR, Lexeme: "*", Line: line}
	case '/':
		return token.Token{Type: token.SLASH, Lexeme: "/", Line: line}
	case '=':
		if l.peek() == '=' {
			l.advance()
			return token.Token{Type: token.EQ, Lexeme: "==", Line: line}
		}
		return token.Token{Type: token.ASSIGN, Lexeme: "=", Line: line}
	case '!':
		if l.peek() == '=' {
			l.advance()
			return token.Token{Type: token.NEQ, Lexeme: "!=", Line: line}
		}
		return token.Token{Type: token.BANG, Lexeme: "!", Line: line}
	case '<':
		if l.peek() == '=' {
			l.advance()
			return token.Token{Type: token.LTE, Lexeme: "<=", Line: line}
		}
		return token.Token{Type: token.LT, Lexeme: "<", Line: line}
	case '>':
		if l.peek() == '=' {
			l.advance()
			return token.Token{Type: token.GTE, Lexeme: ">=", Line: line}
		}
		return token.Token{Type: token.GT, Lexeme: ">", Line: line}
	case '&':
		if l.peek() == '&' {
			l.advance()
			return token.Token{Type: token.AND, Lexeme: "&&", Line: line}
		}
	case '|':
		if l.peek() == '|' {
			l.advance()
			return token.Token{Type: token.OR, Lexeme: "||", Line: line}
		}
	case '(':
		return token.Token{Type: token.LPAREN, Lexeme: "(", Line: line}
	case ')':
		return token.Token{Type: token.RPAREN, Lexeme: ")", Line: line}
	case '{':
		return token.Token{Type: token.LBRACE, Lexeme: "{", Line: line}
	case '}':
		return token.Token{Type: token.RBRACE, Lexeme: "}", Line: line}
	case ':':
		return token.Token{Type: token.COLON, Lexeme: ":", Line: line}
	case ',':
		return token.Token{Type: token.COMMA, Lexeme: ",", Line: line}
	case '.':
		return token.Token{Type: token.DOT, Lexeme: ".", Line: line}
	}
	return token.Token{Type: token.ILLEGAL, Lexeme: string(ch), Line: line}
}

func (l *Lexer) Tokenize() ([]token.Token, error) {
	var tokens []token.Token
	for {
		tok := l.next()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
		if tok.Type == token.ILLEGAL {
			return nil, fmt.Errorf("line %d: illegal character %q", tok.Line, tok.Lexeme)
		}
	}
	return tokens, nil
}
