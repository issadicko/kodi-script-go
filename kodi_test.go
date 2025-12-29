package kodi

import (
	"testing"
)

func TestBasicVariableDeclaration(t *testing.T) {
	result := Run(`let x = 42`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Value != float64(42) {
		t.Errorf("expected 42, got %v", result.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	result := Run(`let name = "Kodi"
let greeting = "Hello " + name
greeting`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Value != "Hello Kodi" {
		t.Errorf("expected 'Hello Kodi', got %v", result.Value)
	}
}

func TestNullSafetyElvis(t *testing.T) {
	result := Run(`let x = null
let y = x ?: "default"
y`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Value != "default" {
		t.Errorf("expected 'default', got %v", result.Value)
	}
}

func TestHostVariables(t *testing.T) {
	vars := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Alice",
			"age":  30.0,
		},
	}
	result := Run(`user.name`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Value != "Alice" {
		t.Errorf("expected 'Alice', got %v", result.Value)
	}
}

func TestSafeAccess(t *testing.T) {
	vars := map[string]interface{}{
		"user": nil,
	}
	result := Run(`let status = user?.name ?: "unknown"
status`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Value != "unknown" {
		t.Errorf("expected 'unknown', got %v", result.Value)
	}
}

func TestIfStatement(t *testing.T) {
	result := Run(`let x = 10
let result = "none"
if (x > 5) {
    result = "big"
} else {
    result = "small"
}
result`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Value != "big" {
		t.Errorf("expected 'big', got %v", result.Value)
	}
}

func TestPrintOutput(t *testing.T) {
	result := Run(`print("Hello")
print("World")`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if len(result.Output) != 2 {
		t.Errorf("expected 2 output lines, got %d", len(result.Output))
	}
	if result.Output[0] != "Hello" || result.Output[1] != "World" {
		t.Errorf("unexpected output: %v", result.Output)
	}
}

func TestArithmeticOperations(t *testing.T) {
	tests := []struct {
		source   string
		expected float64
	}{
		{"5 + 3", 8},
		{"10 - 4", 6},
		{"3 * 4", 12},
		{"20 / 5", 4},
		{"(2 + 3) * 4", 20},
	}

	for _, tt := range tests {
		result := Run(tt.source, nil)
		if len(result.Errors) > 0 {
			t.Fatalf("'%s' errors: %v", tt.source, result.Errors)
		}
		if result.Value != tt.expected {
			t.Errorf("'%s': expected %v, got %v", tt.source, tt.expected, result.Value)
		}
	}
}

func TestBooleanLogic(t *testing.T) {
	tests := []struct {
		source   string
		expected bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false || true", true},
		{"!false", true},
		{"5 > 3", true},
		{"5 == 5", true},
		{"5 != 3", true},
	}

	for _, tt := range tests {
		result := Run(tt.source, nil)
		if len(result.Errors) > 0 {
			t.Fatalf("'%s' errors: %v", tt.source, result.Errors)
		}
		if result.Value != tt.expected {
			t.Errorf("'%s': expected %v, got %v", tt.source, tt.expected, result.Value)
		}
	}
}

func TestNativeFunctions(t *testing.T) {
	// Test base64
	result := Run(`base64Encode("hello")`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("base64Encode errors: %v", result.Errors)
	}
	if result.Value != "aGVsbG8=" {
		t.Errorf("expected 'aGVsbG8=', got %v", result.Value)
	}

	// Test jsonStringify
	result = Run(`jsonStringify("test")`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("jsonStringify errors: %v", result.Errors)
	}
	if result.Value != `"test"` {
		t.Errorf("expected '\"test\"', got %v", result.Value)
	}
}

func TestMultilineExpression(t *testing.T) {
	result := Run(`let total = 10 +
20 +
30
total`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Value != float64(60) {
		t.Errorf("expected 60, got %v", result.Value)
	}
}

func TestOptionalSemicolons(t *testing.T) {
	// Both with and without semicolons should work
	result1 := Run(`let x = 1; let y = 2; x + y`, nil)
	result2 := Run(`let x = 1
let y = 2
x + y`, nil)

	if len(result1.Errors) > 0 {
		t.Fatalf("with semicolons errors: %v", result1.Errors)
	}
	if len(result2.Errors) > 0 {
		t.Fatalf("without semicolons errors: %v", result2.Errors)
	}
	if result1.Value != result2.Value {
		t.Errorf("results differ: %v vs %v", result1.Value, result2.Value)
	}
}
