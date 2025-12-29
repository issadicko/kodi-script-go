package interpreter

import (
	"testing"

	"github.com/issadicko/kodi-script-go/ast"
	"github.com/issadicko/kodi-script-go/lexer"
	"github.com/issadicko/kodi-script-go/parser"
)

func parseAndEval(source string, vars map[string]interface{}) (Value, error, []string) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, nil, p.Errors()
	}

	var interp *Interpreter
	if vars != nil {
		interp = NewWithEnv(vars)
	} else {
		interp = New()
	}

	result, err := interp.Eval(program)
	return result, err, nil
}

func TestBasicExpressions(t *testing.T) {
	tests := []struct {
		source   string
		expected Value
	}{
		{"42", float64(42)},
		{`"hello"`, "hello"},
		{"true", true},
		{"false", false},
		{"null", nil},
		{"5 + 3", float64(8)},
		{"10 - 4", float64(6)},
		{"3 * 4", float64(12)},
		{"20 / 5", float64(4)},
		{"-5", float64(-5)},
		{"!true", false},
		{"!false", true},
	}

	for _, tt := range tests {
		result, err, errs := parseAndEval(tt.source, nil)
		if len(errs) > 0 {
			t.Fatalf("parse errors for '%s': %v", tt.source, errs)
		}
		if err != nil {
			t.Fatalf("eval error for '%s': %v", tt.source, err)
		}
		if result != tt.expected {
			t.Errorf("'%s': expected %v, got %v", tt.source, tt.expected, result)
		}
	}
}

func TestComparisons(t *testing.T) {
	tests := []struct {
		source   string
		expected bool
	}{
		{"5 > 3", true},
		{"3 > 5", false},
		{"5 < 3", false},
		{"3 < 5", true},
		{"5 >= 5", true},
		{"5 <= 5", true},
		{"5 == 5", true},
		{"5 != 3", true},
	}

	for _, tt := range tests {
		result, err, _ := parseAndEval(tt.source, nil)
		if err != nil {
			t.Fatalf("'%s' error: %v", tt.source, err)
		}
		if result != tt.expected {
			t.Errorf("'%s': expected %v, got %v", tt.source, tt.expected, result)
		}
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		source   string
		expected bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false || true", true},
		{"false || false", false},
	}

	for _, tt := range tests {
		result, err, _ := parseAndEval(tt.source, nil)
		if err != nil {
			t.Fatalf("'%s' error: %v", tt.source, err)
		}
		if result != tt.expected {
			t.Errorf("'%s': expected %v, got %v", tt.source, tt.expected, result)
		}
	}
}

func TestStringConcatenation(t *testing.T) {
	result, err, _ := parseAndEval(`"hello" + " " + "world"`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %v", result)
	}

	// String + Number
	result, err, _ = parseAndEval(`"value: " + 42`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != "value: 42" {
		t.Errorf("expected 'value: 42', got %v", result)
	}
}

func TestVariables(t *testing.T) {
	result, err, _ := parseAndEval(`let x = 10
let y = 20
x + y`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != float64(30) {
		t.Errorf("expected 30, got %v", result)
	}
}

func TestAssignment(t *testing.T) {
	result, err, _ := parseAndEval(`let x = 10
x = 20
x`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != float64(20) {
		t.Errorf("expected 20, got %v", result)
	}
}

func TestIfStatement(t *testing.T) {
	result, err, _ := parseAndEval(`let x = 10
if (x > 5) {
    x = 100
}
x`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != float64(100) {
		t.Errorf("expected 100, got %v", result)
	}
}

func TestIfElse(t *testing.T) {
	result, err, _ := parseAndEval(`let x = 3
if (x > 5) {
    x = 100
} else {
    x = 50
}
x`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != float64(50) {
		t.Errorf("expected 50, got %v", result)
	}
}

func TestReturnStatement(t *testing.T) {
	result, err, _ := parseAndEval(`let x = 10
return x * 2
let y = 100
y`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != float64(20) {
		t.Errorf("expected 20, got %v", result)
	}
}

func TestReturnWithoutValue(t *testing.T) {
	result, err, _ := parseAndEval(`return`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestReturnInBlock(t *testing.T) {
	result, err, _ := parseAndEval(`if (true) {
    return "from block"
}
"after"`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != "from block" {
		t.Errorf("expected 'from block', got %v", result)
	}
}

func TestElvisOperator(t *testing.T) {
	result, err, _ := parseAndEval(`let x = null
x ?: "default"`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != "default" {
		t.Errorf("expected 'default', got %v", result)
	}

	result, err, _ = parseAndEval(`let x = "value"
x ?: "default"`, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != "value" {
		t.Errorf("expected 'value', got %v", result)
	}
}

func TestPropertyAccess(t *testing.T) {
	vars := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"age":  float64(30),
		},
	}

	result, err, _ := parseAndEval(`user.name`, vars)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != "Alice" {
		t.Errorf("expected 'Alice', got %v", result)
	}
}

func TestSafeAccess(t *testing.T) {
	vars := map[string]interface{}{
		"user": nil,
	}

	result, err, _ := parseAndEval(`user?.name`, vars)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestPrintCapture(t *testing.T) {
	l := lexer.New(`print("hello")
print("world")`)
	p := parser.New(l)
	program := p.ParseProgram()

	interp := New()
	interp.Eval(program)

	output := interp.GetOutput()
	if len(output) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(output))
	}
	if output[0] != "hello" || output[1] != "world" {
		t.Errorf("unexpected output: %v", output)
	}
}

func TestEnvironment(t *testing.T) {
	env := NewEnvironment()
	env.Set("x", float64(10))

	val, ok := env.Get("x")
	if !ok || val != float64(10) {
		t.Errorf("expected 10, got %v", val)
	}

	_, ok = env.Get("unknown")
	if ok {
		t.Error("expected false for unknown variable")
	}

	// Test enclosed environment
	inner := NewEnclosedEnvironment(env)
	inner.Set("y", float64(20))

	// Inner can access outer
	val, ok = inner.Get("x")
	if !ok || val != float64(10) {
		t.Errorf("inner should access outer: %v", val)
	}

	// Outer cannot access inner
	_, ok = env.Get("y")
	if ok {
		t.Error("outer should not access inner")
	}
}

func TestOutputCapture(t *testing.T) {
	env := NewEnvironment()
	env.AddOutput("line1")
	env.AddOutput("line2")

	output := env.GetOutput()
	if len(output) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(output))
	}
}

func TestUndefinedVariable(t *testing.T) {
	_, err, _ := parseAndEval(`undefined_var`, nil)
	if err == nil {
		t.Error("expected error for undefined variable")
	}
}

func TestDivisionByZero(t *testing.T) {
	_, err, _ := parseAndEval(`10 / 0`, nil)
	if err == nil {
		t.Error("expected error for division by zero")
	}
}

func TestNullPropertyAccess(t *testing.T) {
	_, err, _ := parseAndEval(`let x = null
x.property`, nil)
	if err == nil {
		t.Error("expected error for property access on null")
	}
}

func TestReturnValue(t *testing.T) {
	rv := &ReturnValue{Value: float64(42)}
	if rv.Value != float64(42) {
		t.Errorf("expected 42, got %v", rv.Value)
	}
}

func TestIsTruthy(t *testing.T) {
	if isTruthy(nil) != false {
		t.Error("nil should be falsy")
	}
	if isTruthy(false) != false {
		t.Error("false should be falsy")
	}
	if isTruthy(true) != true {
		t.Error("true should be truthy")
	}
	if isTruthy("") != true {
		t.Error("empty string should be truthy")
	}
	if isTruthy(float64(0)) != true {
		t.Error("0 should be truthy")
	}
}

func TestToNumber(t *testing.T) {
	n, ok := toNumber(float64(42))
	if !ok || n != 42 {
		t.Errorf("expected 42, got %v", n)
	}

	n, ok = toNumber(int(10))
	if !ok || n != 10 {
		t.Errorf("expected 10, got %v", n)
	}

	n, ok = toNumber(int64(5))
	if !ok || n != 5 {
		t.Errorf("expected 5, got %v", n)
	}

	_, ok = toNumber("not a number")
	if ok {
		t.Error("expected false for string")
	}
}

func TestUnknownStatementType(t *testing.T) {
	interp := New()

	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	_, err := interp.Eval(program)
	if err != nil {
		t.Errorf("empty program should not error: %v", err)
	}
}
