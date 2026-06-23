// Package interpreter evaluates KodiScript AST nodes.
package interpreter

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/issadicko/kodi-script-go/ast"
	"github.com/issadicko/kodi-script-go/natives"
	"github.com/issadicko/kodi-script-go/token"
)

// ErrMaxOperationsExceeded is returned when the operation limit is exceeded.
var ErrMaxOperationsExceeded = errors.New("max operations exceeded")

// ErrTimeout is returned when the execution deadline is exceeded.
var ErrTimeout = errors.New("execution timeout")

// ErrStackOverflow is returned when the function call depth limit is exceeded,
// guarding against unbounded recursion exhausting the native stack.
var ErrStackOverflow = errors.New("maximum call depth exceeded")

// maxCallDepth bounds nested user-function calls.
const maxCallDepth = 1000

// Value represents a runtime value in KodiScript.
type Value interface{}

// ReturnValue wraps a value to signal an early return from execution.
type ReturnValue struct {
	Value Value
}

// BreakValue signals a `break` out of the nearest enclosing loop.
type BreakValue struct{}

// ContinueValue signals a `continue` to the next iteration of the nearest loop.
type ContinueValue struct{}

var (
	breakSignal    = &BreakValue{}
	continueSignal = &ContinueValue{}
)

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
	store map[string]Value
	outer *Environment
}

// envPool pools Environment objects to reduce allocations.
var envPool = sync.Pool{
	New: func() interface{} {
		return &Environment{
			store: make(map[string]Value, 16),
		}
	},
}

// NewEnvironment creates a new environment from pool.
func NewEnvironment() *Environment {
	env := envPool.Get().(*Environment)
	env.outer = nil
	return env
}

// Release returns the environment to the pool.
func (e *Environment) Release() {
	// Clear the store
	for k := range e.store {
		delete(e.store, k)
	}
	e.outer = nil
	envPool.Put(e)
}

// NewEnclosedEnvironment creates a new environment enclosed by an outer one.
// Note: We don't use pooling here because closures may capture this environment.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{
		store: make(map[string]Value, 8),
		outer: outer,
	}
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

// Interpreter evaluates AST nodes.
type Interpreter struct {
	env         *Environment
	natives     *natives.Registry
	opCount     int64           // Current operation count
	maxOps      int64           // Maximum allowed operations (0 = unlimited)
	ctx         context.Context // Context for timeout support
	output      []string        // captured output from print()
	silentPrint bool            // when true, print() does not write to stdout
	outputSink  func(string)    // when set, print() routes here instead of stdout
	callDepth   int             // current user-function call depth (recursion guard)
}

// New creates a new Interpreter.
func New() *Interpreter {
	return &Interpreter{
		env:     NewEnvironment(),
		natives: natives.DefaultBuiltins, // Use shared builtins by default
	}
}

// SetNatives sets the native function registry for the interpreter.
// This allows using a custom registry with per-script functions.
func (i *Interpreter) SetNatives(registry *natives.Registry) {
	i.natives = registry
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
		// Ignore a stray break/continue used outside any loop.
		switch val.(type) {
		case *BreakValue, *ContinueValue:
			continue
		}
		result = val
	}

	return result, nil
}

// GetOutput returns captured print() output.
func (i *Interpreter) GetOutput() []string {
	return i.output
}

// SetSilentPrint controls whether print() writes to stdout.
// Output is always captured and available via GetOutput regardless of this setting.
func (i *Interpreter) SetSilentPrint(silent bool) {
	i.silentPrint = silent
}

// SetOutputSink routes print() output to the given callback instead of stdout.
// Output is still captured and available via GetOutput.
func (i *Interpreter) SetOutputSink(sink func(string)) {
	i.outputSink = sink
}

// SetGlobal sets a global variable in the interpreter's environment.
func (i *Interpreter) SetGlobal(name string, value Value) {
	i.env.Set(name, value)
}

// SetMaxOperations sets the maximum number of operations allowed.
// If maxOps is 0, there is no limit (default behavior).
func (i *Interpreter) SetMaxOperations(maxOps int64) {
	i.maxOps = maxOps
	i.opCount = 0
}

// checkOperationLimit increments the operation counter and returns an error if the limit is exceeded.
func (i *Interpreter) checkOperationLimit() error {
	if i.maxOps > 0 {
		i.opCount++
		if i.opCount > i.maxOps {
			return ErrMaxOperationsExceeded
		}
	}
	return nil
}

// SetContext sets a context for timeout support.
func (i *Interpreter) SetContext(ctx context.Context) {
	i.ctx = ctx
}

// checkTimeout checks if the context deadline has been exceeded.
func (i *Interpreter) checkTimeout() error {
	if i.ctx != nil {
		select {
		case <-i.ctx.Done():
			return ErrTimeout
		default:
			return nil
		}
	}
	return nil
}

func (i *Interpreter) evalStatement(stmt ast.Statement) (Value, error) {
	// Check operation limit at each statement
	if err := i.checkOperationLimit(); err != nil {
		return nil, err
	}
	// Check timeout at each statement
	if err := i.checkTimeout(); err != nil {
		return nil, err
	}

	val, err := i.evalStatementBody(stmt)
	if err != nil {
		return nil, i.positionError(err, stmt)
	}
	return val, nil
}

func (i *Interpreter) evalStatementBody(stmt ast.Statement) (Value, error) {
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

	case *ast.ArrayDestructure:
		val, err := i.evalExpression(s.Value)
		if err != nil {
			return nil, err
		}
		arr, ok := val.([]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot destructure non-array value (%T)", val)
		}
		for idx, name := range s.Names {
			if idx < len(arr) {
				i.env.Set(name.Value, arr[idx])
			} else {
				i.env.Set(name.Value, nil)
			}
		}
		return val, nil

	case *ast.ObjectDestructure:
		val, err := i.evalExpression(s.Value)
		if err != nil {
			return nil, err
		}
		m, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot destructure non-object value (%T)", val)
		}
		for _, name := range s.Names {
			i.env.Set(name.Value, m[name.Value])
		}
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

	case *ast.WhileStatement:
		return i.evalWhileStatement(s)

	case *ast.TryStatement:
		return i.evalTryStatement(s)

	case *ast.BreakStatement:
		return breakSignal, nil

	case *ast.ContinueStatement:
		return continueSignal, nil

	default:
		return nil, fmt.Errorf("unknown statement type: %T", stmt)
	}
}

// evalTryStatement runs the protected block; on a (non-limit) runtime error it
// binds the error message to the catch variable and runs the handler.
func (i *Interpreter) evalTryStatement(stmt *ast.TryStatement) (Value, error) {
	val, err := i.evalBlockStatement(stmt.Body)
	if err == nil {
		return val, nil
	}
	// Timeout / operation-limit errors are not catchable.
	if errors.Is(err, ErrTimeout) || errors.Is(err, ErrMaxOperationsExceeded) {
		return nil, err
	}
	if stmt.CatchVar != nil {
		i.env.Set(stmt.CatchVar.Value, err.Error())
	}
	return i.evalBlockStatement(stmt.Catch)
}

// RuntimeError carries source position information for a runtime failure.
type RuntimeError struct {
	Msg  string
	Line int
	Col  int
}

func (e *RuntimeError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d, col %d: %s", e.Line, e.Col, e.Msg)
	}
	return e.Msg
}

// positionError attaches the failing statement's position to a plain runtime
// error. Limit errors and already-positioned errors pass through unchanged, so
// the innermost statement's position wins.
func (i *Interpreter) positionError(err error, stmt ast.Statement) error {
	if errors.Is(err, ErrTimeout) || errors.Is(err, ErrMaxOperationsExceeded) {
		return err
	}
	if _, ok := err.(*RuntimeError); ok {
		return err
	}
	tok := stmtToken(stmt)
	return &RuntimeError{Msg: err.Error(), Line: tok.Line, Col: tok.Column}
}

// stmtToken returns the leading token (with position) of a statement.
func stmtToken(stmt ast.Statement) token.Token {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		return s.Token
	case *ast.ArrayDestructure:
		return s.Token
	case *ast.ObjectDestructure:
		return s.Token
	case *ast.Assignment:
		return s.Token
	case *ast.ExpressionStatement:
		return s.Token
	case *ast.IfStatement:
		return s.Token
	case *ast.ForStatement:
		return s.Token
	case *ast.WhileStatement:
		return s.Token
	case *ast.ReturnStatement:
		return s.Token
	case *ast.TryStatement:
		return s.Token
	case *ast.BreakStatement:
		return s.Token
	case *ast.ContinueStatement:
		return s.Token
	case *ast.BlockStatement:
		return s.Token
	}
	return token.Token{}
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
		// Check operation limit at each iteration
		if err := i.checkOperationLimit(); err != nil {
			return nil, err
		}
		// Check timeout at each iteration
		if err := i.checkTimeout(); err != nil {
			return nil, err
		}

		// Set loop variable in current environment
		i.env.Set(varName, item)

		// Execute body
		val, err := i.evalBlockStatement(stmt.Body)
		if err != nil {
			return nil, err
		}

		// Handle return/break/continue signals
		switch val.(type) {
		case *ReturnValue:
			return val, nil
		case *BreakValue:
			return result, nil
		case *ContinueValue:
			continue
		}

		result = val
	}

	return result, nil
}

func (i *Interpreter) evalWhileStatement(stmt *ast.WhileStatement) (Value, error) {
	var result Value

	for {
		// Check operation limit at each iteration
		if err := i.checkOperationLimit(); err != nil {
			return nil, err
		}
		// Check timeout at each iteration
		if err := i.checkTimeout(); err != nil {
			return nil, err
		}

		// Evaluate condition
		condVal, err := i.evalExpression(stmt.Condition)
		if err != nil {
			return nil, err
		}

		// Exit if condition is false
		if !isTruthy(condVal) {
			break
		}

		// Execute body
		val, err := i.evalBlockStatement(stmt.Body)
		if err != nil {
			return nil, err
		}

		// Handle return/break/continue signals
		switch val.(type) {
		case *ReturnValue:
			return val, nil
		case *BreakValue:
			return result, nil
		case *ContinueValue:
			continue
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
		// Propagate return/break/continue signals up to the nearest handler
		switch val.(type) {
		case *ReturnValue, *BreakValue, *ContinueValue:
			return val, nil
		}
		result = val
	}

	return result, nil
}

// evalStringTemplate evaluates a string template by evaluating each part
// and concatenating the results into a single string.
func (i *Interpreter) evalStringTemplate(tmpl *ast.StringTemplate) (Value, error) {
	var sb strings.Builder

	for _, part := range tmpl.Parts {
		val, err := i.evalExpression(part)
		if err != nil {
			return nil, err
		}

		// Convert to string using fast path
		sb.WriteString(toString(val))
	}

	return sb.String(), nil
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
		return i.evalElements(e.Elements)

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

	case *ast.TernaryExpr:
		cond, err := i.evalExpression(e.Condition)
		if err != nil {
			return nil, err
		}
		if isTruthy(cond) {
			return i.evalExpression(e.Consequent)
		}
		return i.evalExpression(e.Alternative)

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
	case "%":
		return i.evalArithmetic(left, right, "%")
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
	// String concatenation with fast path
	if ls, ok := left.(string); ok {
		return ls + toString(right), nil
	}
	if rs, ok := right.(string); ok {
		return toString(left) + rs, nil
	}

	// Numeric addition
	leftNum, lok := toNumber(left)
	rightNum, rok := toNumber(right)
	if lok && rok {
		return leftNum + rightNum, nil
	}

	return nil, fmt.Errorf("cannot add %T and %T", left, right)
}

// toString renders a value in canonical form (shared with natives.Stringify).
func toString(val Value) string {
	return natives.Stringify(val)
}

func (i *Interpreter) evalArithmetic(left, right Value, op string) (Value, error) {
	// Fast path for float64 (most common case)
	if lf, lok := left.(float64); lok {
		if rf, rok := right.(float64); rok {
			switch op {
			case "-":
				return lf - rf, nil
			case "*":
				return lf * rf, nil
			case "/":
				if rf == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				return lf / rf, nil
			case "%":
				if rf == 0 {
					return nil, fmt.Errorf("modulo by zero")
				}
				return math.Mod(lf, rf), nil
			}
		}
	}

	// Fallback to toNumber conversion
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
	case "%":
		if rightNum == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return math.Mod(leftNum, rightNum), nil
	}
	return nil, fmt.Errorf("unknown arithmetic operator: %s", op)
}

func (i *Interpreter) evalComparison(left, right Value, op string) (Value, error) {
	// Fast path for float64 (most common case)
	if lf, lok := left.(float64); lok {
		if rf, rok := right.(float64); rok {
			switch op {
			case "<":
				return lf < rf, nil
			case ">":
				return lf > rf, nil
			case "<=":
				return lf <= rf, nil
			case ">=":
				return lf >= rf, nil
			}
		}
	}

	// Fallback to toNumber conversion
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

	// First check for map access (existing behavior)
	if m, ok := object.(map[string]interface{}); ok {
		return m[expr.Property.Value], nil
	}

	// Use reflection to access methods and fields on Go objects
	return i.reflectivePropertyAccess(object, expr.Property.Value)
}

// interpBuiltins holds built-in functions that need the interpreter itself
// (to capture output or to call back into user functions). Registering them
// in a table — rather than special-casing them in evalCallExpr — keeps the
// dispatch data-driven: adding a new one no longer touches the evaluator, and
// scripts/hosts can override them by name.
var interpBuiltins map[string]func(*Interpreter, []Value) (Value, error)

func init() {
	// Populated in init() (not as a var initializer) to avoid a false
	// initialization-cycle error: the method expressions transitively reference
	// evalCallExpr, which reads this table.
	interpBuiltins = map[string]func(*Interpreter, []Value) (Value, error){
		"print":     (*Interpreter).builtinPrint,
		"map":       (*Interpreter).builtinMap,
		"filter":    (*Interpreter).builtinFilter,
		"reduce":    (*Interpreter).builtinReduce,
		"find":      (*Interpreter).builtinFind,
		"findIndex": (*Interpreter).builtinFindIndex,
		"some":      (*Interpreter).builtinSome,
		"every":     (*Interpreter).builtinEvery,
		"flatMap":   (*Interpreter).builtinFlatMap,
	}
}

func (i *Interpreter) evalCallExpr(expr *ast.CallExpr) (Value, error) {
	funcExpr := expr.Function

	// Method-call syntax: receiver.method(args)
	if pa, ok := funcExpr.(*ast.PropertyAccessExpr); ok {
		return i.evalMethodCall(pa, expr.Arguments)
	}

	// Interpreter builtins (print, map, filter, ...), unless overridden by a
	// user binding or a registered native of the same name.
	if ident, ok := funcExpr.(*ast.Identifier); ok {
		if b, isBuiltin := interpBuiltins[ident.Value]; isBuiltin {
			_, inEnv := i.env.Get(ident.Value)
			if !inEnv && i.natives.Get(ident.Value) == nil {
				args, err := i.evalArgs(expr.Arguments)
				if err != nil {
					return nil, err
				}
				return b(i, args)
			}
		}
	}

	// Default: evaluate the callee then apply it (covers user functions and
	// registry natives resolved through evalExpression).
	function, err := i.evalExpression(funcExpr)
	if err != nil {
		return nil, err
	}
	args, err := i.evalArgs(expr.Arguments)
	if err != nil {
		return nil, err
	}
	return i.applyFunction(function, args)
}

// evalMethodCall implements method-call syntax: receiver.method(args).
func (i *Interpreter) evalMethodCall(pa *ast.PropertyAccessExpr, argExprs []ast.Expression) (Value, error) {
	receiver, err := i.evalExpression(pa.Object)
	if err != nil {
		return nil, err
	}
	method := pa.Property.Value

	args, err := i.evalArgs(argExprs)
	if err != nil {
		return nil, err
	}

	// 1. A callable stored under that key on an object wins (obj.fn()).
	if m, ok := receiver.(map[string]interface{}); ok {
		if v, present := m[method]; present && isCallable(v) {
			return i.applyFunction(v, args)
		}
	}

	// 2. Interpreter builtin invoked as a method: prepend the receiver.
	if b, ok := interpBuiltins[method]; ok {
		return b(i, prepend(receiver, args))
	}

	// 3. Registry native invoked as a method: prepend the receiver.
	if nfn := i.natives.Get(method); nfn != nil {
		return i.applyNative(nfn, prepend(receiver, args))
	}

	// 4. Bound Go object: method/field via reflection.
	if receiver == nil {
		return nil, fmt.Errorf("cannot call method '%s' on null", method)
	}
	if _, isMap := receiver.(map[string]interface{}); !isMap {
		fnVal, rerr := i.reflectivePropertyAccess(receiver, method)
		if rerr != nil {
			return nil, rerr
		}
		return i.applyFunction(fnVal, args)
	}

	return nil, fmt.Errorf("undefined method '%s'", method)
}

// evalArgs evaluates a list of argument expressions (expanding spreads).
func (i *Interpreter) evalArgs(argExprs []ast.Expression) ([]Value, error) {
	elems, err := i.evalElements(argExprs)
	if err != nil {
		return nil, err
	}
	args := make([]Value, len(elems))
	for idx, e := range elems {
		args[idx] = e
	}
	return args, nil
}

// evalElements evaluates a list of expressions, expanding any ...spread elements.
func (i *Interpreter) evalElements(exprs []ast.Expression) ([]interface{}, error) {
	result := make([]interface{}, 0, len(exprs))
	for _, el := range exprs {
		if sp, ok := el.(*ast.SpreadExpr); ok {
			v, err := i.evalExpression(sp.Value)
			if err != nil {
				return nil, err
			}
			arr, ok := v.([]interface{})
			if !ok {
				return nil, fmt.Errorf("spread operator requires an array, got %T", v)
			}
			result = append(result, arr...)
		} else {
			v, err := i.evalExpression(el)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}
	}
	return result, nil
}

// applyNative calls a registry native function with KodiScript values.
func (i *Interpreter) applyNative(fn natives.NativeFunc, args []Value) (Value, error) {
	ifaceArgs := make([]interface{}, len(args))
	for idx, a := range args {
		ifaceArgs[idx] = a
	}
	return fn(ifaceArgs...)
}

func isCallable(v Value) bool {
	switch v.(type) {
	case *Function, *NativeFunction:
		return true
	}
	return false
}

func prepend(head Value, tail []Value) []Value {
	out := make([]Value, 0, len(tail)+1)
	out = append(out, head)
	return append(out, tail...)
}

func (i *Interpreter) applyFunction(fn Value, args []Value) (Value, error) {
	switch function := fn.(type) {
	case *Function:
		if i.callDepth >= maxCallDepth {
			return nil, ErrStackOverflow
		}
		i.callDepth++
		defer func() { i.callDepth-- }()

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
		// A stray break/continue must not escape the function as a value.
		switch val.(type) {
		case *BreakValue, *ContinueValue:
			return nil, nil
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

// ============ Interpreter builtins (need interpreter context) ============

// builtinPrint writes each argument to stdout (unless silenced) and captures it.
func (i *Interpreter) builtinPrint(args []Value) (Value, error) {
	for _, arg := range args {
		output := toString(arg)
		if i.outputSink != nil {
			i.outputSink(output)
		} else if !i.silentPrint {
			fmt.Println(output)
		}
		i.output = append(i.output, output)
	}
	return nil, nil
}

func (i *Interpreter) builtinMap(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("map requires 2 arguments: array and function")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return []interface{}{}, nil
	}
	fnVal := args[1]
	result := make([]interface{}, len(arr))
	for idx, item := range arr {
		val, err := i.applyFunction(fnVal, []Value{item, float64(idx)})
		if err != nil {
			return nil, err
		}
		result[idx] = val
	}
	return result, nil
}

func (i *Interpreter) builtinFilter(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("filter requires 2 arguments: array and function")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return []interface{}{}, nil
	}
	fnVal := args[1]
	result := []interface{}{}
	for idx, item := range arr {
		val, err := i.applyFunction(fnVal, []Value{item, float64(idx)})
		if err != nil {
			return nil, err
		}
		if isTruthy(val) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (i *Interpreter) builtinReduce(args []Value) (Value, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("reduce requires 3 arguments: array, function, and initial value")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, nil
	}
	fnVal := args[1]
	accumulator := args[2]
	var err error
	for idx, item := range arr {
		accumulator, err = i.applyFunction(fnVal, []Value{accumulator, item, float64(idx)})
		if err != nil {
			return nil, err
		}
	}
	return accumulator, nil
}

func (i *Interpreter) builtinFind(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("find requires 2 arguments: array and function")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, nil
	}
	fnVal := args[1]
	for idx, item := range arr {
		val, err := i.applyFunction(fnVal, []Value{item, float64(idx)})
		if err != nil {
			return nil, err
		}
		if isTruthy(val) {
			return item, nil
		}
	}
	return nil, nil
}

func (i *Interpreter) builtinSome(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("some requires 2 arguments: array and function")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return false, nil
	}
	fnVal := args[1]
	for idx, item := range arr {
		val, err := i.applyFunction(fnVal, []Value{item, float64(idx)})
		if err != nil {
			return nil, err
		}
		if isTruthy(val) {
			return true, nil
		}
	}
	return false, nil
}

func (i *Interpreter) builtinEvery(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("every requires 2 arguments: array and function")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return true, nil
	}
	fnVal := args[1]
	for idx, item := range arr {
		val, err := i.applyFunction(fnVal, []Value{item, float64(idx)})
		if err != nil {
			return nil, err
		}
		if !isTruthy(val) {
			return false, nil
		}
	}
	return true, nil
}

func (i *Interpreter) builtinFlatMap(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("flatMap requires 2 arguments: array and function")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return []interface{}{}, nil
	}
	fnVal := args[1]
	result := []interface{}{}
	for idx, item := range arr {
		val, err := i.applyFunction(fnVal, []Value{item, float64(idx)})
		if err != nil {
			return nil, err
		}
		if inner, ok := val.([]interface{}); ok {
			result = append(result, inner...)
		} else {
			result = append(result, val)
		}
	}
	return result, nil
}

func (i *Interpreter) builtinFindIndex(args []Value) (Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("findIndex requires 2 arguments: array and function")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return float64(-1), nil
	}
	fnVal := args[1]
	for idx, item := range arr {
		val, err := i.applyFunction(fnVal, []Value{item, float64(idx)})
		if err != nil {
			return nil, err
		}
		if isTruthy(val) {
			return float64(idx), nil
		}
	}
	return float64(-1), nil
}
