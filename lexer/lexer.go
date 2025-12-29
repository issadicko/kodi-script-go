// Package lexer provides tokenization for KodiScript source code.
package lexer

import (
	"github.com/issadicko/kodi-script-go/token"
)

// Lexer tokenizes KodiScript source code.
type Lexer struct {
	input        string
	position     int         // current position in input (points to current char)
	readPosition int         // current reading position in input (after current char)
	ch           byte        // current char under examination
	line         int         // current line number
	column       int         // current column number
	prevToken    token.Token // previous token for ASI
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	// Initialize prevToken to a type that cannot end a statement
	// This ensures leading newlines are skipped correctly
	l.prevToken = token.Token{Type: token.ILLEGAL}
	l.readChar()
	return l
}

// readChar advances the lexer by one character.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// peekChar returns the next character without advancing.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "==", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.ASSIGN, l.ch)
		}
	case '+':
		tok = l.newToken(token.PLUS, l.ch)
	case '-':
		tok = l.newToken(token.MINUS, l.ch)
	case '*':
		tok = l.newToken(token.ASTERISK, l.ch)
	case '/':
		if l.peekChar() == '/' {
			l.skipLineComment()
			return l.NextToken()
		}
		tok = l.newToken(token.SLASH, l.ch)
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "!=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.NOT, l.ch)
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LT_EQ, Literal: "<=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GT_EQ, Literal: ">=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.GT, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: "&&", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: "||", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case '?':
		if l.peekChar() == '.' {
			l.readChar()
			tok = token.Token{Type: token.SAFE_ACCESS, Literal: "?.", Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == ':' {
			l.readChar()
			tok = token.Token{Type: token.ELVIS, Literal: "?:", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case ',':
		tok = l.newToken(token.COMMA, l.ch)
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		tok = l.newToken(token.RPAREN, l.ch)
	case '{':
		tok = l.newToken(token.LBRACE, l.ch)
	case '}':
		tok = l.newToken(token.RBRACE, l.ch)
	case '.':
		tok = l.newToken(token.DOT, l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '\n':
		// Check if previous token can end a statement (ASI)
		if l.prevToken.Type.CanEndStatement() {
			tok = token.Token{Type: token.NEWLINE, Literal: "\\n", Line: l.line, Column: l.column}
		} else {
			// Skip newline and continue (expression continues on next line)
			l.line++
			l.column = 0
			l.readChar()
			return l.NextToken()
		}
		l.line++
		l.column = 0
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			l.prevToken = tok
			return tok
		} else if isDigit(l.ch) {
			tok.Literal = l.readNumber()
			tok.Type = token.NUMBER
			l.prevToken = tok
			return tok
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	l.prevToken = tok
	return tok
}

// newToken creates a new token with the given type and character.
func (l *Lexer) newToken(tokenType token.Type, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: l.line, Column: l.column}
}

// skipWhitespace skips spaces and tabs (but NOT newlines).
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// skipLineComment skips a // comment until end of line.
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// readIdentifier reads an identifier (letter followed by letters/digits).
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number (integer or float).
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	// Handle decimal numbers
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[position:l.position]
}

// readString reads a string literal (with escape support).
func (l *Lexer) readString() string {
	var result []byte
	l.readChar() // skip opening quote
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case '"':
				result = append(result, '"')
			case '\\':
				result = append(result, '\\')
			default:
				result = append(result, l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}
	return string(result)
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
