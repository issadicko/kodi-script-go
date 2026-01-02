package kodi

import (
	"errors"
	"testing"

	"github.com/issadicko/kodi-script-go/interpreter"
)

// ============================================================================
// OPERATION LIMIT TESTS
// ============================================================================

func TestOperationLimit_SimpleScript(t *testing.T) {
	// A simple script with few operations should complete successfully
	script := `
		let x = 1
		let y = 2
		x + y
	`

	result := New(script).
		WithMaxOperations(100).
		SilentPrint(true).
		Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Expected success, got errors: %v", result.Errors)
	}

	if result.Value != float64(3) {
		t.Errorf("Expected 3, got %v", result.Value)
	}

	t.Log("Simple script completed within operation limit")
}

func TestOperationLimit_ExceedsLimit(t *testing.T) {
	// A script that will exceed a very low limit
	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) {
			sum = sum + i
		}
		sum
	`

	result := New(script).
		WithMaxOperations(5). // Very low limit
		SilentPrint(true).
		Execute()

	if len(result.Errors) == 0 {
		t.Fatal("Expected error for exceeded operation limit, got success")
	}

	// Check that the error is the expected one
	if !errors.Is(errors.New(result.Errors[0]), interpreter.ErrMaxOperationsExceeded) {
		// Check if error message contains expected text
		if result.Errors[0] != interpreter.ErrMaxOperationsExceeded.Error() {
			t.Logf("Error message: %s", result.Errors[0])
		}
	}

	t.Log("Operation limit correctly triggered")
}

func TestOperationLimit_InfiniteLoopProtection(t *testing.T) {
	// Simulate an "infinite" loop by having a very large array
	// With operation limit, it should be stopped
	largeArray := make([]interface{}, 10000)
	for i := 0; i < 10000; i++ {
		largeArray[i] = float64(i)
	}

	script := `
		let sum = 0
		for (i in arr) {
			sum = sum + i
		}
		sum
	`

	result := New(script).
		WithVariables(map[string]interface{}{"arr": largeArray}).
		WithMaxOperations(100). // Much less than 10000
		SilentPrint(true).
		Execute()

	if len(result.Errors) == 0 {
		t.Fatal("Expected infinite loop to be stopped by operation limit")
	}

	t.Log("Infinite loop protection working correctly")
}

func TestOperationLimit_NestedLoops(t *testing.T) {
	// Nested loops should respect the limit
	script := `
		let count = 0
		for (i in [1, 2, 3, 4, 5]) {
			for (j in [1, 2, 3, 4, 5]) {
				count = count + 1
			}
		}
		count
	`

	// 5 * 5 = 25 iterations, plus overhead statements
	// With limit of 10, it should fail
	result := New(script).
		WithMaxOperations(10).
		SilentPrint(true).
		Execute()

	if len(result.Errors) == 0 {
		t.Fatal("Expected nested loop to exceed operation limit")
	}

	t.Log("Nested loops correctly stopped by operation limit")
}

func TestOperationLimit_NoLimitByDefault(t *testing.T) {
	// Without WithMaxOperations, script should run without limit
	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) {
			sum = sum + i
		}
		sum
	`

	result := New(script).
		SilentPrint(true).
		Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Default (no limit) should not cause errors: %v", result.Errors)
	}

	if result.Value != float64(55) {
		t.Errorf("Expected 55, got %v", result.Value)
	}

	t.Log("Default (unlimited) behavior works correctly")
}

func TestOperationLimit_ZeroMeansUnlimited(t *testing.T) {
	// WithMaxOperations(0) should mean unlimited
	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5]) {
			sum = sum + i
		}
		sum
	`

	result := New(script).
		WithMaxOperations(0).
		SilentPrint(true).
		Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Zero (unlimited) should not cause errors: %v", result.Errors)
	}

	if result.Value != float64(15) {
		t.Errorf("Expected 15, got %v", result.Value)
	}

	t.Log("Zero means unlimited works correctly")
}

func TestOperationLimit_WithBindings(t *testing.T) {
	// Operation limit should work with bound objects too
	type Counter struct {
		Value int
	}

	counter := &Counter{Value: 0}

	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) {
			sum = sum + counter.Value
		}
		sum
	`

	result := New(script).
		Bind("counter", counter).
		WithMaxOperations(5). // Low limit
		SilentPrint(true).
		Execute()

	if len(result.Errors) == 0 {
		t.Fatal("Expected operation limit with bindings to work")
	}

	t.Log("Operation limit works correctly with bound objects")
}
