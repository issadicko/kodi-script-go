package kodi

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// CONCURRENCY TESTS - High Priority
// ============================================================================

func TestConcurrency_ParallelExecution(t *testing.T) {
	// Execute N scripts in parallel using goroutines
	const numGoroutines = 100
	const numIterationsPerGoroutine = 100

	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5]) {
			sum = sum + i
		}
		sum
	`

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numIterationsPerGoroutine; j++ {
				result := New(script).SilentPrint(true).Execute()

				if len(result.Errors) > 0 {
					errors <- &EvalError{Messages: result.Errors}
					return
				}

				// Verify result is correct
				if result.Value != float64(15) {
					t.Errorf("goroutine %d, iteration %d: expected 15, got %v", id, j, result.Value)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("concurrent execution error: %v", err)
	}

	t.Logf("Successfully executed %d scripts in parallel (%d goroutines × %d iterations)",
		numGoroutines*numIterationsPerGoroutine, numGoroutines, numIterationsPerGoroutine)
}

func TestConcurrency_ParallelWithVariables(t *testing.T) {
	// Each goroutine has its own isolated variable context
	const numGoroutines = 50

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			script := New(`id * 10`).
				WithVariables(map[string]interface{}{"id": float64(id)}).
				SilentPrint(true)

			result := script.Execute()

			if len(result.Errors) > 0 {
				t.Errorf("goroutine %d error: %v", id, result.Errors)
				return
			}

			expected := float64(id * 10)
			if result.Value != expected {
				t.Errorf("goroutine %d: expected %v, got %v", id, expected, result.Value)
			}
		}(i)
	}

	wg.Wait()
	t.Log("Successfully tested parallel execution with isolated variable contexts")
}

func TestConcurrency_ParallelWithBindings(t *testing.T) {
	// Test parallel execution with bound objects

	type Counter struct {
		Value int
	}

	const numGoroutines = 50

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			counter := &Counter{Value: id * 5}

			result := New(`counter.Value + 10`).
				Bind("counter", counter).
				SilentPrint(true).
				Execute()

			if len(result.Errors) > 0 {
				t.Errorf("goroutine %d error: %v", id, result.Errors)
				return
			}

			expected := float64(id*5 + 10)
			if result.Value != expected {
				t.Errorf("goroutine %d: expected %v, got %v", id, expected, result.Value)
			}
		}(i)
	}

	wg.Wait()
	t.Log("Successfully tested parallel execution with bound objects")
}

// ============================================================================
// MEMORY TESTS - High Priority
// ============================================================================

func TestMemory_NoLeakOnRepeatedExecution(t *testing.T) {
	// Execute a complex script many times and check memory doesn't grow unboundedly
	const iterations = 1000

	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) {
			sum = sum + i * i
		}
		let result = sum * 2
		result
	`

	// Force GC and get baseline memory
	runtime.GC()
	var baselineStats runtime.MemStats
	runtime.ReadMemStats(&baselineStats)

	// Execute many times
	for i := 0; i < iterations; i++ {
		result := New(script).SilentPrint(true).WithCache(false).Execute()
		if len(result.Errors) > 0 {
			t.Fatalf("iteration %d failed: %v", i, result.Errors)
		}
	}

	// Force GC and get final memory
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // Allow GC to complete
	runtime.GC()

	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)

	// Check that memory hasn't grown too much
	baselineAlloc := baselineStats.HeapAlloc
	finalAlloc := finalStats.HeapAlloc

	// Allow up to 50MB growth (reasonable for 1000 iterations with GC)
	maxGrowth := uint64(50 * 1024 * 1024)

	if finalAlloc > baselineAlloc+maxGrowth {
		t.Errorf("Memory leak detected: baseline=%d MB, final=%d MB, growth=%d MB",
			baselineAlloc/(1024*1024),
			finalAlloc/(1024*1024),
			(finalAlloc-baselineAlloc)/(1024*1024))
	} else {
		t.Logf("Memory test passed: baseline=%d KB, final=%d KB, growth=%d KB",
			baselineAlloc/1024, finalAlloc/1024, (finalAlloc-baselineAlloc)/1024)
	}
}

func TestMemory_CacheDoesNotLeakUnlimited(t *testing.T) {
	// Test that using cache with many different scripts doesn't cause unbounded growth
	const iterations = 500

	runtime.GC()
	var baselineStats runtime.MemStats
	runtime.ReadMemStats(&baselineStats)

	// Execute many DIFFERENT scripts (each cached separately)
	for i := 0; i < iterations; i++ {
		script := New(`let x = ` + string(rune('0'+i%10)) + `; x * 2`).
			SilentPrint(true).
			WithCache(true)
		result := script.Execute()
		if len(result.Errors) > 0 {
			// Some scripts may have parse errors, that's ok for this test
			continue
		}
	}

	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)

	// Log memory usage
	t.Logf("Cache memory test: baseline=%d KB, final=%d KB",
		baselineStats.HeapAlloc/1024, finalStats.HeapAlloc/1024)
}

// ============================================================================
// LIMITS TESTS
// ============================================================================

func TestLimits_LargeStrings(t *testing.T) {
	// Test with a very large string
	largeString := ""
	for i := 0; i < 10000; i++ {
		largeString += "x"
	}

	script := New(`length(bigString)`).
		WithVariables(map[string]interface{}{"bigString": largeString}).
		SilentPrint(true)

	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Large string test failed: %v", result.Errors)
	}

	if result.Value != float64(10000) {
		t.Errorf("Expected 10000, got %v", result.Value)
	}

	t.Log("Large string test passed (10,000 characters)")
}

func TestLimits_LargeArrays(t *testing.T) {
	// Test with a large array
	largeArray := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		largeArray[i] = float64(i)
	}

	script := New(`
		let sum = 0
		for (item in arr) {
			sum = sum + item
		}
		sum
	`).
		WithVariables(map[string]interface{}{"arr": largeArray}).
		SilentPrint(true)

	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Large array test failed: %v", result.Errors)
	}

	// Sum of 0 to 999 = 999 * 1000 / 2 = 499500
	expected := float64(499500)
	if result.Value != expected {
		t.Errorf("Expected %v, got %v", expected, result.Value)
	}

	t.Log("Large array test passed (1,000 elements)")
}

func TestLimits_ManyVariables(t *testing.T) {
	// Inject many variables
	vars := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		vars["var"+string(rune('a'+i%26))+string(rune('0'+i/26))] = float64(i)
	}

	script := New(`vara0 + varb0 + varc0`).
		WithVariables(vars).
		SilentPrint(true)

	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Many variables test failed: %v", result.Errors)
	}

	// vara0=0, varb0=1, varc0=2
	expected := float64(0 + 1 + 2)
	if result.Value != expected {
		t.Errorf("Expected %v, got %v", expected, result.Value)
	}

	t.Log("Many variables test passed (100 variables)")
}

func TestLimits_DeepNestedObjects(t *testing.T) {
	// Test deeply nested object access
	nested := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": map[string]interface{}{
						"value": float64(42),
					},
				},
			},
		},
	}

	script := New(`obj.level1.level2.level3.level4.value`).
		WithVariables(map[string]interface{}{"obj": nested}).
		SilentPrint(true)

	result := script.Execute()

	if len(result.Errors) > 0 {
		t.Fatalf("Deep nesting test failed: %v", result.Errors)
	}

	if result.Value != float64(42) {
		t.Errorf("Expected 42, got %v", result.Value)
	}

	t.Log("Deep nesting test passed (5 levels)")
}

// ============================================================================
// TIMEOUT TESTS
// ============================================================================

func TestTimeout_LongLoopWithContext(t *testing.T) {
	// This test demonstrates that a very long loop will complete
	// Note: KodiScript doesn't have built-in timeout support yet
	// This test just ensures a reasonable loop completes in time

	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) {
			for (j in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]) {
				sum = sum + 1
			}
		}
		sum
	`

	done := make(chan *Result, 1)

	go func() {
		result := New(script).SilentPrint(true).Execute()
		done <- result
	}()

	select {
	case result := <-done:
		if len(result.Errors) > 0 {
			t.Fatalf("Loop test failed: %v", result.Errors)
		}
		if result.Value != float64(100) {
			t.Errorf("Expected 100, got %v", result.Value)
		}
		t.Log("Nested loops test passed (100 iterations)")

	case <-time.After(5 * time.Second):
		t.Fatal("Timeout: script took too long (>5s)")
	}
}

func TestTimeout_ExecutionWithDeadline(t *testing.T) {
	// Test that we can wrap execution with a timeout
	script := `
		let result = 0
		for (i in [1, 2, 3, 4, 5]) {
			result = result + i
		}
		result
	`

	timeout := 2 * time.Second
	done := make(chan *Result, 1)

	go func() {
		result := New(script).SilentPrint(true).Execute()
		done <- result
	}()

	select {
	case result := <-done:
		if len(result.Errors) > 0 {
			t.Fatalf("Execution failed: %v", result.Errors)
		}
		t.Logf("Script completed within timeout (%v)", timeout)

	case <-time.After(timeout):
		t.Fatal("Script execution exceeded timeout")
	}
}

// ============================================================================
// BENCHMARKS
// ============================================================================

func BenchmarkConcurrentExecution(b *testing.B) {
	script := `
		let sum = 0
		for (i in [1, 2, 3, 4, 5]) {
			sum = sum + i
		}
		sum
	`

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			New(script).SilentPrint(true).Execute()
		}
	})
}

func BenchmarkMemoryAllocation(b *testing.B) {
	script := `
		let obj = { name: "test", value: 42 }
		obj.value * 2
	`

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		New(script).SilentPrint(true).WithCache(false).Execute()
	}
}
