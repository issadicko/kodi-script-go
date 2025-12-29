// Package interpreter evaluates KodiScript AST nodes.
package interpreter

import (
	"fmt"

	"github.com/kodi-script/kodi-go/ast"
	"github.com/kodi-script/kodi-go/natives"
)

// Value represents a runtime value in KodiScript.
type Value interface{}

// Environment holds variable bindings.
type Environment struct {
	store  map[string]Value
	outer  *Environment
	output []string // captured output from print()
}

// NewEnvironment creates a new environment.
func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Value), output: []string{}}
}

// NewEnclosedEnvironment creates a new environment enclosed by an outer one.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// Get retrieves a variable value.
func (e *Environment) Get(name string) (Value, bool) {
	val, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return val, ok
}

// Set sets a variable value.
func (e *Environment) Set(name string, val Value) {
	e.store[name] = val
}

// GetOutput returns all captured output.
func (e *Environment) GetOutput() []string {
	return e.output
}

// AddOutput adds a line to captured output.
func (e *Environment) AddOutput(line string) {
	e.output = append(e.output, line)
}

// Interpreter evaluates AST nodes.
type Interpreter struct {
	env     *Environment
	natives *natives.Registry
}

// New creates a new Interpreter.
func New() *Interpreter {
	return &Interpreter{
		env:     NewEnvironment(),
		natives: natives.NewRegistry(),
	}
}

// NewWithEnv creates an Interpreter with pre-injected variables.
func NewWithEnv(variables map[string]interface{}) *Interpreter {
	interp := New()
	for k, v := range variables {
		interp.env.Set(k, v)
	}
	return interp
}

// Eval evaluates a program and returns the final result.
func (i *Interpreter) Eval(program *ast.Program) (Value, error) {
	var result Value

	for _, stmt := range program.Statements {
		val, err := i.evalStatement(stmt)
		if err != nil {
			return nil, err
		}
		result = val
	}

	return result, nil
}

// GetOutput returns captured print() output.
func (i *Interpreter) GetOutput() []string {
	return i.env.GetOutput()
}

func (i *Interpreter) evalStatement(stmt ast.Statement) (Value, error) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		val, err := i.evalExpression(s.Value)
		if err != nil {
			return nil, err
		}
		i.env.Set(s.Name.Value, val)
		return val, nil

	case *ast.Assignment:
		val, err := i.evalExpression(s.Value)
		if err != nil {
			return nil, err
		}
		i.env.Set(s.Name.Value, val)
		return val, nil

	case *ast.ExpressionStatement:
		return i.evalExpression(s.Expression)

	case *ast.IfStatement:
		return i.evalIfStatement(s)

	default:
		return nil, fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (i *Interpreter) evalIfStatement(stmt *ast.IfStatement) (Value, error) {
	condition, err := i.evalExpression(stmt.Condition)
	if err != nil {
		return nil, err
	}

	if isTruthy(condition) {
		return i.evalBlockStatement(stmt.Consequence)
	} else if stmt.Alternative != nil {
		return i.evalBlockStatement(stmt.Alternative)
	}

	return nil, nil
}

func (i *Interpreter) evalBlockStatement(block *ast.BlockStatement) (Value, error) {
	var result Value

	for _, stmt := range block.Statements {
		val, err := i.evalStatement(stmt)
		if err != nil {
			return nil, err
		}
		result = val
	}

	return result, nil
}

func (i *Interpreter) evalExpression(expr ast.Expression) (Value, error) {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return e.Value, nil

	case *ast.StringLiteral:
		return e.Value, nil

	case *ast.BooleanLiteral:
		return e.Value, nil

	case *ast.NullLiteral:
		return nil, nil

	case *ast.Identifier:
		val, ok := i.env.Get(e.Value)
		if !ok {
			return nil, fmt.Errorf("undefined variable: %s", e.Value)
		}
		return val, nil

	case *ast.BinaryExpr:
		return i.evalBinaryExpr(e)

	case *ast.UnaryExpr:
		return i.evalUnaryExpr(e)

	case *ast.SafeAccessExpr:
		return i.evalSafeAccess(e)

	case *ast.ElvisExpr:
		return i.evalElvisExpr(e)

	case *ast.PropertyAccessExpr:
		return i.evalPropertyAccess(e)

	case *ast.CallExpr:
		return i.evalCallExpr(e)

	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

func (i *Interpreter) evalBinaryExpr(expr *ast.BinaryExpr) (Value, error) {
	left, err := i.evalExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	// Short-circuit evaluation for && and ||
	if expr.Operator == "&&" {
		if !isTruthy(left) {
			return false, nil
		}
		right, err := i.evalExpression(expr.Right)
		if err != nil {
			return nil, err
		}
		return isTruthy(right), nil
	}

	if expr.Operator == "||" {
		if isTruthy(left) {
			return true, nil
		}
		right, err := i.evalExpression(expr.Right)
		if err != nil {
			return nil, err
		}
		return isTruthy(right), nil
	}

	right, err := i.evalExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "+":
		return i.evalPlus(left, right)
	case "-":
		return i.evalArithmetic(left, right, "-")
	case "*":
		return i.evalArithmetic(left, right, "*")
	case "/":
		return i.evalArithmetic(left, right, "/")
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case "<":
		return i.evalComparison(left, right, "<")
	case ">":
		return i.evalComparison(left, right, ">")
	case "<=":
		return i.evalComparison(left, right, "<=")
	case ">=":
		return i.evalComparison(left, right, ">=")
	default:
		return nil, fmt.Errorf("unknown operator: %s", expr.Operator)
	}
}

func (i *Interpreter) evalPlus(left, right Value) (Value, error) {
	// String concatenation
	if ls, ok := left.(string); ok {
		return ls + fmt.Sprintf("%v", right), nil
	}
	if rs, ok := right.(string); ok {
		return fmt.Sprintf("%v", left) + rs, nil
	}

	// Numeric addition
	leftNum, lok := toNumber(left)
	rightNum, rok := toNumber(right)
	if lok && rok {
		return leftNum + rightNum, nil
	}

	return nil, fmt.Errorf("cannot add %T and %T", left, right)
}

func (i *Interpreter) evalArithmetic(left, right Value, op string) (Value, error) {
	leftNum, lok := toNumber(left)
	rightNum, rok := toNumber(right)
	if !lok || !rok {
		return nil, fmt.Errorf("cannot perform %s on %T and %T", op, left, right)
	}

	switch op {
	case "-":
		return leftNum - rightNum, nil
	case "*":
		return leftNum * rightNum, nil
	case "/":
		if rightNum == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return leftNum / rightNum, nil
	}
	return nil, fmt.Errorf("unknown arithmetic operator: %s", op)
}

func (i *Interpreter) evalComparison(left, right Value, op string) (Value, error) {
	leftNum, lok := toNumber(left)
	rightNum, rok := toNumber(right)
	if !lok || !rok {
		return nil, fmt.Errorf("cannot compare %T and %T", left, right)
	}

	switch op {
	case "<":
		return leftNum < rightNum, nil
	case ">":
		return leftNum > rightNum, nil
	case "<=":
		return leftNum <= rightNum, nil
	case ">=":
		return leftNum >= rightNum, nil
	}
	return nil, fmt.Errorf("unknown comparison operator: %s", op)
}

func (i *Interpreter) evalUnaryExpr(expr *ast.UnaryExpr) (Value, error) {
	right, err := i.evalExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "-":
		if num, ok := toNumber(right); ok {
			return -num, nil
		}
		return nil, fmt.Errorf("cannot negate %T", right)
	case "!":
		return !isTruthy(right), nil
	}

	return nil, fmt.Errorf("unknown unary operator: %s", expr.Operator)
}

func (i *Interpreter) evalSafeAccess(expr *ast.SafeAccessExpr) (Value, error) {
	object, err := i.evalExpression(expr.Object)
	if err != nil {
		return nil, err
	}

	// If object is null, return null (safe navigation)
	if object == nil {
		return nil, nil
	}

	// Try to access property on map
	if m, ok := object.(map[string]interface{}); ok {
		return m[expr.Property.Value], nil
	}

	return nil, nil
}

func (i *Interpreter) evalElvisExpr(expr *ast.ElvisExpr) (Value, error) {
	left, err := i.evalExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	if left != nil {
		return left, nil
	}

	return i.evalExpression(expr.Default)
}

func (i *Interpreter) evalPropertyAccess(expr *ast.PropertyAccessExpr) (Value, error) {
	object, err := i.evalExpression(expr.Object)
	if err != nil {
		return nil, err
	}

	if object == nil {
		return nil, fmt.Errorf("cannot access property '%s' on null", expr.Property.Value)
	}

	if m, ok := object.(map[string]interface{}); ok {
		return m[expr.Property.Value], nil
	}

	return nil, fmt.Errorf("cannot access property on %T", object)
}

func (i *Interpreter) evalCallExpr(expr *ast.CallExpr) (Value, error) {
	// Get function name
	ident, ok := expr.Function.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected function identifier")
	}

	// Evaluate arguments
	args := make([]Value, len(expr.Arguments))
	for idx, arg := range expr.Arguments {
		val, err := i.evalExpression(arg)
		if err != nil {
			return nil, err
		}
		args[idx] = val
	}

	// Handle print specially (capture output)
	if ident.Value == "print" {
		for _, arg := range args {
			i.env.AddOutput(fmt.Sprintf("%v", arg))
		}
		return nil, nil
	}

	// Look up native function
	fn := i.natives.Get(ident.Value)
	if fn == nil {
		return nil, fmt.Errorf("undefined function: %s", ident.Value)
	}

	// Convert []Value to []interface{} for native function call
	ifaceArgs := make([]interface{}, len(args))
	for i, arg := range args {
		ifaceArgs[i] = arg
	}

	return fn(ifaceArgs...)
}

// Utility functions

func isTruthy(val Value) bool {
	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return true
}

func toNumber(val Value) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	}
	return 0, false
}
