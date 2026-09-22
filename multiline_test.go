package kodi

import (
	"fmt"
	"testing"
)

// Literals can be laid out one element per line, with a trailing comma, in
// arrays, objects and call arguments alike.
func TestMultilineLiterals(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{"object", "let o = {\n  \"a\": 1,\n  \"b\": 2,\n}\nprint(o.a + o.b)", "3"},
		{"array", "let t = [\n  1,\n  2,\n  3,\n]\nprint(t[2])", "3"},
		{"nested", "let o = {\n  \"a\": [\n    1,\n    {\"b\": 2}\n  ]\n}\nprint(o.a[1].b)", "2"},
		{"call arguments", "let f = fn(a, b) { return a + b }\nprint(f(\n  1,\n  2,\n))", "3"},
		{"element then comma on next line", "let t = [\n  1\n  , 2\n]\nprint(t[1])", "2"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := outOf(t, c.src)
			if len(out) != 1 || out[0] != c.want {
				t.Errorf("expected %q, got %v", c.want, out)
			}
		})
	}
}

// An empty literal may span lines too.
func TestEmptyLiteralOverLines(t *testing.T) {
	v, err := Eval("[\n]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if items, ok := v.([]interface{}); !ok || len(items) != 0 {
		t.Errorf("expected an empty array, got %#v", v)
	}
}

// A multi-line object as the last expression is the value of the script.
func TestMultilineObjectIsAValue(t *testing.T) {
	v, err := Eval("let x = 1\n{\n  \"a\": x,\n  \"b\": [\n    x\n  ]\n}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := fmt.Sprint(v); got != "map[a:1 b:[1]]" {
		t.Errorf("unexpected value: %s", got)
	}
}
