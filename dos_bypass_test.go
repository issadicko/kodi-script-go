package kodi

import "testing"

// buildShared returns a script prologue that builds `a`, a value whose logical
// size doubles at every level while its memory stays tiny (a shared DAG). A
// naive stringify expands it to 2^40 nodes; a bounded one must refuse instead.
const buildShared = "let a = [1]\nfor (i in range(40)) { a = [a, a] }\n"

// TestDoSMemoryBypassesClosed guards the paths that the first #3 fix (a size cap
// on repeat/pad/range/evalPlus/evalElements only) left open: string templates,
// the string natives, flatten/concat, the range overflow, and every operation
// that stringifies a value. Each must return an error instead of allocating
// without bound; several would hang or crash the process without the guard.
func TestDoSMemoryBypassesClosed(t *testing.T) {
	refused := map[string]string{
		// A template doubles the string without going through evalPlus.
		"string template doubling": "let a = \"x\"\nfor (i in range(30)) { a = \"${a}${a}\" }\na",
		// Natives that build a string far larger than their inputs.
		"replace expansion":      `replace(repeat("a", 2000), "a", repeat("b", 100000))`,
		"regexReplace expansion": `regexReplace(repeat("a", 2000), "a", repeat("b", 100000))`,
		"join separator":         `join(range(100), repeat("a", 1000000))`,
		// Natives that build an array past the element cap.
		"concat arrays":  `concat(range(10000000), range(10000000))`,
		"flatten arrays": `flatten([range(10000000), range(10000000)])`,
		// range: the int64 span subtraction used to overflow and loop forever.
		"range int64 overflow": `range(-9000000000000000000, 9000000000000000000)`,
		"range span too large": `range(9000000000000000000)`,
		// base64 grows its input by 4/3 with no cap of its own.
		"base64 growth": `base64Encode(repeat("a", 13000000))`,
		// Every path that renders a value must bound a shared structure.
		"evalPlus shared":      buildShared + `"x" + a`,
		"template shared":      buildShared + `"${a}"`,
		"toString shared":      buildShared + `toString(a)`,
		"jsonStringify shared": buildShared + `jsonStringify(a)`,
		"print shared":         buildShared + `print(a)`,
	}
	for name, src := range refused {
		t.Run(name, func(t *testing.T) {
			res := New(src).Execute()
			if len(res.Errors) == 0 {
				t.Fatalf("%s: unbounded operation was not refused (value %v)", name, res.Value)
			}
		})
	}
}

// TestDoSMemoryLegitimateStillWorks checks the guards do not reject ordinary
// use, and that the raised default (16 MiB) admits a payload the old 1 MiB cap
// wrongly refused (e.g. a document carried in an invocation body).
func TestDoSMemoryLegitimateStillWorks(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want interface{}
	}{
		{"replace", `replace("hello", "l", "L")`, "heLLo"},
		{"regexReplace", `regexReplace("a1b2", "[0-9]", "#")`, "a#b#"},
		{"join", `join(["a", "b", "c"], "-")`, "a-b-c"},
		{"concat", `sum(concat([1, 2], [3]))`, float64(6)},
		{"flatten", `sum(flatten([[1, 2], [3]]))`, float64(6)},
		{"range negative", `sum(range(-3, 3))`, float64(-3)},
		{"template", "let n = 5\n\"val=${n}\"", "val=5"},
		{"base64", `base64Encode("hello")`, "aGVsbG8="},
		// 2 MiB document concatenation: refused by the old 1 MiB cap, fine now.
		{"large document concat", `length("x" + repeat("a", 2000000))`, float64(2000001)},
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

// TestMaxValueSizeConfigurable verifies the host can tighten or raise the
// ceiling at startup through the public API.
func TestMaxValueSizeConfigurable(t *testing.T) {
	old := GetMaxValueSize()
	defer SetMaxValueSize(old)

	SetMaxValueSize(1000)
	if res := New(`repeat("a", 2000)`).Execute(); len(res.Errors) == 0 {
		t.Fatalf("expected refusal above the lowered limit, got %v", res.Value)
	}
	if res := New(`length(repeat("a", 500))`).Execute(); len(res.Errors) > 0 || res.Value != float64(500) {
		t.Fatalf("expected 500 under the lowered limit, errors=%v value=%v", res.Errors, res.Value)
	}

	SetMaxValueSize(64 << 20)
	if res := New(`length(repeat("a", 2000000))`).Execute(); len(res.Errors) > 0 || res.Value != float64(2000000) {
		t.Fatalf("expected 2000000 after raising the limit, errors=%v value=%v", res.Errors, res.Value)
	}
}
