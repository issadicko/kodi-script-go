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

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected interface{}
	}{
		{
			name:     "basic return",
			source:   `return 42`,
			expected: float64(42),
		},
		{
			name:     "return expression",
			source:   `return 10 + 20`,
			expected: float64(30),
		},
		{
			name:     "return string",
			source:   `return "hello"`,
			expected: "hello",
		},
		{
			name:     "return null",
			source:   `return null`,
			expected: nil,
		},
		{
			name:     "return without value",
			source:   `return`,
			expected: nil,
		},
		{
			name: "early exit",
			source: `let x = 10
return x * 2
let y = 100
y`,
			expected: float64(20),
		},
		{
			name: "return in if block",
			source: `let x = 5
if (x > 3) {
    return "big"
}
return "small"`,
			expected: "big",
		},
		{
			name: "return in else block",
			source: `let x = 1
if (x > 3) {
    return "big"
} else {
    return "small"
}`,
			expected: "small",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Run(tt.source, nil)
			if len(result.Errors) > 0 {
				t.Fatalf("unexpected errors: %v", result.Errors)
			}
			if result.Value != tt.expected {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result.Value, result.Value)
			}
		})
	}
}

func TestReturnStopsExecution(t *testing.T) {
	result := Run(`
let x = 1
print("before return")
return x
print("after return")
`, nil)

	if len(result.Errors) > 0 {
		t.Fatalf("unexpected errors: %v", result.Errors)
	}
	if result.Value != float64(1) {
		t.Errorf("expected 1, got %v", result.Value)
	}
	if len(result.Output) != 1 {
		t.Errorf("expected 1 output line, got %d: %v", len(result.Output), result.Output)
	}
	if result.Output[0] != "before return" {
		t.Errorf("expected 'before return', got %v", result.Output[0])
	}
}

func TestMathFunctions(t *testing.T) {
	tests := []struct {
		source   string
		expected float64
	}{
		{`abs(-5)`, 5},
		{`abs(5)`, 5},
		{`floor(3.7)`, 3},
		{`ceil(3.2)`, 4},
		{`round(3.5)`, 4},
		{`round(3.4)`, 3},
		{`min(5, 3, 8, 1)`, 1},
		{`max(5, 3, 8, 1)`, 8},
		{`pow(2, 3)`, 8},
		{`sqrt(16)`, 4},
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

func TestStringFunctions(t *testing.T) {
	tests := []struct {
		source   string
		expected interface{}
	}{
		{`toUpperCase("hello")`, "HELLO"},
		{`toLowerCase("HELLO")`, "hello"},
		{`trim("  hello  ")`, "hello"},
		{`replace("hello world", "world", "kodi")`, "hello kodi"},
		{`contains("hello world", "world")`, true},
		{`contains("hello world", "foo")`, false},
		{`startsWith("hello world", "hello")`, true},
		{`endsWith("hello world", "world")`, true},
		{`indexOf("hello world", "world")`, float64(6)},
		{`indexOf("hello world", "foo")`, float64(-1)},
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

func TestCryptoFunctions(t *testing.T) {
	// Test md5
	result := Run(`md5("hello")`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("md5 errors: %v", result.Errors)
	}
	if result.Value != "5d41402abc4b2a76b9719d911017c592" {
		t.Errorf("md5: expected '5d41402abc4b2a76b9719d911017c592', got %v", result.Value)
	}

	// Test sha1
	result = Run(`sha1("hello")`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("sha1 errors: %v", result.Errors)
	}
	if result.Value != "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d" {
		t.Errorf("sha1: expected 'aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d', got %v", result.Value)
	}

	// Test sha256
	result = Run(`sha256("hello")`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("sha256 errors: %v", result.Errors)
	}
	if result.Value != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Errorf("sha256: expected '2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824', got %v", result.Value)
	}
}

func TestRandomFunctions(t *testing.T) {
	// Test random returns a float between 0 and 1
	result := Run(`random()`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("random errors: %v", result.Errors)
	}
	if v, ok := result.Value.(float64); !ok || v < 0 || v >= 1 {
		t.Errorf("random: expected float in [0,1), got %v", result.Value)
	}

	// Test randomInt returns an integer in range
	result = Run(`randomInt(1, 10)`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("randomInt errors: %v", result.Errors)
	}
	if v, ok := result.Value.(float64); !ok || v < 1 || v > 10 {
		t.Errorf("randomInt: expected int in [1,10], got %v", result.Value)
	}

	// Test randomUUID returns a valid-looking UUID
	result = Run(`randomUUID()`, nil)
	if len(result.Errors) > 0 {
		t.Fatalf("randomUUID errors: %v", result.Errors)
	}
	if s, ok := result.Value.(string); !ok || len(s) != 36 {
		t.Errorf("randomUUID: expected 36-char string, got %v", result.Value)
	}
}

func TestTypeCheckFunctions(t *testing.T) {
	tests := []struct {
		source   string
		expected interface{}
	}{
		{`isNumber(42)`, true},
		{`isNumber("42")`, false},
		{`isString("hello")`, true},
		{`isString(42)`, false},
		{`isBool(true)`, true},
		{`isBool("true")`, false},
		{`typeOf(42)`, "number"},
		{`typeOf("hello")`, "string"},
		{`typeOf(true)`, "boolean"},
		{`typeOf(null)`, "null"},
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

func TestArraySortFunctions(t *testing.T) {
	// Test sort on numbers (from JSON)
	vars := map[string]interface{}{
		"numbers": []interface{}{float64(3), float64(1), float64(4), float64(1), float64(5)},
		"users": []interface{}{
			map[string]interface{}{"name": "Charlie", "age": float64(30)},
			map[string]interface{}{"name": "Alice", "age": float64(25)},
			map[string]interface{}{"name": "Bob", "age": float64(35)},
		},
	}

	// Test sort ascending
	result := Run(`sort(numbers)`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("sort errors: %v", result.Errors)
	}
	sorted := result.Value.([]interface{})
	if sorted[0] != float64(1) || sorted[len(sorted)-1] != float64(5) {
		t.Errorf("sort asc failed: %v", sorted)
	}

	// Test sort descending
	result = Run(`sort(numbers, "desc")`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("sort desc errors: %v", result.Errors)
	}
	sorted = result.Value.([]interface{})
	if sorted[0] != float64(5) || sorted[len(sorted)-1] != float64(1) {
		t.Errorf("sort desc failed: %v", sorted)
	}

	// Test sortBy on objects
	result = Run(`sortBy(users, "age")`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("sortBy errors: %v", result.Errors)
	}
	sortedUsers := result.Value.([]interface{})
	firstUser := sortedUsers[0].(map[string]interface{})
	if firstUser["name"] != "Alice" {
		t.Errorf("sortBy age asc failed, first should be Alice: %v", firstUser["name"])
	}

	// Test sortBy descending
	result = Run(`sortBy(users, "age", "desc")`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("sortBy desc errors: %v", result.Errors)
	}
	sortedUsers = result.Value.([]interface{})
	firstUser = sortedUsers[0].(map[string]interface{})
	if firstUser["name"] != "Bob" {
		t.Errorf("sortBy age desc failed, first should be Bob: %v", firstUser["name"])
	}

	// Test sortBy on string field
	result = Run(`sortBy(users, "name")`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("sortBy name errors: %v", result.Errors)
	}
	sortedUsers = result.Value.([]interface{})
	firstUser = sortedUsers[0].(map[string]interface{})
	if firstUser["name"] != "Alice" {
		t.Errorf("sortBy name failed, first should be Alice: %v", firstUser["name"])
	}
}

func TestArrayUtilityFunctions(t *testing.T) {
	vars := map[string]interface{}{
		"arr": []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)},
	}

	// Test size
	result := Run(`size(arr)`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("size errors: %v", result.Errors)
	}
	if result.Value != float64(5) {
		t.Errorf("size: expected 5, got %v", result.Value)
	}

	// Test first
	result = Run(`first(arr)`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("first errors: %v", result.Errors)
	}
	if result.Value != float64(1) {
		t.Errorf("first: expected 1, got %v", result.Value)
	}

	// Test last
	result = Run(`last(arr)`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("last errors: %v", result.Errors)
	}
	if result.Value != float64(5) {
		t.Errorf("last: expected 5, got %v", result.Value)
	}

	// Test reverse
	result = Run(`reverse(arr)`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("reverse errors: %v", result.Errors)
	}
	reversed := result.Value.([]interface{})
	if reversed[0] != float64(5) || reversed[4] != float64(1) {
		t.Errorf("reverse failed: %v", reversed)
	}

	// Test slice
	result = Run(`slice(arr, 1, 3)`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("slice errors: %v", result.Errors)
	}
	sliced := result.Value.([]interface{})
	if len(sliced) != 2 || sliced[0] != float64(2) || sliced[1] != float64(3) {
		t.Errorf("slice failed: %v", sliced)
	}
}

func TestForLoop(t *testing.T) {
	// Test basic for loop with sum
	vars := map[string]interface{}{
		"numbers": []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)},
	}

	result := Run(`
		let sum = 0
		for (n in numbers) {
			sum = sum + n
		}
		sum
	`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("for loop errors: %v", result.Errors)
	}
	if result.Value != float64(15) {
		t.Errorf("expected 15, got %v", result.Value)
	}

	// Test for loop with print
	result = Run(`
		for (item in numbers) {
			print(item)
		}
	`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("for loop print errors: %v", result.Errors)
	}
	if len(result.Output) != 5 {
		t.Errorf("expected 5 outputs, got %d", len(result.Output))
	}

	// Test for loop with objects
	users := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice"},
			map[string]interface{}{"name": "Bob"},
		},
	}
	result = Run(`
		for (user in users) {
			print(user.name)
		}
	`, users)
	if len(result.Errors) > 0 {
		t.Fatalf("for loop objects errors: %v", result.Errors)
	}
	if len(result.Output) != 2 || result.Output[0] != "Alice" {
		t.Errorf("expected Alice/Bob, got %v", result.Output)
	}

	// Test early return in for loop (simpler case)
	result = Run(`
let found = "no"
for (n in numbers) {
    found = n
}
return found
`, vars)
	if len(result.Errors) > 0 {
		t.Fatalf("for loop return errors: %v", result.Errors)
	}
	if result.Value != float64(5) {
		t.Errorf("expected 5, got %v", result.Value)
	}
}
