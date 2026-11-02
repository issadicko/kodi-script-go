package kodi

import (
	"errors"
	"testing"
)

// These tests guard the unrecovered panics that could crash the shared
// process (present since v0.3.0): == / != on arrays or objects, substring with
// an end before the start, and randomInt over a span that overflowed int64.
// Each must now return an ordinary result, and Execute must turn any residual
// panic into an error rather than letting it escape the goroutine.

func TestEqualityNeverPanics(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want interface{}
	}{
		{"array equal", `[1, 2] == [1, 2]`, true},
		{"array unequal", `[1, 2] == [1, 3]`, false},
		{"array different length", `[1] == [1, 2]`, false},
		{"array not-equal, equal operands", `[1] != [1]`, false},
		{"array not-equal, distinct operands", `[1] != [2]`, true},
		{"nested arrays equal", `[[1], [2]] == [[1], [2]]`, true},
		{"object equal", `{a: 1, b: 2} == {a: 1, b: 2}`, true},
		{"object key order irrelevant", `{a: 1, b: 2} == {b: 2, a: 1}`, true},
		{"object unequal", `{a: 1} == {a: 2}`, false},
		{"array vs scalar", `[1] == 1`, false},
		{"object vs array", `{a: 1} == [1]`, false},
		// Scalar semantics must stay exactly as before.
		{"number equal", `1 == 1`, true},
		{"number unequal", `1 == 2`, false},
		{"string equal", `"a" == "a"`, true},
		{"string unequal", `"a" == "b"`, false},
		{"bool equal", `true == true`, true},
		{"null equal", `null == null`, true},
		{"null vs number", `null == 0`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := New(c.src).Execute()
			if len(res.Errors) > 0 {
				t.Fatalf("%s: unexpected errors %v", c.name, res.Errors)
			}
			if res.Value != c.want {
				t.Fatalf("%s: got %v (%T), want %v", c.name, res.Value, res.Value, c.want)
			}
		})
	}
}

func TestSubstringNeverPanics(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want interface{}
	}{
		{"negative end", `substring("a", 0, -1)`, ""},
		{"end before start", `substring("abc", 2, 1)`, ""},
		// The realistic trigger: indexOf misses and returns -1.
		{"failed indexOf as end", `substring("abc", 0, indexOf("abc", "@"))`, ""},
		{"normal three-arg", `substring("hello", 1, 3)`, "el"},
		{"two-arg", `substring("hello", 2)`, "llo"},
		{"end past length clamped", `substring("hi", 0, 99)`, "hi"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := New(c.src).Execute()
			if len(res.Errors) > 0 {
				t.Fatalf("%s: unexpected errors %v", c.name, res.Errors)
			}
			if res.Value != c.want {
				t.Fatalf("%s: got %q, want %q", c.name, res.Value, c.want)
			}
		})
	}
}

func TestRandomIntNeverPanics(t *testing.T) {
	// The extreme span overflowed int64 and handed rand.Intn a non-positive
	// argument, which panicked. It must now be refused with an error.
	for _, src := range []string{
		`randomInt(-9000000000000000000, 9000000000000000000)`,
		`randomInt(0, 9000000000000000000)`,
	} {
		res := New(src).Execute()
		if len(res.Errors) == 0 {
			t.Fatalf("%s: expected an out-of-range error, got %v", src, res.Value)
		}
	}
	// Ordinary ranges still work and stay within bounds.
	for i := 0; i < 300; i++ {
		res := New(`randomInt(1, 6)`).Execute()
		if len(res.Errors) > 0 {
			t.Fatalf("randomInt(1, 6) errored: %v", res.Errors)
		}
		n, ok := res.Value.(float64)
		if !ok || n < 1 || n > 6 {
			t.Fatalf("randomInt(1, 6) out of range: %v (%T)", res.Value, res.Value)
		}
	}
}

// TestExecuteRecoversPanic checks the safety net: a native that panics must be
// turned into an error instead of taking the process down.
func TestExecuteRecoversPanic(t *testing.T) {
	res := New(`boom()`).
		RegisterFunction("boom", func(args ...interface{}) (interface{}, error) {
			panic("deliberate panic from a native")
		}).
		Execute()
	if len(res.Errors) == 0 {
		t.Fatal("expected the recovered panic to surface as an error")
	}
	if !errors.Is(res.Err, ErrInternalError) {
		t.Fatalf("expected ErrInternalError, got %v", res.Err)
	}
	if res.Kind != ErrorKindInternal {
		t.Fatalf("expected ErrorKindInternal, got %v", res.Kind)
	}
}
