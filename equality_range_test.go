package kodi

import "testing"

// TestValuesEqualIsTypeFaithful guards the correctness bug where == compared a
// rendered key: StringifyBounded dropped the quotes and the types, so a number
// equalled a string and an array equalled a string that rendered the same.
// Equality must be by type and exact value.
func TestValuesEqualIsTypeFaithful(t *testing.T) {
	falseCases := []string{
		`[1] == ["1"]`,
		`{a: "1"} == {a: 1}`,
		`[true] == ["true"]`,
		`[null] == ["null"]`,
		`["a, b"] == ["a", "b"]`,
		`[[1, 2]] == ["[1, 2]"]`,
		`[1, "2"] == [1, 2]`,
	}
	for _, src := range falseCases {
		t.Run("false/"+src, func(t *testing.T) {
			res := New(src).Execute()
			if len(res.Errors) > 0 {
				t.Fatalf("unexpected errors %v", res.Errors)
			}
			if res.Value != false {
				t.Fatalf("got %v, want false", res.Value)
			}
		})
	}
	trueCases := []string{
		`[1] == [1]`,
		`["1"] == ["1"]`,
		`{a: 1} == {a: 1}`,
		`[[1, 2]] == [[1, 2]]`,
		`[1, "2"] == [1, "2"]`,
	}
	for _, src := range trueCases {
		t.Run("true/"+src, func(t *testing.T) {
			res := New(src).Execute()
			if len(res.Errors) > 0 {
				t.Fatalf("unexpected errors %v", res.Errors)
			}
			if res.Value != true {
				t.Fatalf("got %v, want true", res.Value)
			}
		})
	}
}

// Truncation must not make two different values look equal: a key-prefix
// comparison called [s, 1] and [s, 2] equal once s filled the budget.
func TestValuesEqualExactPastKeyBudget(t *testing.T) {
	src := "let s = repeat(\"a\", 200000)\n[s, 1] == [s, 2]"
	res := New(src).Execute()
	if len(res.Errors) > 0 {
		t.Fatalf("unexpected errors %v", res.Errors)
	}
	if res.Value != false {
		t.Fatalf("got %v, want false", res.Value)
	}
}

// unique and has must not merge a number with a string, nor an array with a
// string that renders the same (valueKey/EqualityKey had the same flaw).
func TestKeyFunctionsAreTypeFaithful(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want interface{}
	}{
		{"unique keeps distinct element types", `size(unique([[1], ["1"]]))`, float64(2)},
		{"unique keeps number and string", `size(unique([1, "1"]))`, float64(2)},
		{"has distinguishes type", `has([1], "1")`, false},
		{"has finds exact", `has([1], 1)`, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := New(c.src).Execute()
			if len(res.Errors) > 0 {
				t.Fatalf("unexpected errors %v", res.Errors)
			}
			if res.Value != c.want {
				t.Fatalf("got %v (%T), want %v", res.Value, res.Value, c.want)
			}
		})
	}
}

// A shared structure must compare in linear time (no exponential blow-up via
// the pointer-pair memo), and a structure too deep to compare must error, not
// overflow the stack.
func TestValuesEqualSharedAndDeep(t *testing.T) {
	shared := New(buildShared + `a == a`).Execute()
	if len(shared.Errors) > 0 {
		t.Fatalf("shared structure equality errored: %v", shared.Errors)
	}
	if shared.Value != true {
		t.Fatalf("shared structure: got %v, want true", shared.Value)
	}
	deep := New("let a = [1]\nfor (i in range(20000)) { a = [a] }\na == a").Execute()
	if len(deep.Errors) == 0 {
		t.Fatalf("expected an error comparing an over-deep structure, got %v", deep.Value)
	}
}

// TestRangeAmd64Overflow guards the amd64-only crash: int(f) of a float beyond
// int64 wraps to MinInt64, so an out-of-range start slipped past the span
// check and then looped from MinInt64 (arm64 saturated and returned empty).
func TestRangeAmd64Overflow(t *testing.T) {
	for _, src := range []string{
		`range(10000000000000000000, 0)`,
		`range(0, 10000000000000000000)`,
		`range(10000000000000000000, 9000000000000000000)`,
	} {
		res := New(src).Execute()
		if len(res.Errors) == 0 {
			t.Fatalf("%s: expected an out-of-range error, got %v", src, res.Value)
		}
	}
	ok := []struct {
		src  string
		want interface{}
	}{
		{`size(range(0, 1000))`, float64(1000)},
		{`size(range(5, 2))`, float64(0)}, // end <= start is empty, not an error
		{`sum(range(-3, 3))`, float64(-3)},
	}
	for _, c := range ok {
		res := New(c.src).Execute()
		if len(res.Errors) > 0 {
			t.Fatalf("%s: unexpected errors %v", c.src, res.Errors)
		}
		if res.Value != c.want {
			t.Fatalf("%s: got %v, want %v", c.src, res.Value, c.want)
		}
	}
}
