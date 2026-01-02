package kodi

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"strconv"
)

func TestCompliance(t *testing.T) {
	compliancePath := os.Getenv("KODI_COMPLIANCE_TESTS_PATH")
	if compliancePath == "" {
		// Default for local development
		compliancePath = "../compliance-tests"
	}

	absPath, err := filepath.Abs(compliancePath)
	if err != nil {
		t.Fatalf("Failed to get absolute path for %s: %v", compliancePath, err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skipf("Compliance tests directory not found at %s. Skipping compliance tests.", absPath)
	}

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".kodi") {
			runComplianceTest(t, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Error walking compliance directory: %v", err)
	}
}

func runComplianceTest(t *testing.T, sourcePath string) {
	t.Run(filepath.Base(sourcePath), func(t *testing.T) {
		// Read source code
		sourceBytes, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}
		source := string(sourceBytes)

		// Parse directives
		var maxOps int64
		var expectError bool

		lines := strings.Split(source, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "// config:") {
				parts := strings.SplitN(strings.TrimPrefix(line, "// config:"), "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.TrimSpace(parts[1])
					if key == "maxOps" {
						if n, err := strconv.ParseInt(val, 10, 64); err == nil {
							maxOps = n
						}
					}
				}
			}
			if strings.HasPrefix(line, "// expect: error") {
				expectError = true
			}
		}

		// Read expected output
		outPath := strings.TrimSuffix(sourcePath, ".kodi") + ".out"
		expectedOutBytes, err := os.ReadFile(outPath)
		if err != nil && !expectError {
			// It is okay if .out is missing if we expect an error, but usually we should have it
			// For this specific logic, we'll try to read it.
			t.Fatalf("Failed to read expected output file: %v", err)
		}
		expectedOut := string(expectedOutBytes)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute using public API
		script := New(source)
		if maxOps > 0 {
			script.WithMaxOperations(maxOps)
		}

		result := script.Execute()

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout
		var buf bytes.Buffer
		io.Copy(&buf, r)
		actualOut := buf.String()

		// Check expectations
		if expectError {
			if len(result.Errors) == 0 {
				t.Errorf("Expected execution error but got none")
			}
		} else {
			if len(result.Errors) > 0 {
				t.Fatalf("Execution failed: %v", result.Errors)
			}

			// Normalize output for cross-platform line endings
			normalize := func(s string) string {
				return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
			}
			expectedOut = normalize(expectedOut)
			actualOut = normalize(actualOut)

			if actualOut != expectedOut {
				t.Errorf("Output mismatch for %s.\nExpected:\n%s\nActual:\n%s",
					filepath.Base(sourcePath), expectedOut, actualOut)
			}
		}
	})
}
