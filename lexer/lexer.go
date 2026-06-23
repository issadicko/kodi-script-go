// Package lexer provides tokenization for KodiScript source code.
package lexer

import (
	"unicode"

	"github.com/issadicko/kodi-script-go/token"
)

// Lexer tokenizes KodiScript source code.
// It operates on runes (not bytes) so that Unicode identifiers and string
// contents are handled correctly.
type Lexer struct {
	runes        []rune      // source decoded as runes
	position     int         // current position in runes (points to current char)
	readPosition int         // current reading position (after current char)
	ch           rune        // current char under examination
	line         int         // current line number
	column       int         // current column number
	prevToken    token.Token // previous token for ASI
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	l := &Lexer{runes: []rune(input), line: 1, column: 0}
	// Initialize prevToken to a type that cannot end a statement
	// This ensures leading newlines are skipped correctly
	l.prevToken = token.Token{Type: token.ILLEGAL}
	l.readChar()
	return l
}

// readChar advances the lexer by one character.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.runes) {
		l.ch = 0
	} else {
		l.ch = l.runes[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// peekChar returns the next character without advancing.
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.runes) {
		return 0
	}
	return l.runes[l.readPosition]
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
		if l.peekChar() == '+' {
			l.readChar()
			tok = token.Token{Type: token.PLUS_PLUS, Literal: "++", Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.PLUS_EQ, Literal: "+=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.PLUS, l.ch)
		}
	case '-':
		if l.peekChar() == '-' {
			l.readChar()
			tok = token.Token{Type: token.MINUS_MINUS, Literal: "--", Line: l.line, Column: l.column - 1}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.MINUS_EQ, Literal: "-=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.MINUS, l.ch)
		}
	case '*':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.ASTERISK_EQ, Literal: "*=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.ASTERISK, l.ch)
		}
	case '/':
		if l.peekChar() == '/' {
			l.skipLineComment()
			return l.NextToken()
		} else if l.peekChar() == '*' {
			l.skipBlockComment()
			return l.NextToken()
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.SLASH_EQ, Literal: "/=", Line: l.line, Column: l.column - 1}
		} else {
			tok = l.newToken(token.SLASH, l.ch)
		}
	case '%':
		tok = l.newToken(token.PERCENT, l.ch)
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
			tok = l.newToken(token.QUESTION, l.ch)
		}
	case ',':
		tok = l.newToken(token.COMMA, l.ch)
	case ';':
		tok = l.newToken(token.SEMICOLON, l.ch)
	case ':':
		tok = l.newToken(token.COLON, l.ch)
	case '(':
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		tok = l.newToken(token.RPAREN, l.ch)
	case '{':
		tok = l.newToken(token.LBRACE, l.ch)
	case '}':
		tok = l.newToken(token.RBRACE, l.ch)
	case '[':
		tok = l.newToken(token.LBRACKET, l.ch)
	case ']':
		tok = l.newToken(token.RBRACKET, l.ch)
	case '.':
		if l.peekChar() == '.' {
			l.readChar() // consume second '.'
			if l.peekChar() == '.' {
				l.readChar() // consume third '.'
				tok = token.Token{Type: token.ELLIPSIS, Literal: "...", Line: l.line, Column: l.column - 2}
			} else {
				tok = l.newToken(token.ILLEGAL, l.ch)
			}
		} else {
			tok = l.newToken(token.DOT, l.ch)
		}
	case '"', '\'', '`':
		str, isTemplate := l.readString(l.ch)
		if isTemplate {
			tok.Type = token.STRING_TEMPLATE
		} else {
			tok.Type = token.STRING
		}
		tok.Literal = str
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
func (l *Lexer) newToken(tokenType token.Type, ch rune) token.Token {
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

// skipBlockComment skips a /* ... */ comment (l.ch is '/', peek is '*').
func (l *Lexer) skipBlockComment() {
	l.readChar() // consume '/'
	l.readChar() // consume '*'
	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // consume '*'
			l.readChar() // consume '/'
			return
		}
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

// readIdentifier reads an identifier (letter followed by letters/digits).
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return string(l.runes[position:l.position])
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
	return string(l.runes[position:l.position])
}

// readString reads a string literal delimited by the given quote rune
// (" ' or `), with escape support. Returns the string content and whether it
// contains template expressions (${...}).
func (l *Lexer) readString(quote rune) (string, bool) {
	var result []rune
	isTemplate := false
	l.readChar() // skip opening quote
	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case '"':
				result = append(result, '"')
			case '\'':
				result = append(result, '\'')
			case '`':
				result = append(result, '`')
			case '\\':
				result = append(result, '\\')
			case '$':
				result = append(result, '$')
			default:
				result = append(result, l.ch)
			}
		} else if l.ch == '$' && l.peekChar() == '{' {
			// Template expression detected
			isTemplate = true
			result = append(result, '$', '{')
			l.readChar() // consume $
		} else {
			if l.ch == '\n' {
				l.line++
				l.column = 0
			}
			result = append(result, l.ch)
		}
		l.readChar()
	}
	return string(result), isTemplate
}

// isLetter reports whether ch can start or continue an identifier.
// Unicode letters are allowed so identifiers like "café" or "naïve" work.
func isLetter(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

// isDigit reports whether ch is an ASCII decimal digit.
func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
