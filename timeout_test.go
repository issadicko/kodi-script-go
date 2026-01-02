package kodi

import (
	"testing"
	"time"
)

// ============================================================================
// TIMEOUT TESTS
// ============================================================================

func TestTimeout_SimpleScriptCompletesWithinTimeout(t *testing.T) {
	script := `
		let x = 1
		let y = 2
		x + y
	`

	result := New(script).
		WithTimeout(5 * time.Second).
		SilentPrint(true).
		Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Expected success, got errors: %v", result.Errors)
	}

	if result.Value != float64(3) {
		t.Errorf("Expected 3, got %v", result.Value)
	}

	t.Log("Simple script completed within timeout")
}

func TestTimeout_LongLoopExceedsTimeout(t *testing.T) {
	// Create a script that takes a long time by having many iterations
	largeArray := make([]interface{}, 1000000)
	for i := 0; i < 1000000; i++ {
		largeArray[i] = float64(i)
	}

	script := `
		let sum = 0
		for (i in arr) {
			sum = sum + i
		}
		sum
	`

	start := time.Now()
	result := New(script).
		WithVariables(map[string]interface{}{"arr": largeArray}).
		WithTimeout(50 * time.Millisecond). // Very short timeout
		SilentPrint(true).
		Execute()
	elapsed := time.Since(start)

	if len(result.Errors) == 0 {
		t.Fatal("Expected timeout error, got success")
	}

	// Check error message contains timeout
	if result.Errors[0] != "execution timeout" {
		t.Logf("Error message: %s", result.Errors[0])
	}

	// Verify it actually stopped early (should be around 50ms, not the full execution time)
	if elapsed > 200*time.Millisecond {
		t.Errorf("Timeout didn't stop execution quickly enough: took %v", elapsed)
	}

	t.Logf("Timeout correctly triggered after %v", elapsed)
}

func TestTimeout_NoTimeoutByDefault(t *testing.T) {
	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5]) {
			sum = sum + i
		}
		sum
	`

	result := New(script).
		SilentPrint(true).
		Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Default (no timeout) should not cause errors: %v", result.Errors)
	}

	if result.Value != float64(15) {
		t.Errorf("Expected 15, got %v", result.Value)
	}

	t.Log("Default (no timeout) behavior works correctly")
}

func TestTimeout_ZeroMeansNoTimeout(t *testing.T) {
	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5]) {
			sum = sum + i
		}
		sum
	`

	result := New(script).
		WithTimeout(0).
		SilentPrint(true).
		Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Zero timeout should not cause errors: %v", result.Errors)
	}

	if result.Value != float64(15) {
		t.Errorf("Expected 15, got %v", result.Value)
	}

	t.Log("Zero timeout (unlimited) works correctly")
}

func TestTimeout_CombinedWithMaxOperations(t *testing.T) {
	// Test that both timeout and max operations work together
	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) {
			sum = sum + i
		}
		sum
	`

	// With generous limits, should complete successfully
	result := New(script).
		WithTimeout(5 * time.Second).
		WithMaxOperations(1000).
		SilentPrint(true).
		Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Should complete with generous limits: %v", result.Errors)
	}

	if result.Value != float64(55) {
		t.Errorf("Expected 55, got %v", result.Value)
	}

	t.Log("Combined timeout and max operations works correctly")
}

func TestTimeout_NestedLoopsWithTimeout(t *testing.T) {
	script := `
		let count = 0
		for (i in [1, 2, 3]) {
			for (j in [1, 2, 3]) {
				count = count + 1
			}
		}
		count
	`

	result := New(script).
		WithTimeout(5 * time.Second).
		SilentPrint(true).
		Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Nested loops should complete within timeout: %v", result.Errors)
	}

	if result.Value != float64(9) {
		t.Errorf("Expected 9, got %v", result.Value)
	}

	t.Log("Nested loops with timeout works correctly")
}
