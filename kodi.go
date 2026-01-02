// Package kodi provides a simple API for executing KodiScript code.
package kodi

import (
	"context"
	"time"

	"github.com/issadicko/kodi-script-go/ast"
	"github.com/issadicko/kodi-script-go/cache"
	"github.com/issadicko/kodi-script-go/interpreter"
	"github.com/issadicko/kodi-script-go/lexer"
	"github.com/issadicko/kodi-script-go/natives"
	"github.com/issadicko/kodi-script-go/parser"
)

// Script represents a compiled KodiScript program.
type Script struct {
	source      string
	program     *ast.Program // cached parsed program
	interp      *interpreter.Interpreter
	natives     *natives.Registry
	silentPrint bool
	useCache    bool
	maxOps      int64         // Maximum operations (0 = unlimited)
	timeout     time.Duration // Execution timeout (0 = no timeout)
}

// Result represents the result of script execution.
type Result struct {
	Value  interface{}
	Output []string
	Errors []string
}

// New creates a new Script from source code.
func New(source string) *Script {
	return &Script{
		source:   source,
		natives:  natives.NewRegistry(),
		useCache: true, // Enable cache by default
	}
}

// WithCache enables or disables AST caching.
func (s *Script) WithCache(enabled bool) *Script {
	s.useCache = enabled
	return s
}

// WithVariables injects host variables into the script context.
func (s *Script) WithVariables(vars map[string]interface{}) *Script {
	s.interp = interpreter.NewWithEnv(vars)
	return s
}

// SilentPrint disables console output for print() calls.
func (s *Script) SilentPrint(silent bool) *Script {
	s.silentPrint = silent
	return s
}

// RegisterFunction adds a custom native function.
func (s *Script) RegisterFunction(name string, fn natives.NativeFunc) *Script {
	s.natives.Register(name, fn)
	return s
}

// Bind adds a Go object to the script context with reflective access.
// All public methods and fields of the object will be accessible from KodiScript.
func (s *Script) Bind(name string, obj interface{}) *Script {
	if s.interp == nil {
		s.interp = interpreter.New()
	}
	s.interp.SetGlobal(name, obj)
	return s
}

// WithMaxOperations sets the maximum number of operations allowed.
// If the limit is exceeded, execution will stop with ErrMaxOperationsExceeded.
// Use this to protect against infinite loops or overly complex scripts.
func (s *Script) WithMaxOperations(maxOps int64) *Script {
	s.maxOps = maxOps
	return s
}

// WithTimeout sets a timeout for script execution.
// If the timeout is exceeded, execution will stop with ErrTimeout.
func (s *Script) WithTimeout(timeout time.Duration) *Script {
	s.timeout = timeout
	return s
}

// Execute runs the script and returns the result.
func (s *Script) Execute() *Result {
	result := &Result{}

	var program *ast.Program

	// Try to get from cache first
	if s.useCache {
		if cached, ok := cache.DefaultCache.Get(s.source); ok {
			program = cached
		}
	}

	// Parse if not cached
	if program == nil {
		l := lexer.New(s.source)
		p := parser.New(l)
		program = p.ParseProgram()

		if len(p.Errors()) > 0 {
			result.Errors = p.Errors()
			return result
		}

		// Store in cache
		if s.useCache {
			cache.DefaultCache.Set(s.source, program)
		}
	}

	// Interpreter
	if s.interp == nil {
		s.interp = interpreter.New()
	}

	// Apply operation limit if set
	if s.maxOps > 0 {
		s.interp.SetMaxOperations(s.maxOps)
	}

	// Apply timeout if set
	if s.timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()
		s.interp.SetContext(ctx)
	}

	val, err := s.interp.Eval(program)
	if err != nil {
		result.Errors = []string{err.Error()}
		return result
	}

	result.Value = val
	result.Output = s.interp.GetOutput()

	return result
}

// Run is a convenience function to execute KodiScript code with optional variables.
func Run(source string, variables map[string]interface{}) *Result {
	script := New(source)
	if variables != nil {
		script.WithVariables(variables)
	}
	return script.Execute()
}

// Eval is the simplest way to execute KodiScript code.
func Eval(source string) (interface{}, error) {
	result := Run(source, nil)
	if len(result.Errors) > 0 {
		return nil, &EvalError{Messages: result.Errors}
	}
	return result.Value, nil
}

// EvalError represents evaluation errors.
type EvalError struct {
	Messages []string
}

func (e *EvalError) Error() string {
	if len(e.Messages) == 0 {
		return "unknown error"
	}
	return e.Messages[0]
}
