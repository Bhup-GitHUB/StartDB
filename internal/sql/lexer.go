package sql

import (
	"strings"
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenString
	TokenNumber
	TokenKeyword
	TokenLeftParen
	TokenRightParen
	TokenComma
	TokenSemicolon
	TokenEquals
	TokenNotEquals
	TokenLessThan
	TokenGreaterThan
	TokenLessThanOrEqual
	TokenGreaterThanOrEqual
	TokenPlus
	TokenMinus
	TokenAsterisk
	TokenSlash
	TokenAnd
	TokenOr
	TokenNot
	TokenNull
	TokenTrue
	TokenFalse
	TokenIllegal
)

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Lexer represents a SQL lexer
type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

// NewLexer creates a new SQL lexer
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		column: 1,
	}
	l.readChar()
	return l
}

// Next returns the next token
func (l *Lexer) Next() Token {
	var tok Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case 0:
		tok.Type = TokenEOF
		tok.Literal = ""
	case '(':
		tok.Type = TokenLeftParen
		tok.Literal = string(l.ch)
		l.readChar()
	case ')':
		tok.Type = TokenRightParen
		tok.Literal = string(l.ch)
		l.readChar()
	case ',':
		tok.Type = TokenComma
		tok.Literal = string(l.ch)
		l.readChar()
	case ';':
		tok.Type = TokenSemicolon
		tok.Literal = string(l.ch)
		l.readChar()
	case '=':
		tok.Type = TokenEquals
		tok.Literal = string(l.ch)
		l.readChar()
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenNotEquals
			tok.Literal = "!="
		} else {
			tok.Type = TokenNot
			tok.Literal = string(l.ch)
		}
		l.readChar()
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenLessThanOrEqual
			tok.Literal = "<="
		} else if l.peekChar() == '>' {
			l.readChar()
			tok.Type = TokenNotEquals
			tok.Literal = "<>"
		} else {
			tok.Type = TokenLessThan
			tok.Literal = string(l.ch)
		}
		l.readChar()
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenGreaterThanOrEqual
			tok.Literal = ">="
		} else {
			tok.Type = TokenGreaterThan
			tok.Literal = string(l.ch)
		}
		l.readChar()
	case '+':
		tok.Type = TokenPlus
		tok.Literal = string(l.ch)
		l.readChar()
	case '-':
		tok.Type = TokenMinus
		tok.Literal = string(l.ch)
		l.readChar()
	case '*':
		tok.Type = TokenAsterisk
		tok.Literal = string(l.ch)
		l.readChar()
	case '/':
		tok.Type = TokenSlash
		tok.Literal = string(l.ch)
		l.readChar()
	case '\'':
		tok.Type = TokenString
		tok.Literal = l.readString()
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = lookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = TokenNumber
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok.Type = TokenIllegal
			tok.Literal = string(l.ch)
			l.readChar()
		}
	}

	return tok
}

// Peek returns the next token without advancing the position
func (l *Lexer) Peek() Token {
	pos := l.position
	line := l.line
	column := l.column

	tok := l.Next()

	l.position = pos
	l.readPosition = pos + 1
	l.line = line
	l.column = column
	if pos < len(l.input) {
		l.ch = l.input[pos]
	} else {
		l.ch = 0
	}

	return tok
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if l.ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
	}
	// Consume the closing quote
	if l.ch == '\'' {
		l.readChar()
	}
	return l.input[position:l.position-1]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

func lookupIdent(ident string) TokenType {
	switch strings.ToUpper(ident) {
	case "SELECT":
		return TokenKeyword
	case "FROM":
		return TokenKeyword
	case "WHERE":
		return TokenKeyword
	case "INSERT":
		return TokenKeyword
	case "INTO":
		return TokenKeyword
	case "VALUES":
		return TokenKeyword
	case "UPDATE":
		return TokenKeyword
	case "SET":
		return TokenKeyword
	case "DELETE":
		return TokenKeyword
	case "CREATE":
		return TokenKeyword
	case "TABLE":
		return TokenKeyword
	case "DROP":
		return TokenKeyword
	case "INDEX":
		return TokenKeyword
	case "ON":
		return TokenKeyword
	case "ORDER":
		return TokenKeyword
	case "BY":
		return TokenKeyword
	case "LIMIT":
		return TokenKeyword
	case "OFFSET":
		return TokenKeyword
	case "AND":
		return TokenAnd
	case "OR":
		return TokenOr
	case "NOT":
		return TokenNot
	case "NULL":
		return TokenNull
	case "TRUE":
		return TokenTrue
	case "FALSE":
		return TokenFalse
	default:
		return TokenIdentifier
	}
}
