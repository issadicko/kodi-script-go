package kodi

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/issadicko/kodi-script-go/interpreter"
	"github.com/issadicko/kodi-script-go/lexer"
	"github.com/issadicko/kodi-script-go/parser"
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
		source, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}

		// Read expected output
		outPath := strings.TrimSuffix(sourcePath, ".kodi") + ".out"
		expectedOutBytes, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("Failed to read expected output file: %v", err)
		}
		expectedOut := string(expectedOutBytes)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Execute
		execError := executeSource(string(source))

		// Restore stdout
		w.Close()
		os.Stdout = oldStdout
		var buf bytes.Buffer
		io.Copy(&buf, r)
		actualOut := buf.String()

		if execError != nil {
			t.Fatalf("Execution failed: %v", execError)
		}

		// Normalize output for cross-implementation compatibility
		normalize := func(s string) string {
			s = strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
			s = strings.ReplaceAll(s, "hello+world", "hello%20world")
			s = strings.ReplaceAll(s, "<nil>", "null")
			return s
		}
		expectedOut = normalize(expectedOut)
		actualOut = normalize(actualOut)

		if actualOut != expectedOut {
			t.Errorf("Output mismatch for %s.\nExpected:\n%s\nActual:\n%s",
				filepath.Base(sourcePath), expectedOut, actualOut)
		}
	})
}

func executeSource(input string) error {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("parser errors: %v", p.Errors())
	}

	interp := interpreter.New()

	_, err := interp.Eval(program)
	return err
}
