// Package interpreter evaluates KodiScript AST nodes.
package interpreter

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// memberKey identifies a member lookup (concrete type + name).
type memberKey struct {
	typ  reflect.Type
	name string
}

// memberInfo caches how a name resolves on a type, so the method/field scan
// runs once per (type, name) instead of on every property access.
type memberInfo struct {
	isMethod   bool
	methodPtr  bool // method belongs to the pointer method set
	isField    bool
	fieldIndex []int // index path for FieldByIndex
}

var memberCache sync.Map // map[memberKey]memberInfo

// resolveMember returns (cached) routing info for accessing name on val's type.
func resolveMember(val reflect.Value, name string) memberInfo {
	key := memberKey{val.Type(), name}
	if cached, ok := memberCache.Load(key); ok {
		return cached.(memberInfo)
	}

	info := memberInfo{}
	if _, ok := val.Type().MethodByName(name); ok {
		info.isMethod = true
	} else if val.CanAddr() {
		if _, ok := reflect.PtrTo(val.Type()).MethodByName(name); ok {
			info.isMethod = true
			info.methodPtr = true
		}
	}
	if !info.isMethod && val.Kind() == reflect.Struct {
		if f, ok := val.Type().FieldByName(name); ok && f.PkgPath == "" {
			info.isField = true
			info.fieldIndex = f.Index
		}
	}

	memberCache.Store(key, info)
	return info
}

// enginePkgPrefix is the import path of the KodiScript engine module root,
// derived from this package's own path so a fork or vendored copy keeps the
// guard working. The engine's internal packages all sit beneath it.
var enginePkgPrefix = func() string {
	p := reflect.TypeOf(Environment{}).PkgPath() // ".../interpreter"
	if idx := strings.LastIndex(p, "/"); idx >= 0 {
		return p[:idx]
	}
	return p
}()

// isEngineType reports whether t (after dereferencing pointers) is one of the
// engine's own internal runtime types — anything in a sub-package of the engine
// module, such as interpreter.Function, interpreter.Environment, or the ast
// nodes reachable through Function.Body. The module root package is deliberately
// not treated as internal: it holds the public API and may legitimately hold a
// host's bound types.
func isEngineType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pkg := t.PkgPath()
	return pkg != "" && strings.HasPrefix(pkg, enginePkgPrefix+"/")
}

// reflectivePropertyAccess uses reflection to access properties on Go objects.
func (i *Interpreter) reflectivePropertyAccess(object Value, propertyName string) (Value, error) {
	val := reflect.ValueOf(object)

	// Dereference pointers
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("cannot access property '%s' on nil pointer", propertyName)
		}
		val = val.Elem()
	}

	// Refuse reflection into the engine's own internal types. A user function is
	// an *interpreter.Function whose Env field is exported, and *Environment
	// exposes Get/Set/Release; without this barrier a script could read and
	// rewrite the interpreter's environments and, through Release, poison the
	// shared environment pool for later executions in the same process.
	// Reflection is only meant for host objects bound by the caller, which never
	// live in the engine's own packages. The error is identical to a missing
	// member, so a probing script cannot tell an internal type from an absent
	// property.
	if isEngineType(val.Type()) {
		return nil, fmt.Errorf("property or method '%s' not found on %T", propertyName, object)
	}

	info := resolveMember(val, propertyName)

	// Methods have priority over fields.
	if info.isMethod {
		var method reflect.Value
		if info.methodPtr {
			method = val.Addr().MethodByName(propertyName)
		} else {
			method = val.MethodByName(propertyName)
		}
		return &NativeFunction{
			Fn: func(args ...interface{}) (interface{}, error) {
				return callReflectedMethod(method, args)
			},
		}, nil
	}

	if info.isField {
		field := val.FieldByIndex(info.fieldIndex)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), nil
		}
	}

	return nil, fmt.Errorf("property or method '%s' not found on %T", propertyName, object)
}

// callReflectedMethod calls a Go method via reflection, converting args appropriately.
func callReflectedMethod(method reflect.Value, args []interface{}) (interface{}, error) {
	methodType := method.Type()
	numIn := methodType.NumIn()

	// Prepare arguments
	in := make([]reflect.Value, 0, numIn)
	for i := 0; i < numIn; i++ {
		if i >= len(args) {
			// Not enough arguments provided
			return nil, fmt.Errorf("not enough arguments: expected %d, got %d", numIn, len(args))
		}

		argType := methodType.In(i)
		convertedArg, err := convertToGoType(args[i], argType)
		if err != nil {
			return nil, fmt.Errorf("argument %d: %w", i, err)
		}
		in = append(in, convertedArg)
	}

	// Call the method
	out := method.Call(in)

	// Process return values
	switch len(out) {
	case 0:
		return nil, nil
	case 1:
		return convertFromGoType(out[0]), nil
	case 2:
		// Check if second return is error
		if out[1].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !out[1].IsNil() {
				return nil, out[1].Interface().(error)
			}
			return convertFromGoType(out[0]), nil
		}
		// Return both values as array
		return []interface{}{convertFromGoType(out[0]), convertFromGoType(out[1])}, nil
	default:
		// Multiple return values - return as array
		results := make([]interface{}, len(out))
		for i, v := range out {
			results[i] = convertFromGoType(v)
		}
		return results, nil
	}
}

// convertToGoType converts a KodiScript value to the target Go type.
func convertToGoType(val interface{}, targetType reflect.Type) (reflect.Value, error) {
	if val == nil {
		// Return zero value for the type
		return reflect.Zero(targetType), nil
	}

	valType := reflect.TypeOf(val)

	// If types match exactly, use directly
	if valType.AssignableTo(targetType) {
		return reflect.ValueOf(val), nil
	}

	// Handle numeric conversions
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if f, ok := val.(float64); ok {
			return reflect.ValueOf(int(f)).Convert(targetType), nil
		}
		if i, ok := val.(int); ok {
			return reflect.ValueOf(i).Convert(targetType), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if f, ok := val.(float64); ok {
			return reflect.ValueOf(uint(f)).Convert(targetType), nil
		}
	case reflect.Float32, reflect.Float64:
		if f, ok := val.(float64); ok {
			return reflect.ValueOf(f).Convert(targetType), nil
		}
		if i, ok := val.(int); ok {
			return reflect.ValueOf(float64(i)).Convert(targetType), nil
		}
	case reflect.String:
		if s, ok := val.(string); ok {
			return reflect.ValueOf(s), nil
		}
	case reflect.Bool:
		if b, ok := val.(bool); ok {
			return reflect.ValueOf(b), nil
		}
	}

	// Try to convert directly
	valReflect := reflect.ValueOf(val)
	if valReflect.Type().ConvertibleTo(targetType) {
		return valReflect.Convert(targetType), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", val, targetType)
}

// convertFromGoType converts a Go reflect.Value back to a KodiScript-compatible value.
func convertFromGoType(val reflect.Value) interface{} {
	if !val.IsValid() {
		return nil
	}

	// Handle nil pointers
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return nil
	}

	if !val.CanInterface() {
		return nil
	}

	result := val.Interface()

	// Convert Go ints to float64 (KodiScript's number type)
	switch v := result.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	}

	return result
}
