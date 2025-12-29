// Package natives provides built-in functions for KodiScript.
package natives

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
	r.funcs["trim"] = nativeTrim
	r.funcs["split"] = nativeSplit
	r.funcs["join"] = nativeJoin
	r.funcs["replace"] = nativeReplace
	r.funcs["contains"] = nativeContains
	r.funcs["startsWith"] = nativeStartsWith
	r.funcs["endsWith"] = nativeEndsWith
	r.funcs["indexOf"] = nativeIndexOf

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
	r.funcs["isNumber"] = nativeIsNumber
	r.funcs["isString"] = nativeIsString
	r.funcs["isBool"] = nativeIsBool

	// Math functions
	r.funcs["abs"] = nativeAbs
	r.funcs["floor"] = nativeFloor
	r.funcs["ceil"] = nativeCeil
	r.funcs["round"] = nativeRound
	r.funcs["min"] = nativeMin
	r.funcs["max"] = nativeMax
	r.funcs["pow"] = nativePow
	r.funcs["sqrt"] = nativeSqrt
	r.funcs["sin"] = nativeSin
	r.funcs["cos"] = nativeCos
	r.funcs["tan"] = nativeTan
	r.funcs["log"] = nativeLog
	r.funcs["log10"] = nativeLog10
	r.funcs["exp"] = nativeExp

	// Random functions
	r.funcs["random"] = nativeRandom
	r.funcs["randomInt"] = nativeRandomInt
	r.funcs["randomUUID"] = nativeRandomUUID

	// Crypto/Hash functions
	r.funcs["md5"] = nativeMd5
	r.funcs["sha1"] = nativeSha1
	r.funcs["sha256"] = nativeSha256

	// Array functions
	r.funcs["sort"] = nativeSort
	r.funcs["sortBy"] = nativeSortBy
	r.funcs["reverse"] = nativeReverse
	r.funcs["size"] = nativeSize
	r.funcs["first"] = nativeFirst
	r.funcs["last"] = nativeLast
	r.funcs["slice"] = nativeSlice
}

// ============ String functions ============

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
		return strings.ToUpper(s), nil
	}
	return nil, fmt.Errorf("toUpperCase requires a string argument")
}

func nativeToLowerCase(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toLowerCase requires 1 argument")
	}
	if s, ok := args[0].(string); ok {
		return strings.ToLower(s), nil
	}
	return nil, fmt.Errorf("toLowerCase requires a string argument")
}

func nativeTrim(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("trim requires 1 argument")
	}
	if s, ok := args[0].(string); ok {
		return strings.TrimSpace(s), nil
	}
	return nil, fmt.Errorf("trim requires a string argument")
}

func nativeSplit(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("split requires 2 arguments")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("split requires a string as first argument")
	}
	sep, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("split requires a string as second argument")
	}
	parts := strings.Split(s, sep)
	result := make([]interface{}, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result, nil
}

func nativeJoin(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("join requires 2 arguments")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("join requires an array as first argument")
	}
	sep, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("join requires a string as second argument")
	}
	strs := make([]string, len(arr))
	for i, v := range arr {
		strs[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(strs, sep), nil
}

func nativeReplace(args ...interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("replace requires 3 arguments")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("replace requires a string as first argument")
	}
	old, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("replace requires a string as second argument")
	}
	new, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("replace requires a string as third argument")
	}
	return strings.ReplaceAll(s, old, new), nil
}

func nativeContains(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("contains requires 2 arguments")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("contains requires a string as first argument")
	}
	substr, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("contains requires a string as second argument")
	}
	return strings.Contains(s, substr), nil
}

func nativeStartsWith(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("startsWith requires 2 arguments")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("startsWith requires a string as first argument")
	}
	prefix, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("startsWith requires a string as second argument")
	}
	return strings.HasPrefix(s, prefix), nil
}

func nativeEndsWith(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("endsWith requires 2 arguments")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("endsWith requires a string as first argument")
	}
	suffix, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("endsWith requires a string as second argument")
	}
	return strings.HasSuffix(s, suffix), nil
}

func nativeIndexOf(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("indexOf requires 2 arguments")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("indexOf requires a string as first argument")
	}
	substr, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("indexOf requires a string as second argument")
	}
	return float64(strings.Index(s, substr)), nil
}

// ============ JSON functions ============

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

// ============ Base64 functions ============

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

// ============ URL functions ============

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

// ============ Type functions ============

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
	case []interface{}:
		return "array", nil
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

func nativeIsNumber(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isNumber requires 1 argument")
	}
	switch args[0].(type) {
	case float64, int, int64:
		return true, nil
	default:
		return false, nil
	}
}

func nativeIsString(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isString requires 1 argument")
	}
	_, ok := args[0].(string)
	return ok, nil
}

func nativeIsBool(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("isBool requires 1 argument")
	}
	_, ok := args[0].(bool)
	return ok, nil
}

// ============ Math functions ============

func nativeAbs(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("abs requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("abs requires a number argument")
	}
	return math.Abs(n), nil
}

func nativeFloor(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("floor requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("floor requires a number argument")
	}
	return math.Floor(n), nil
}

func nativeCeil(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ceil requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("ceil requires a number argument")
	}
	return math.Ceil(n), nil
}

func nativeRound(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("round requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("round requires a number argument")
	}
	return math.Round(n), nil
}

func nativeMin(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("min requires at least 2 arguments")
	}
	result, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("min requires number arguments")
	}
	for i := 1; i < len(args); i++ {
		n, ok := toFloat(args[i])
		if !ok {
			return nil, fmt.Errorf("min requires number arguments")
		}
		if n < result {
			result = n
		}
	}
	return result, nil
}

func nativeMax(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("max requires at least 2 arguments")
	}
	result, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("max requires number arguments")
	}
	for i := 1; i < len(args); i++ {
		n, ok := toFloat(args[i])
		if !ok {
			return nil, fmt.Errorf("max requires number arguments")
		}
		if n > result {
			result = n
		}
	}
	return result, nil
}

func nativePow(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("pow requires 2 arguments")
	}
	base, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("pow requires number arguments")
	}
	exp, ok := toFloat(args[1])
	if !ok {
		return nil, fmt.Errorf("pow requires number arguments")
	}
	return math.Pow(base, exp), nil
}

func nativeSqrt(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sqrt requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("sqrt requires a number argument")
	}
	if n < 0 {
		return nil, fmt.Errorf("sqrt of negative number")
	}
	return math.Sqrt(n), nil
}

func nativeSin(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sin requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("sin requires a number argument")
	}
	return math.Sin(n), nil
}

func nativeCos(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("cos requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("cos requires a number argument")
	}
	return math.Cos(n), nil
}

func nativeTan(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tan requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("tan requires a number argument")
	}
	return math.Tan(n), nil
}

func nativeLog(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("log requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("log requires a number argument")
	}
	if n <= 0 {
		return nil, fmt.Errorf("log of non-positive number")
	}
	return math.Log(n), nil
}

func nativeLog10(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("log10 requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("log10 requires a number argument")
	}
	if n <= 0 {
		return nil, fmt.Errorf("log10 of non-positive number")
	}
	return math.Log10(n), nil
}

func nativeExp(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("exp requires 1 argument")
	}
	n, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("exp requires a number argument")
	}
	return math.Exp(n), nil
}

// ============ Random functions ============

func nativeRandom(args ...interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("random takes no arguments")
	}
	return rand.Float64(), nil
}

func nativeRandomInt(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("randomInt requires 2 arguments (min, max)")
	}
	min, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("randomInt requires number arguments")
	}
	max, ok := toFloat(args[1])
	if !ok {
		return nil, fmt.Errorf("randomInt requires number arguments")
	}
	if min >= max {
		return nil, fmt.Errorf("randomInt: min must be less than max")
	}
	return float64(rand.Intn(int(max)-int(min)+1) + int(min)), nil
}

func nativeRandomUUID(args ...interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("randomUUID takes no arguments")
	}
	// Generate a simple UUID v4
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// ============ Crypto/Hash functions ============

func nativeMd5(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("md5 requires 1 argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("md5 requires a string argument")
	}
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:]), nil
}

func nativeSha1(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sha1 requires 1 argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("sha1 requires a string argument")
	}
	hash := sha1.Sum([]byte(s))
	return hex.EncodeToString(hash[:]), nil
}

func nativeSha256(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sha256 requires 1 argument")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("sha256 requires a string argument")
	}
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:]), nil
}

// ============ Utility ============

func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

// ============ Array functions ============

func nativeSort(args ...interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("sort requires 1 or 2 arguments (array, [order])")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("sort requires an array as first argument")
	}

	// Determine order: "asc" (default) or "desc"
	ascending := true
	if len(args) == 2 {
		order, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("sort requires a string as second argument (asc/desc)")
		}
		if order == "desc" {
			ascending = false
		}
	}

	// Create a copy to avoid mutating the original
	result := make([]interface{}, len(arr))
	copy(result, arr)

	// Sort using bubble sort (simple, works for mixed types)
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			should := compareValues(result[j], result[j+1])
			if ascending {
				if should > 0 {
					result[j], result[j+1] = result[j+1], result[j]
				}
			} else {
				if should < 0 {
					result[j], result[j+1] = result[j+1], result[j]
				}
			}
		}
	}

	return result, nil
}

func nativeSortBy(args ...interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("sortBy requires 2 or 3 arguments (array, field, [order])")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("sortBy requires an array as first argument")
	}
	field, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("sortBy requires a string field name as second argument")
	}

	// Determine order
	ascending := true
	if len(args) == 3 {
		order, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("sortBy requires a string as third argument (asc/desc)")
		}
		if order == "desc" {
			ascending = false
		}
	}

	// Create a copy
	result := make([]interface{}, len(arr))
	copy(result, arr)

	// Sort by field
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			val1 := getFieldValue(result[j], field)
			val2 := getFieldValue(result[j+1], field)
			should := compareValues(val1, val2)
			if ascending {
				if should > 0 {
					result[j], result[j+1] = result[j+1], result[j]
				}
			} else {
				if should < 0 {
					result[j], result[j+1] = result[j+1], result[j]
				}
			}
		}
	}

	return result, nil
}

func nativeReverse(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("reverse requires 1 argument")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("reverse requires an array argument")
	}

	result := make([]interface{}, len(arr))
	for i, v := range arr {
		result[len(arr)-1-i] = v
	}
	return result, nil
}

func nativeSize(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("size requires 1 argument")
	}
	switch v := args[0].(type) {
	case []interface{}:
		return float64(len(v)), nil
	case string:
		return float64(len(v)), nil
	case map[string]interface{}:
		return float64(len(v)), nil
	default:
		return nil, fmt.Errorf("size requires an array, string, or object")
	}
}

func nativeFirst(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("first requires 1 argument")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("first requires an array argument")
	}
	if len(arr) == 0 {
		return nil, nil
	}
	return arr[0], nil
}

func nativeLast(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("last requires 1 argument")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("last requires an array argument")
	}
	if len(arr) == 0 {
		return nil, nil
	}
	return arr[len(arr)-1], nil
}

func nativeSlice(args ...interface{}) (interface{}, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("slice requires 2 or 3 arguments (array, start, [end])")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("slice requires an array as first argument")
	}
	start, ok := toFloat(args[1])
	if !ok {
		return nil, fmt.Errorf("slice requires a number as second argument")
	}
	startIdx := int(start)
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(arr) {
		return []interface{}{}, nil
	}

	if len(args) == 3 {
		end, ok := toFloat(args[2])
		if !ok {
			return nil, fmt.Errorf("slice requires a number as third argument")
		}
		endIdx := int(end)
		if endIdx > len(arr) {
			endIdx = len(arr)
		}
		if endIdx < startIdx {
			return []interface{}{}, nil
		}
		return arr[startIdx:endIdx], nil
	}

	return arr[startIdx:], nil
}

// Helper: compare two values
func compareValues(a, b interface{}) int {
	// Handle nil
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Try numeric comparison
	aNum, aOk := toFloat(a)
	bNum, bOk := toFloat(b)
	if aOk && bOk {
		if aNum < bNum {
			return -1
		}
		if aNum > bNum {
			return 1
		}
		return 0
	}

	// Try string comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return strings.Compare(aStr, bStr)
}

// Helper: get field value from object
func getFieldValue(obj interface{}, field string) interface{} {
	if m, ok := obj.(map[string]interface{}); ok {
		return m[field]
	}
	return nil
}
