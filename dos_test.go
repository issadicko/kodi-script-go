package kodi

import (
	"strings"
	"testing"
)

// TestDoSParserDepthBounded guards #2 at the parser: pathological nesting must
// return a parse error instead of overflowing the native stack and killing the
// process (a parse-only host such as form-builder is affected too). The input is
// well above maxParseDepth; if the guard were absent this test would crash the
// test binary rather than fail.
func TestDoSParserDepthBounded(t *testing.T) {
	const over = 8000 // > maxParseDepth (4000)

	cases := map[string]string{
		"nested parentheses":     strings.Repeat("(", over) + "1" + strings.Repeat(")", over),
		"nested prefix negation": strings.Repeat("!", over) + "true",
		"nested arrays":          strings.Repeat("[", over) + "1" + strings.Repeat("]", over),
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			res := New(src).Execute()
			if len(res.Errors) == 0 {
				t.Fatalf("%s: parser accepted pathological nesting (value %v)", name, res.Value)
			}
		})
	}

	// Moderate nesting still parses and evaluates.
	res := New(strings.Repeat("(", 100) + "1" + strings.Repeat(")", 100)).Execute()
	if len(res.Errors) > 0 {
		t.Fatalf("moderate nesting wrongly rejected: %v", res.Errors)
	}
	if res.Value != float64(1) {
		t.Errorf("moderate nesting: expected 1, got %v", res.Value)
	}
}

// TestDoSEvalDepthBounded guards #2 at the evaluator: a long associative chain
// parses iteratively but builds a deep AST; evaluating it must return an error
// instead of overflowing the stack.
func TestDoSEvalDepthBounded(t *testing.T) {
	const over = 8000 // AST depth > maxExprDepth (4000)

	res := New("1" + strings.Repeat("+1", over)).Execute()
	if len(res.Errors) == 0 {
		t.Fatalf("evaluator accepted a pathologically deep expression (value %v)", res.Value)
	}

	// A shorter chain evaluates correctly.
	ok := New("1" + strings.Repeat("+1", 100)).Execute()
	if len(ok.Errors) > 0 {
		t.Fatalf("moderate chain wrongly rejected: %v", ok.Errors)
	}
	if ok.Value != float64(101) {
		t.Errorf("moderate chain: expected 101, got %v", ok.Value)
	}
}

// TestDoSMemoryBounded guards #3: a single operation must not allocate without
// bound. Each vector below must return an error instead of exhausting memory and
// having the OS kill the process.
func TestDoSMemoryBounded(t *testing.T) {
	cases := map[string]string{
		"repeat count":    `repeat("a", 6000000000)`,
		"range span":      `range(2000000000)`,
		"padLeft length":  `padLeft("x", 9000000000)`,
		"string doubling": "let a = \"a\"\nfor (i in range(40)) { a = a + a }\na",
		"array doubling":  "let a = [1]\nfor (i in range(40)) { a = [...a, ...a] }\na",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			res := New(src).Execute()
			if len(res.Errors) == 0 {
				t.Fatalf("%s: unbounded allocation was not rejected", name)
			}
		})
	}

	// Ordinary sizes still work.
	if v := New(`length(repeat("ab", 3))`).Execute(); len(v.Errors) > 0 || v.Value != float64(6) {
		t.Errorf("repeat: errors=%v value=%v", v.Errors, v.Value)
	}
	if v := New(`sum(range(5))`).Execute(); len(v.Errors) > 0 || v.Value != float64(10) {
		t.Errorf("range: errors=%v value=%v", v.Errors, v.Value)
	}
	if v := New(`padLeft("7", 4, "0")`).Execute(); len(v.Errors) > 0 || v.Value != "0007" {
		t.Errorf("padLeft: errors=%v value=%v", v.Errors, v.Value)
	}
}
