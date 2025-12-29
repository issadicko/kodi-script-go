// Package kodi provides a simple API for executing KodiScript code.
package kodi

import (
	"github.com/kodi-script/kodi-go/interpreter"
	"github.com/kodi-script/kodi-go/lexer"
	"github.com/kodi-script/kodi-go/natives"
	"github.com/kodi-script/kodi-go/parser"
)

// Script represents a compiled KodiScript program.
type Script struct {
	source  string
	interp  *interpreter.Interpreter
	natives *natives.Registry
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
		source:  source,
		natives: natives.NewRegistry(),
	}
}

// WithVariables injects host variables into the script context.
func (s *Script) WithVariables(vars map[string]interface{}) *Script {
	s.interp = interpreter.NewWithEnv(vars)
	return s
}

// RegisterFunction adds a custom native function.
func (s *Script) RegisterFunction(name string, fn natives.NativeFunc) *Script {
	s.natives.Register(name, fn)
	return s
}

// Execute runs the script and returns the result.
func (s *Script) Execute() *Result {
	result := &Result{}

	// Lexer
	l := lexer.New(s.source)

	// Parser
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		result.Errors = p.Errors()
		return result
	}

	// Interpreter
	if s.interp == nil {
		s.interp = interpreter.New()
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
