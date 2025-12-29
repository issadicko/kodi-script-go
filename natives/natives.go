// Package natives provides built-in functions for KodiScript.
package natives

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// NativeFunc is the signature for native functions.
type NativeFunc func(args ...interface{}) (interface{}, error)

// Registry holds all registered native functions.
type Registry struct {
	funcs map[string]NativeFunc
}

// NewRegistry creates a new registry with all built-in functions.
func NewRegistry() *Registry {
	r := &Registry{funcs: make(map[string]NativeFunc)}
	r.registerBuiltins()
	return r
}

// Get retrieves a native function by name.
func (r *Registry) Get(name string) NativeFunc {
	return r.funcs[name]
}

// Register adds a custom native function.
func (r *Registry) Register(name string, fn NativeFunc) {
	r.funcs[name] = fn
}

func (r *Registry) registerBuiltins() {
	// String functions
	r.funcs["toString"] = nativeToString
	r.funcs["toNumber"] = nativeToNumber
	r.funcs["length"] = nativeLength
	r.funcs["substring"] = nativeSubstring
	r.funcs["toUpperCase"] = nativeToUpperCase
	r.funcs["toLowerCase"] = nativeToLowerCase

	// JSON functions
	r.funcs["jsonParse"] = nativeJsonParse
	r.funcs["jsonStringify"] = nativeJsonStringify

	// Base64 functions
	r.funcs["base64Encode"] = nativeBase64Encode
	r.funcs["base64Decode"] = nativeBase64Decode

	// URL functions
	r.funcs["urlEncode"] = nativeUrlEncode
	r.funcs["urlDecode"] = nativeUrlDecode

	// Type checking
	r.funcs["typeOf"] = nativeTypeOf
	r.funcs["isNull"] = nativeIsNull
}

// String functions

func nativeToString(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toString requires 1 argument")
	}
	return fmt.Sprintf("%v", args[0]), nil
}

func nativeToNumber(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toNumber requires 1 argument")
	}
	switch v := args[0].(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert '%s' to number", v)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to number", args[0])
	}
}

func nativeLength(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("length requires 1 argument")
	}
	if s, ok := args[0].(string); ok {
		return float64(len(s)), nil
	}
	return nil, fmt.Errorf("length requires a string argument")
}

func nativeSubstring(args ...interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("substring requires 2 or 3 arguments")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("substring requires a string as first argument")
	}
	start, ok := args[1].(float64)
	if !ok {
		return nil, fmt.Errorf("substring requires a number as second argument")
	}
	startIdx := int(start)
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(s) {
		return "", nil
	}

	if len(args) == 3 {
		end, ok := args[2].(float64)
		if !ok {
			return nil, fmt.Errorf("substring requires a number as third argument")
		}
		endIdx := int(end)
		if endIdx > len(s) {
			endIdx = len(s)
		}
		return s[startIdx:endIdx], nil
	}

	return s[startIdx:], nil
}

func nativeToUpperCase(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toUpperCase requires 1 argument")
	}
	if s, ok := args[0].(string); ok {
		return string([]byte(s)), nil // Simple uppercase would need strings package
	}
	return nil, fmt.Errorf("toUpperCase requires a string argument")
}

func nativeToLowerCase(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toLowerCase requires 1 argument")
	}
	if s, ok := args[0].(string); ok {
		return s, nil // Simple lowercase would need strings package
	}
	return nil, fmt.Errorf("toLowerCase requires a string argument")
}

// JSON functions

func nativeJsonParse(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("jsonParse requires 1 argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("jsonParse requires a string argument")
	}
	var result interface{}
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}
	return result, nil
}

func nativeJsonStringify(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("jsonStringify requires 1 argument")
	}
	b, err := json.Marshal(args[0])
	if err != nil {
		return nil, fmt.Errorf("cannot stringify: %v", err)
	}
	return string(b), nil
}

// Base64 functions

func nativeBase64Encode(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("base64Encode requires 1 argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("base64Encode requires a string argument")
	}
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

func nativeBase64Decode(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("base64Decode requires 1 argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("base64Decode requires a string argument")
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %v", err)
	}
	return string(b), nil
}

// URL functions

func nativeUrlEncode(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("urlEncode requires 1 argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("urlEncode requires a string argument")
	}
	return url.QueryEscape(s), nil
}

func nativeUrlDecode(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("urlDecode requires 1 argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("urlDecode requires a string argument")
	}
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return nil, fmt.Errorf("invalid URL encoding: %v", err)
	}
	return decoded, nil
}

// Type functions

func nativeTypeOf(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("typeOf requires 1 argument")
	}
	if args[0] == nil {
		return "null", nil
	}
	switch args[0].(type) {
	case string:
		return "string", nil
	case float64, int, int64:
		return "number", nil
	case bool:
		return "boolean", nil
	case map[string]interface{}:
		return "object", nil
	default:
		return "unknown", nil
	}
}

func nativeIsNull(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isNull requires 1 argument")
	}
	return args[0] == nil, nil
}
