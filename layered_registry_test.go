package kodi

import (
	"testing"
)

// TestLayeredRegistry verifies the layered registry pattern:
// - Global builtins are shared across all scripts
// - Custom functions are isolated per script instance
func TestLayeredRegistry_CustomFunctionIsolation(t *testing.T) {
	// Script 1: Register a custom function "secret"
	script1 := New(`
		let result = secret()
		result
	`)
	script1.RegisterFunction("secret", func(args ...interface{}) (interface{}, error) {
		return "script1_secret", nil
	})

	result1 := script1.SilentPrint(true).Execute()
	if len(result1.Errors) > 0 {
		t.Fatalf("Script 1 failed: %v", result1.Errors)
	}
	if result1.Value != "script1_secret" {
		t.Errorf("Script 1: expected 'script1_secret', got %v", result1.Value)
	}

	// Script 2: Should NOT have access to script1's "secret" function
	script2 := New(`
		let result = secret()
		result
	`)
	// No RegisterNative call for "secret"

	result2 := script2.SilentPrint(true).Execute()
	// Should fail because "secret" is not defined
	if len(result2.Errors) == 0 {
		t.Fatal("Script 2 should have failed - 'secret' should not be accessible")
	}
	t.Logf("Script 2 correctly failed: %v", result2.Errors[0])
}

func TestLayeredRegistry_BuiltinsShared(t *testing.T) {
	// Both scripts should have access to built-in functions
	script1 := New(`toUpperCase("hello")`)
	script2 := New(`toUpperCase("world")`)

	result1 := script1.SilentPrint(true).Execute()
	result2 := script2.SilentPrint(true).Execute()

	if len(result1.Errors) > 0 {
		t.Fatalf("Script 1 failed: %v", result1.Errors)
	}
	if len(result2.Errors) > 0 {
		t.Fatalf("Script 2 failed: %v", result2.Errors)
	}

	if result1.Value != "HELLO" {
		t.Errorf("Script 1: expected 'HELLO', got %v", result1.Value)
	}
	if result2.Value != "WORLD" {
		t.Errorf("Script 2: expected 'WORLD', got %v", result2.Value)
	}
}

func TestLayeredRegistry_CustomOverridesBuiltin(t *testing.T) {
	// Custom function should override a built-in with the same name
	script := New(`toUpperCase("hello")`)
	script.RegisterFunction("toUpperCase", func(args ...interface{}) (interface{}, error) {
		return "CUSTOM_OVERRIDE", nil
	})

	result := script.SilentPrint(true).Execute()
	if len(result.Errors) > 0 {
		t.Fatalf("Script failed: %v", result.Errors)
	}

	if result.Value != "CUSTOM_OVERRIDE" {
		t.Errorf("Expected 'CUSTOM_OVERRIDE', got %v", result.Value)
	}
}

func TestLayeredRegistry_CustomOverrideDoesNotAffectOtherScripts(t *testing.T) {
	// Script 1: Override toUpperCase
	script1 := New(`toUpperCase("hello")`)
	script1.RegisterFunction("toUpperCase", func(args ...interface{}) (interface{}, error) {
		return "CUSTOM", nil
	})

	// Script 2: Should use the original built-in
	script2 := New(`toUpperCase("hello")`)

	result1 := script1.SilentPrint(true).Execute()
	result2 := script2.SilentPrint(true).Execute()

	if result1.Value != "CUSTOM" {
		t.Errorf("Script 1: expected 'CUSTOM', got %v", result1.Value)
	}
	if result2.Value != "HELLO" {
		t.Errorf("Script 2: expected 'HELLO' (original builtin), got %v", result2.Value)
	}
}
