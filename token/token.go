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
	IDENT  Type = "IDENT"  // variable names
	NUMBER Type = "NUMBER" // 123, 45.67
	STRING Type = "STRING" // "hello"

	// Operators
	ASSIGN   Type = "="
	PLUS     Type = "+"
	MINUS    Type = "-"
	ASTERISK Type = "*"
	SLASH    Type = "/"

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

// keywords maps keyword strings to their token types.
var keywords = map[string]Type{
	"let":    LET,
	"if":     IF,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
	"null":   NULL,
	"return": RETURN,
	"for":    FOR,
	"in":     IN,
	"fn":     FN,
}

// LookupIdent checks if an identifier is a keyword and returns the appropriate token type.
func LookupIdent(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// CanEndStatement returns true if this token type can end a statement (for ASI).
func (t Type) CanEndStatement() bool {
	switch t {
	case IDENT, NUMBER, STRING, TRUE, FALSE, NULL, RPAREN, RBRACE, RBRACKET:
		return true
	default:
		return false
	}
}

// IsOperatorContinuation returns true if this token type indicates the statement continues.
func (t Type) IsOperatorContinuation() bool {
	switch t {
	case PLUS, MINUS, ASTERISK, SLASH, AND, OR, EQ, NOT_EQ, LT, GT, LT_EQ, GT_EQ, SAFE_ACCESS, ELVIS, DOT, COMMA:
		return true
	default:
		return false
	}
}
