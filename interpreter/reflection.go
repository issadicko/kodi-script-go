// Package interpreter evaluates KodiScript AST nodes.
package interpreter

import (
	"fmt"
	"reflect"
)

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

	// Try to find method first (methods have priority over fields)
	method := findMethod(val, propertyName)
	if method.IsValid() {
		// Return a wrapper function that can be called from KodiScript
		return &NativeFunction{
			Fn: func(args ...interface{}) (interface{}, error) {
				return callReflectedMethod(method, args)
			},
		}, nil
	}

	// Try to access field
	if val.Kind() == reflect.Struct {
		field := val.FieldByName(propertyName)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), nil
		}
	}

	return nil, fmt.Errorf("property or method '%s' not found on %T", propertyName, object)
}

// findMethod tries to find a method on a value or its pointer type.
func findMethod(val reflect.Value, name string) reflect.Value {
	// Try on the value itself
	method := val.MethodByName(name)
	if method.IsValid() {
		return method
	}

	// Try on the pointer type if not already a pointer
	if val.CanAddr() {
		method = val.Addr().MethodByName(name)
		if method.IsValid() {
			return method
		}
	}

	return reflect.Value{}
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
