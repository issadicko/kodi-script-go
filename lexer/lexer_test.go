package lexer

import (
	"testing"

	"github.com/issadicko/kodi-script-go/token"
)

func TestNextToken(t *testing.T) {
	input := `let x = 42
let name = "hello"
if (x > 10) {
    return x + 1
}
`
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.NUMBER, "42"},
		{token.NEWLINE, "\\n"},
		{token.LET, "let"},
		{token.IDENT, "name"},
		{token.ASSIGN, "="},
		{token.STRING, "hello"},
		{token.NEWLINE, "\\n"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.GT, ">"},
		{token.NUMBER, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.NUMBER, "1"},
		{token.NEWLINE, "\\n"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\\n"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}
	}
}

func TestOperators(t *testing.T) {
	input := `== != < > <= >= && || ?: ?. ! + - * /`

	tests := []token.Type{
		token.EQ,
		token.NOT_EQ,
		token.LT,
		token.GT,
		token.LT_EQ,
		token.GT_EQ,
		token.AND,
		token.OR,
		token.ELVIS,
		token.SAFE_ACCESS,
		token.NOT,
		token.PLUS,
		token.MINUS,
		token.ASTERISK,
		token.SLASH,
		token.EOF,
	}

	l := New(input)

	for i, expected := range tests {
		tok := l.NextToken()
		if tok.Type != expected {
			t.Fatalf("tests[%d] - expected=%q, got=%q", i, expected, tok.Type)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := `let if else true false null return`

	tests := []token.Type{
		token.LET,
		token.IF,
		token.ELSE,
		token.TRUE,
		token.FALSE,
		token.NULL,
		token.RETURN,
		token.EOF,
	}

	l := New(input)

	for i, expected := range tests {
		tok := l.NextToken()
		if tok.Type != expected {
			t.Fatalf("tests[%d] - expected=%q, got=%q", i, expected, tok.Type)
		}
	}
}

func TestString(t *testing.T) {
	input := `"hello world"`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.STRING || tok.Literal != "hello world" {
		t.Fatalf("expected STRING 'hello world', got %q %q", tok.Type, tok.Literal)
	}
}

func TestStringEscapes(t *testing.T) {
	input := `"hello\nworld\t\"test\"\\done"`
	l := New(input)
	tok := l.NextToken()
	expected := "hello\nworld\t\"test\"\\done"
	if tok.Literal != expected {
		t.Fatalf("expected %q, got %q", expected, tok.Literal)
	}
}

func TestNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"42", "42"},
		{"3.14", "3.14"},
		{"100.0", "100.0"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != token.NUMBER || tok.Literal != tt.expected {
			t.Fatalf("expected NUMBER %q, got %q %q", tt.expected, tok.Type, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `let x = 1 // this is a comment
let y = 2`

	l := New(input)

	// let
	tok := l.NextToken()
	if tok.Type != token.LET {
		t.Fatalf("expected LET, got %q", tok.Type)
	}
	// x
	l.NextToken()
	// =
	l.NextToken()
	// 1
	l.NextToken()
	// newline
	tok = l.NextToken()
	if tok.Type != token.NEWLINE {
		t.Fatalf("expected NEWLINE, got %q", tok.Type)
	}
	// let (second line)
	tok = l.NextToken()
	if tok.Type != token.LET {
		t.Fatalf("expected LET on second line, got %q", tok.Type)
	}
}

func TestDelimiters(t *testing.T) {
	input := `() {} , ; .`

	tests := []token.Type{
		token.LPAREN,
		token.RPAREN,
		token.LBRACE,
		token.RBRACE,
		token.COMMA,
		token.SEMICOLON,
		token.DOT,
		token.EOF,
	}

	l := New(input)

	for i, expected := range tests {
		tok := l.NextToken()
		if tok.Type != expected {
			t.Fatalf("tests[%d] - expected=%q, got=%q", i, expected, tok.Type)
		}
	}
}

func TestIllegalCharacter(t *testing.T) {
	input := `@`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL, got %q", tok.Type)
	}
}

func TestLineContinuation(t *testing.T) {
	// Expression continuation should skip newlines
	input := `1 +
2`
	l := New(input)

	// Get all tokens
	tok1 := l.NextToken() // 1
	tok2 := l.NextToken() // +
	tok3 := l.NextToken() // 2 (newline should be skipped after +)

	if tok1.Type != token.NUMBER {
		t.Fatalf("expected NUMBER, got %q", tok1.Type)
	}
	if tok2.Type != token.PLUS {
		t.Fatalf("expected PLUS, got %q", tok2.Type)
	}
	if tok3.Type != token.NUMBER {
		t.Fatalf("expected NUMBER after continuation, got %q", tok3.Type)
	}
}
