// Package token defines the token types used by the KodiScript lexer.
package token

// Type represents the type of a token.
type Type string

const (
	// Special tokens
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"
	NEWLINE Type = "NEWLINE"

	// Identifiers and literals
	IDENT           Type = "IDENT"           // variable names
	NUMBER          Type = "NUMBER"          // 123, 45.67
	STRING          Type = "STRING"          // "hello"
	STRING_TEMPLATE Type = "STRING_TEMPLATE" // "hello ${name}"

	// Operators
	ASSIGN   Type = "="
	PLUS     Type = "+"
	MINUS    Type = "-"
	ASTERISK Type = "*"
	SLASH    Type = "/"
	PERCENT  Type = "%"

	// Comparison
	EQ     Type = "=="
	NOT_EQ Type = "!="
	LT     Type = "<"
	GT     Type = ">"
	LT_EQ  Type = "<="
	GT_EQ  Type = ">="

	// Logical
	AND Type = "&&"
	OR  Type = "||"
	NOT Type = "!"

	// Null-safety operators
	SAFE_ACCESS Type = "?." // Optional chaining
	ELVIS       Type = "?:" // Null coalescing

	// Delimiters
	COMMA     Type = ","
	SEMICOLON Type = ";"
	COLON     Type = ":"
	LPAREN    Type = "("
	RPAREN    Type = ")"
	LBRACE    Type = "{"
	RBRACE    Type = "}"
	LBRACKET  Type = "["
	RBRACKET  Type = "]"
	DOT       Type = "."

	// Keywords
	LET    Type = "LET"
	IF     Type = "IF"
	ELSE   Type = "ELSE"
	TRUE   Type = "TRUE"
	FALSE  Type = "FALSE"
	NULL   Type = "NULL"
	RETURN Type = "RETURN"
	FOR    Type = "FOR"
	IN     Type = "IN"
	FN     Type = "FN"
)

// Token represents a single token with its type, literal value, and position.
type Token struct {
	Type    Type
	Literal string
	Line    int
	Column  int
}

// LookupIdent checks if an identifier is a keyword and returns the appropriate token type.
// Uses switch statement for better performance than map lookup.
func LookupIdent(ident string) Type {
	switch ident {
	case "let":
		return LET
	case "if":
		return IF
	case "else":
		return ELSE
	case "true":
		return TRUE
	case "false":
		return FALSE
	case "null":
		return NULL
	case "return":
		return RETURN
	case "for":
		return FOR
	case "in":
		return IN
	case "fn":
		return FN
	default:
		return IDENT
	}
}

// CanEndStatement returns true if this token type can end a statement (for ASI).
func (t Type) CanEndStatement() bool {
	switch t {
	case IDENT, NUMBER, STRING, STRING_TEMPLATE, TRUE, FALSE, NULL, RPAREN, RBRACE, RBRACKET:
		return true
	default:
		return false
	}
}

// IsOperatorContinuation returns true if this token type indicates the statement continues.
func (t Type) IsOperatorContinuation() bool {
	switch t {
	case PLUS, MINUS, ASTERISK, SLASH, PERCENT, AND, OR, EQ, NOT_EQ, LT, GT, LT_EQ, GT_EQ, SAFE_ACCESS, ELVIS, DOT, COMMA:
		return true
	default:
		return false
	}
}
