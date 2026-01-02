package kodi

import (
	"strings"
	"testing"
)

// Test structs for reflective binding
type Address struct {
	City    string
	Country string
}

type User struct {
	Name    string
	Age     int
	Address Address
}

func (u *User) SayHello() string {
	return "Hello, I'm " + u.Name
}

func (u *User) GetAge() int {
	return u.Age
}

func (u *User) Greet(greeting string) string {
	return greeting + ", " + u.Name + "!"
}

func (u *User) GetAddress() Address {
	return u.Address
}

type Calculator struct{}

func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}

func (c *Calculator) Multiply(x, y int) int {
	return x * y
}

func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, nil // KodiScript doesn't support error returns well yet
	}
	return a / b, nil
}

// TestBindFieldAccess tests accessing fields on bound objects
func TestBindFieldAccess(t *testing.T) {
	user := &User{
		Name: "Alice",
		Age:  30,
		Address: Address{
			City:    "Paris",
			Country: "France",
		},
	}

	script := New("user.Name").Bind("user", user)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	if result.Value != "Alice" {
		t.Errorf("Expected 'Alice', got %v", result.Value)
	}
}

// TestBindMethodCall tests calling methods on bound objects
func TestBindMethodCall(t *testing.T) {
	user := &User{Name: "Bob", Age: 25}

	script := New("user.SayHello()").Bind("user", user)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	if result.Value != "Hello, I'm Bob" {
		t.Errorf("Expected 'Hello, I'm Bob', got %v", result.Value)
	}
}

// TestBindMethodWithArgs tests calling methods with arguments
func TestBindMethodWithArgs(t *testing.T) {
	user := &User{Name: "Charlie"}

	script := New(`user.Greet("Hi")`).Bind("user", user)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	if result.Value != "Hi, Charlie!" {
		t.Errorf("Expected 'Hi, Charlie!', got %v", result.Value)
	}
}

// TestBindNestedObjects tests accessing nested objects
func TestBindNestedObjects(t *testing.T) {
	user := &User{
		Name: "David",
		Address: Address{
			City:    "London",
			Country: "UK",
		},
	}

	script := New("user.Address.City").Bind("user", user)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	if result.Value != "London" {
		t.Errorf("Expected 'London', got %v", result.Value)
	}
}

// TestBindMethodChaining tests chaining method calls
func TestBindMethodChaining(t *testing.T) {
	user := &User{
		Name: "Emily",
		Address: Address{
			City:    "Tokyo",
			Country: "Japan",
		},
	}

	script := New("user.GetAddress().City").Bind("user", user)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	if result.Value != "Tokyo" {
		t.Errorf("Expected 'Tokyo', got %v", result.Value)
	}
}

// TestBindNumericConversion tests automatic conversion of numeric types
func TestBindNumericConversion(t *testing.T) {
	calc := &Calculator{}

	script := New("calc.Add(10, 20)").Bind("calc", calc)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	if result.Value != 30.0 {
		t.Errorf("Expected 30.0, got %v", result.Value)
	}
}

// TestBindIntReturn tests that ints are converted to float64
func TestBindIntReturn(t *testing.T) {
	user := &User{Age: 42}

	script := New("user.GetAge()").Bind("user", user)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	// KodiScript uses float64 for numbers
	if result.Value != 42.0 {
		t.Errorf("Expected 42.0, got %v", result.Value)
	}
}

// TestBindMultipleObjects tests binding multiple objects
func TestBindMultipleObjects(t *testing.T) {
	user := &User{Name: "Frank"}
	calc := &Calculator{}

	source := `
		let greeting = user.SayHello()
		let sum = calc.Add(5, 3)
		greeting + " " + sum
	`

	script := New(source).
		Bind("user", user).
		Bind("calc", calc)

	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	expected := "Hello, I'm Frank 8"
	if result.Value != expected {
		t.Errorf("Expected '%s', got %v", expected, result.Value)
	}
}

// TestBindComplexScript tests a more complex script using bound objects
func TestBindComplexScript(t *testing.T) {
	user := &User{
		Name: "Grace",
		Age:  28,
		Address: Address{
			City:    "Berlin",
			Country: "Germany",
		},
	}

	source := `
		let greeting = user.SayHello()
		let age = user.GetAge()
		let city = user.Address.City
		
		greeting + " I am " + age + " years old and I live in " + city
	`

	script := New(source).Bind("user", user)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	expected := "Hello, I'm Grace I am 28 years old and I live in Berlin"
	if result.Value != expected {
		t.Errorf("Expected '%s', got %v", expected, result.Value)
	}
}

// TestBindNonExistentProperty tests error handling for non-existent properties
func TestBindNonExistentProperty(t *testing.T) {
	user := &User{Name: "Henry"}

	script := New("user.NonExistent").Bind("user", user)
	result := script.Execute()

	if len(result.Errors) == 0 {
		t.Fatal("Expected error for non-existent property")
	}

	errorMsg := result.Errors[0]
	if !strings.Contains(errorMsg, "NonExistent") {
		t.Errorf("Expected error about NonExistent, got: %s", errorMsg)
	}
}

// TestBindWithVariables tests combining bound objects with script variables
func TestBindWithVariables(t *testing.T) {
	calc := &Calculator{}

	source := `
		let x = 10
		let y = 5
		calc.Add(x, y)
	`

	script := New(source).Bind("calc", calc)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	if result.Value != 15.0 {
		t.Errorf("Expected 15.0, got %v", result.Value)
	}
}

// TestBindInLoop tests using bound objects in loops
func TestBindInLoop(t *testing.T) {
	calc := &Calculator{}

	source := `
		let numbers = [1, 2, 3, 4, 5]
		let sum = 0
		for (n in numbers) {
			sum = calc.Add(sum, n)
		}
		sum
	`

	script := New(source).Bind("calc", calc)
	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected errors: %v", result.Errors)
	}

	if result.Value != 15.0 {
		t.Errorf("Expected 15.0, got %v", result.Value)
	}
}
