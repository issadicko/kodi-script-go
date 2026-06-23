package kodi

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// evalNum runs a script and returns the single numeric value printed via print().
func outOf(t *testing.T, src string) []string {
	t.Helper()
	res := New(src).SilentPrint(true).Execute()
	if len(res.Errors) > 0 {
		t.Fatalf("unexpected errors for %q: %v", src, res.Errors)
	}
	return res.Output
}

// ---- Lot 1: break / continue / compound assignment / ++ -- ----

func TestBreakInForLoop(t *testing.T) {
	out := outOf(t, "let s=0\nfor (i in [1,2,3,4,5]) {\n if (i==3) { break }\n s = s + i\n}\nprint(s)")
	if out[0] != "3" {
		t.Errorf("expected 3, got %v", out)
	}
}

func TestContinueInForLoop(t *testing.T) {
	out := outOf(t, "let s=0\nfor (i in [1,2,3,4]) {\n if (i==2) { continue }\n s += i\n}\nprint(s)")
	if out[0] != "8" {
		t.Errorf("expected 8, got %v", out)
	}
}

func TestBreakInWhileLoop(t *testing.T) {
	out := outOf(t, "let i=0\nwhile (true) {\n i++\n if (i>=5) { break }\n}\nprint(i)")
	if out[0] != "5" {
		t.Errorf("expected 5, got %v", out)
	}
}

func TestCompoundAssignment(t *testing.T) {
	out := outOf(t, "let x=10\nx += 5\nx -= 2\nx *= 2\nx /= 13\nprint(x)")
	if out[0] != "2" {
		t.Errorf("expected 2, got %v", out)
	}
}

func TestIncrementDecrement(t *testing.T) {
	out := outOf(t, "let n=5\nn++\nn++\nn--\nprint(n)")
	if out[0] != "6" {
		t.Errorf("expected 6, got %v", out)
	}
}

func TestNestedBreakOnlyExitsInnerLoop(t *testing.T) {
	out := outOf(t, "let c=0\nfor (i in [1,2]) {\n for (j in [1,2,3]) {\n  if (j==2) { break }\n  c++\n }\n}\nprint(c)")
	if out[0] != "2" {
		t.Errorf("expected 2, got %v", out)
	}
}

// ---- Lot 2: method-call syntax + de-hardcoded builtins ----

func TestMethodCallSyntax(t *testing.T) {
	cases := map[string]string{
		`print("hello".toUpperCase())`:                              "HELLO",
		`print("  Hi ".trim().toLowerCase())`:                       "hi", // chained
		`print([1,2,3].size())`:                                     "3",
		`print(["a","b"].join("-"))`:                                "a-b",
		`print([1,2,3].map(fn(x){ x*2 }))`:                          "[2, 4, 6]",
		`print([1,2,3,4].filter(fn(x){ x>2 }))`:                     "[3, 4]",
		`let o = {greet: fn(){ "hi" }}` + "\n" + `print(o.greet())`: "hi",
	}
	for src, want := range cases {
		out := outOf(t, src)
		if len(out) != 1 || out[0] != want {
			t.Errorf("%q: expected %q, got %v", src, want, out)
		}
	}
}

func TestBuiltinsAreOverridable(t *testing.T) {
	// Redefining print as a user function suppresses default output capture.
	res := New(`let print = fn(x){ x }` + "\n" + `print("hidden")`).SilentPrint(true).Execute()
	if len(res.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(res.Output) != 0 {
		t.Errorf("expected no captured output when print is overridden, got %v", res.Output)
	}
}

func TestBareBuiltinsStillWork(t *testing.T) {
	out := outOf(t, `print(toUpperCase("x"))`+"\n"+`print(map([1,2], fn(x){ x+1 }))`)
	if len(out) != 2 || out[0] != "X" || out[1] != "[2, 3]" {
		t.Errorf("unexpected output: %v", out)
	}
}

// ---- Lot 3: stdlib expansion ----

func TestStdlibExpansion(t *testing.T) {
	cases := map[string]string{
		`print(range(4))`:                 "[0, 1, 2, 3]",
		`print(range(2,5))`:               "[2, 3, 4]",
		`print(sum([1,2,3,4]))`:           "10",
		`print(avg([2,4,6]))`:             "4",
		`print(unique([1,1,2,3,3]))`:      "[1, 2, 3]",
		`print(flatten([[1,2],[3]]))`:     "[1, 2, 3]",
		`print(push([1,2],3,4))`:          "[1, 2, 3, 4]",
		`print(concat([1],[2,3]))`:        "[1, 2, 3]",
		`print(keys({b:2, a:1, c:3}))`:    "[a, b, c]", // sorted for determinism
		`print(values({b:2, a:1}))`:       "[1, 2]",    // sorted by key
		`print(has({a:1}, "a"))`:          "true",
		`print(has([1,2,3], 2))`:          "true",
		`print(parseInt("3.9"))`:          "3",
		`print(parseFloat("3.14"))`:       "3.14",
		`print([3,1,2,1].unique().sum())`: "6", // chained with new natives
	}
	for src, want := range cases {
		out := outOf(t, src)
		if len(out) != 1 || out[0] != want {
			t.Errorf("%q: expected %q, got %v", src, want, out)
		}
	}
}

// ---- B: named functions, ternary, else-if, block comments ----

func TestNamedFunctionAndRecursion(t *testing.T) {
	out := outOf(t, "fn fact(n) {\n if (n <= 1) { return 1 }\n return n * fact(n - 1)\n}\nprint(fact(5))")
	if out[0] != "120" {
		t.Errorf("expected 120, got %v", out)
	}
}

func TestTernary(t *testing.T) {
	if out := outOf(t, `let x = 5`+"\n"+`print(x > 3 ? "big" : "small")`); out[0] != "big" {
		t.Errorf("ternary: %v", out)
	}
	if out := outOf(t, `let n = 2`+"\n"+`print(n == 1 ? "one" : n == 2 ? "two" : "many")`); out[0] != "two" {
		t.Errorf("nested ternary (right-assoc): %v", out)
	}
}

func TestElseIf(t *testing.T) {
	out := outOf(t, `let g = 85`+"\n"+`if (g >= 90) { print("A") } else if (g >= 80) { print("B") } else { print("C") }`)
	if out[0] != "B" {
		t.Errorf("else-if: %v", out)
	}
}

func TestBlockComments(t *testing.T) {
	out := outOf(t, "let x = 5 /* inline */ + 3\n/* multi\nline */\nprint(x)")
	if len(out) != 1 || out[0] != "8" {
		t.Errorf("block comment: %v", out)
	}
}

// ---- C: try/catch + runtime error positions ----

func TestTryCatch(t *testing.T) {
	out := outOf(t, "let r = \"none\"\ntry {\n let x = boom\n} catch (e) {\n r = \"caught\"\n}\nprint(r)\nprint(\"after\")")
	if len(out) != 2 || out[0] != "caught" || out[1] != "after" {
		t.Errorf("try/catch: %v", out)
	}
	if out = outOf(t, "try {\n let y = boom\n} catch {\n print(\"handled\")\n}"); out[0] != "handled" {
		t.Errorf("catch without var: %v", out)
	}
	if out = outOf(t, "try {\n print(\"ok\")\n} catch (e) {\n print(\"bad\")\n}"); len(out) != 1 || out[0] != "ok" {
		t.Errorf("no error: %v", out)
	}
}

func TestReturnInsideTry(t *testing.T) {
	out := outOf(t, "fn safeDiv(a, b) {\n try {\n  return a / b\n } catch (e) {\n  return -1\n }\n}\nprint(safeDiv(10, 2))\nprint(safeDiv(10, 0))")
	if len(out) != 2 || out[0] != "5" || out[1] != "-1" {
		t.Errorf("return inside try: %v", out)
	}
}

func TestRuntimeErrorHasPosition(t *testing.T) {
	res := New("let a = 1\nlet b = undefinedThing").Execute()
	if len(res.Errors) == 0 || !strings.Contains(res.Errors[0], "line 2") {
		t.Errorf("expected positioned error, got %v", res.Errors)
	}
}

// ---- D: recursion guard (robustness) ----

func TestRecursionGuard(t *testing.T) {
	res := New("fn loop() {\n return loop()\n}\nloop()").Execute()
	if len(res.Errors) == 0 {
		t.Error("expected an error from unbounded recursion, got none")
	}
	// Deep but finite recursion still works.
	out := outOf(t, "fn sum(n) {\n if (n == 0) { return 0 }\n return n + sum(n - 1)\n}\nprint(sum(100))")
	if out[0] != "5050" {
		t.Errorf("deep recursion: %v", out)
	}
}

// ---- E: some/every/flatMap, regex, spread, destructuring ----

func TestSomeEveryFlatMap(t *testing.T) {
	if outOf(t, `print(some([1,2,3], fn(x){ x > 2 }))`)[0] != "true" {
		t.Error("some")
	}
	if outOf(t, `print(every([2,4], fn(x){ x % 2 == 0 }))`)[0] != "true" {
		t.Error("every")
	}
	if outOf(t, `print(flatMap([1,2], fn(x){ [x, x] }))`)[0] != "[1, 1, 2, 2]" {
		t.Error("flatMap")
	}
}

func TestRegex(t *testing.T) {
	if outOf(t, `print(regexMatch("abc123", "[0-9]+"))`)[0] != "true" {
		t.Error("regexMatch")
	}
	if outOf(t, `print(regexReplace("a1b2", "[0-9]", "X"))`)[0] != "aXbX" {
		t.Error("regexReplace")
	}
}

func TestSpread(t *testing.T) {
	if outOf(t, "let a = [1,2]\nprint([...a, 3])")[0] != "[1, 2, 3]" {
		t.Error("spread in array")
	}
	if outOf(t, "fn add(x,y,z){ return x+y+z }\nprint(add(...[1,2,3]))")[0] != "6" {
		t.Error("spread in call")
	}
}

func TestDestructuring(t *testing.T) {
	out := outOf(t, "let [a, b] = [10, 20]\nprint(a)\nprint(b)")
	if out[0] != "10" || out[1] != "20" {
		t.Errorf("array destructure: %v", out)
	}
	out = outOf(t, `let {name, age} = {name: "Bob", age: 25}`+"\n"+`print(name)`+"\n"+`print(age)`)
	if out[0] != "Bob" || out[1] != "25" {
		t.Errorf("object destructure: %v", out)
	}
}

// ---- Lot 4: output sink + typed errors ----

func TestOutputSink(t *testing.T) {
	var sink []string
	res := New(`print("a")` + "\n" + `print("b")`).
		WithOutput(func(s string) { sink = append(sink, s) }).
		Execute()
	if len(res.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", res.Errors)
	}
	if len(sink) != 2 || sink[0] != "a" || sink[1] != "b" {
		t.Errorf("sink got %v", sink)
	}
	if len(res.Output) != 2 { // still captured
		t.Errorf("expected output captured, got %v", res.Output)
	}
}

func TestTypedErrors(t *testing.T) {
	r := New("while (true) { let x = 1 }").WithTimeout(50 * time.Millisecond).Execute()
	if r.Kind != ErrorKindTimeout || !errors.Is(r.Err, ErrTimeout) {
		t.Errorf("timeout: kind=%d err=%v", r.Kind, r.Err)
	}
	r = New("while (true) { let x = 1 }").WithMaxOperations(100).Execute()
	if r.Kind != ErrorKindMaxOperations || !errors.Is(r.Err, ErrMaxOperationsExceeded) {
		t.Errorf("maxops: kind=%d err=%v", r.Kind, r.Err)
	}
	if r = New("undefinedVar + 1").Execute(); r.Kind != ErrorKindRuntime {
		t.Errorf("runtime: kind=%d", r.Kind)
	}
	if r = New("let x = 5").Execute(); r.Kind != ErrorKindNone {
		t.Errorf("ok: kind=%d", r.Kind)
	}
}
