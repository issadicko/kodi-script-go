// Package interpreter evaluates KodiScript AST nodes.
package interpreter

import (
	"fmt"

	"github.com/issadicko/kodi-script-go/ast"
	"github.com/issadicko/kodi-script-go/natives"
)

// Value represents a runtime value in KodiScript.
type Value interface{}

// ReturnValue wraps a value to signal an early return from execution.
type ReturnValue struct {
	Value Value
}

// Function represents a user-defined function.
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

// NativeFunction wraps a built-in function.
type NativeFunction struct {
	Fn natives.NativeFunc
}

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
		// Unwrap return values at the top level
		if rv, ok := val.(*ReturnValue); ok {
			return rv.Value, nil
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

	case *ast.ReturnStatement:
		var val Value
		if s.Value != nil {
			var err error
			val, err = i.evalExpression(s.Value)
			if err != nil {
				return nil, err
			}
		}
		return &ReturnValue{Value: val}, nil

	case *ast.ForStatement:
		return i.evalForStatement(s)

	default:
		return nil, fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (i *Interpreter) evalForStatement(stmt *ast.ForStatement) (Value, error) {
	// Evaluate the iterable
	iterableVal, err := i.evalExpression(stmt.Iterable)
	if err != nil {
		return nil, err
	}

	// Must be an array
	arr, ok := iterableVal.([]interface{})
	if !ok {
		return nil, fmt.Errorf("for-in requires an array, got %T", iterableVal)
	}

	var result Value
	varName := stmt.Variable.Value

	for _, item := range arr {
		// Set loop variable in current environment
		i.env.Set(varName, item)

		// Execute body
		val, err := i.evalBlockStatement(stmt.Body)
		if err != nil {
			return nil, err
		}

		// Check for return
		if rv, ok := val.(*ReturnValue); ok {
			return rv, nil
		}

		result = val
	}

	return result, nil
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
		// Propagate return values up the call stack
		if _, ok := val.(*ReturnValue); ok {
			return val, nil
		}
		result = val
	}

	return result, nil
}

// evalStringTemplate evaluates a string template by evaluating each part
// and concatenating the results into a single string.
func (i *Interpreter) evalStringTemplate(tmpl *ast.StringTemplate) (Value, error) {
	var result string

	for _, part := range tmpl.Parts {
		val, err := i.evalExpression(part)
		if err != nil {
			return nil, err
		}

		// Convert to string
		if val == nil {
			result += "null"
		} else {
			result += fmt.Sprintf("%v", val)
		}
	}

	return result, nil
}

func (i *Interpreter) evalExpression(expr ast.Expression) (Value, error) {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return e.Value, nil

	case *ast.StringLiteral:
		return e.Value, nil

	case *ast.StringTemplate:
		return i.evalStringTemplate(e)

	case *ast.BooleanLiteral:
		return e.Value, nil

	case *ast.NullLiteral:
		return nil, nil

	case *ast.Identifier:
		val, ok := i.env.Get(e.Value)
		if ok {
			return val, nil
		}
		if fn := i.natives.Get(e.Value); fn != nil {
			return &NativeFunction{Fn: fn}, nil
		}
		return nil, fmt.Errorf("undefined variable: %s", e.Value)

	case *ast.FunctionLiteral:
		return &Function{Parameters: e.Parameters, Body: e.Body, Env: i.env}, nil

	case *ast.BinaryExpr:
		return i.evalBinaryExpr(e)

	case *ast.UnaryExpr:
		return i.evalUnaryExpr(e)

	case *ast.ArrayLiteral:
		elements := make([]interface{}, len(e.Elements))
		for idx, el := range e.Elements {
			val, err := i.evalExpression(el)
			if err != nil {
				return nil, err
			}
			elements[idx] = val
		}
		return elements, nil

	case *ast.ObjectLiteral:
		pairs := make(map[string]interface{})
		for key, valExpr := range e.Pairs {
			val, err := i.evalExpression(valExpr)
			if err != nil {
				return nil, err
			}
			pairs[key] = val
		}
		return pairs, nil

	case *ast.IndexExpr:
		left, err := i.evalExpression(e.Left)
		if err != nil {
			return nil, err
		}
		index, err := i.evalExpression(e.Index)
		if err != nil {
			return nil, err
		}
		return i.evalIndexExpression(left, index)

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
	// Special handling for print (keep it special to capture output in env)
	if ident, ok := expr.Function.(*ast.Identifier); ok && ident.Value == "print" {
		args := make([]Value, len(expr.Arguments))
		for idx, arg := range expr.Arguments {
			val, err := i.evalExpression(arg)
			if err != nil {
				return nil, err
			}
			args[idx] = val
		}

		for _, arg := range args {
			i.env.AddOutput(fmt.Sprintf("%v", arg))
		}
		return nil, nil
	}

	function, err := i.evalExpression(expr.Function)
	if err != nil {
		return nil, err
	}

	args := make([]Value, len(expr.Arguments))
	for idx, arg := range expr.Arguments {
		val, err := i.evalExpression(arg)
		if err != nil {
			return nil, err
		}
		args[idx] = val
	}

	return i.applyFunction(function, args)
}

func (i *Interpreter) applyFunction(fn Value, args []Value) (Value, error) {
	switch function := fn.(type) {
	case *Function:
		extendedEnv := NewEnclosedEnvironment(function.Env)
		for idx, param := range function.Parameters {
			if idx < len(args) {
				extendedEnv.Set(param.Value, args[idx])
			}
		}
		savedEnv := i.env
		i.env = extendedEnv
		val, err := i.evalBlockStatement(function.Body)
		i.env = savedEnv // Restore env
		if err != nil {
			return nil, err
		}
		if rv, ok := val.(*ReturnValue); ok {
			return rv.Value, nil
		}
		return val, nil

	case *NativeFunction:
		ifaceArgs := make([]interface{}, len(args))
		for i, arg := range args {
			ifaceArgs[i] = arg
		}
		return function.Fn(ifaceArgs...)

	default:
		return nil, fmt.Errorf("not a function: %T", fn)
	}
}

func (i *Interpreter) evalIndexExpression(left, index Value) (Value, error) {
	switch l := left.(type) {
	case []interface{}: // []Value is alias to []interface{}
		return i.evalArrayIndexExpression(l, index)
	case map[string]interface{}: // map[string]Value is alias to map[string]interface{}
		return i.evalHashIndexExpression(l, index)
	default:
		return nil, fmt.Errorf("index operator not supported: %T", left)
	}
}

func (i *Interpreter) evalArrayIndexExpression(array []interface{}, index Value) (Value, error) {
	var idx int

	switch iVal := index.(type) {
	case int:
		idx = iVal
	case float64:
		idx = int(iVal)
	default:
		return nil, fmt.Errorf("index must be a number")
	}

	if idx < 0 || idx >= len(array) {
		return nil, nil // Return null for out of bounds
	}
	return array[idx], nil
}

func (i *Interpreter) evalHashIndexExpression(hash map[string]interface{}, index Value) (Value, error) {
	key, ok := index.(string)
	if !ok {
		return nil, fmt.Errorf("property access must be a string")
	}
	val, ok := hash[key]
	if !ok {
		return nil, nil // Return null if key not found
	}
	return val, nil
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
