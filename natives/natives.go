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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// NativeFunc is the signature for native functions.
type NativeFunc func(args ...interface{}) (interface{}, error)

// Stringify renders a KodiScript value in canonical form. It is shared across
// the interpreter (print, string templates) and natives (toString, join) so
// that output is identical across language implementations:
//   - integral numbers print without a trailing ".0" (3, not 3.0)
//   - arrays print as "[1, 2, 3]"
//   - objects print as "{a: 1, b: 2}" with keys sorted for determinism
//   - strings are not quoted (use jsonStringify for quoted output)
func Stringify(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case []interface{}:
		parts := make([]string, len(val))
		for i, e := range val {
			parts[i] = Stringify(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		keys := sortedKeys(val)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = k + ": " + Stringify(val[k])
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Registry holds all registered native functions.
type Registry struct {
	funcs map[string]NativeFunc
}

// DefaultBuiltins is a global read-only registry containing all built-in functions.
// It is shared across all script executions for memory efficiency.
var DefaultBuiltins = newBuiltinRegistry()

// newBuiltinRegistry creates a registry with all built-in functions (internal).
func newBuiltinRegistry() *Registry {
	r := &Registry{funcs: make(map[string]NativeFunc)}
	r.registerBuiltins()
	return r
}

// NewRegistry creates a new empty registry for custom functions.
// Custom functions are per-script and take priority over builtins.
func NewRegistry() *Registry {
	return &Registry{funcs: make(map[string]NativeFunc)}
}

// Get retrieves a native function by name from customs first, then builtins.
func (r *Registry) Get(name string) NativeFunc {
	if fn, ok := r.funcs[name]; ok {
		return fn // Custom takes priority
	}
	// Fallback to global builtins
	if r != DefaultBuiltins {
		return DefaultBuiltins.funcs[name]
	}
	return nil
}

// Register adds a custom native function to this registry.
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
	r.funcs["padLeft"] = nativePadLeft
	r.funcs["padRight"] = nativePadRight
	r.funcs["repeat"] = nativeRepeat

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
	r.funcs["range"] = nativeRange
	r.funcs["sum"] = nativeSum
	r.funcs["avg"] = nativeAvg
	r.funcs["unique"] = nativeUnique
	r.funcs["flatten"] = nativeFlatten
	r.funcs["push"] = nativePush
	r.funcs["concat"] = nativeConcat

	// Object functions
	r.funcs["keys"] = nativeKeys
	r.funcs["values"] = nativeValues
	r.funcs["entries"] = nativeEntries
	r.funcs["has"] = nativeHas

	// Number parsing
	r.funcs["parseInt"] = nativeParseInt
	r.funcs["parseFloat"] = nativeParseFloat

	// Regex
	r.funcs["regexMatch"] = nativeRegexMatch
	r.funcs["regexReplace"] = nativeRegexReplace

	// Date/Time functions
	r.funcs["now"] = nativeNow
	r.funcs["date"] = nativeDate
	r.funcs["time"] = nativeTime
	r.funcs["datetime"] = nativeDatetime
	r.funcs["timestamp"] = nativeTimestamp
	r.funcs["formatDate"] = nativeFormatDate
	r.funcs["year"] = nativeYear
	r.funcs["month"] = nativeMonth
	r.funcs["day"] = nativeDay
	r.funcs["hour"] = nativeHour
	r.funcs["minute"] = nativeMinute
	r.funcs["second"] = nativeSecond
	r.funcs["dayOfWeek"] = nativeDayOfWeek
	r.funcs["addDays"] = nativeAddDays
	r.funcs["addHours"] = nativeAddHours
	r.funcs["diffDays"] = nativeDiffDays
}

// ============ String functions ============

func nativeToString(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("toString requires 1 argument")
	}
	return Stringify(args[0]), nil
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
		strs[i] = Stringify(v)
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

func nativePadLeft(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("padLeft requires at least 2 arguments")
	}
	s := Stringify(args[0])
	length := int(asFloat(args[1]))
	padChar := " "
	if len(args) > 2 && args[2] != nil {
		padChar = fmt.Sprintf("%v", args[2])
		if len(padChar) == 0 {
			padChar = " "
		}
	}
	for len(s) < length {
		s = padChar[:1] + s
	}
	return s, nil
}

func nativePadRight(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("padRight requires at least 2 arguments")
	}
	s := Stringify(args[0])
	length := int(asFloat(args[1]))
	padChar := " "
	if len(args) > 2 && args[2] != nil {
		padChar = fmt.Sprintf("%v", args[2])
		if len(padChar) == 0 {
			padChar = " "
		}
	}
	for len(s) < length {
		s = s + padChar[:1]
	}
	return s, nil
}

func nativeRepeat(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("repeat requires 2 arguments")
	}
	s := Stringify(args[0])
	count := int(asFloat(args[1]))
	if count < 0 {
		count = 0
	}
	return strings.Repeat(s, count), nil
}

func asFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	default:
		return 0
	}
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
	// Use PathEscape for RFC 3986 compliance (spaces as %20, not +)
	return url.PathEscape(s), nil
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

	// Stable sort (works for mixed types via compareValues)
	sort.SliceStable(result, func(i, j int) bool {
		cmp := compareValues(result[i], result[j])
		if ascending {
			return cmp < 0
		}
		return cmp > 0
	})

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

	// Stable sort by field
	sort.SliceStable(result, func(i, j int) bool {
		cmp := compareValues(getFieldValue(result[i], field), getFieldValue(result[j], field))
		if ascending {
			return cmp < 0
		}
		return cmp > 0
	})

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

// valueKey produces a comparison key for primitive values (used by unique/has).
func valueKey(v interface{}) string {
	return fmt.Sprintf("%T:%v", v, v)
}

func nativeRange(args ...interface{}) (interface{}, error) {
	var start, end int
	switch len(args) {
	case 1:
		n, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("range requires number arguments")
		}
		start, end = 0, int(n)
	case 2:
		s, ok1 := toFloat(args[0])
		e, ok2 := toFloat(args[1])
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("range requires number arguments")
		}
		start, end = int(s), int(e)
	default:
		return nil, fmt.Errorf("range requires 1 or 2 arguments")
	}
	result := []interface{}{}
	for i := start; i < end; i++ {
		result = append(result, float64(i))
	}
	return result, nil
}

func nativeSum(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sum requires 1 argument")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("sum requires an array argument")
	}
	total := 0.0
	for _, v := range arr {
		n, ok := toFloat(v)
		if !ok {
			return nil, fmt.Errorf("sum requires an array of numbers")
		}
		total += n
	}
	return total, nil
}

func nativeAvg(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("avg requires 1 argument")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("avg requires an array argument")
	}
	if len(arr) == 0 {
		return float64(0), nil
	}
	total := 0.0
	for _, v := range arr {
		n, ok := toFloat(v)
		if !ok {
			return nil, fmt.Errorf("avg requires an array of numbers")
		}
		total += n
	}
	return total / float64(len(arr)), nil
}

func nativeUnique(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("unique requires 1 argument")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unique requires an array argument")
	}
	seen := make(map[string]bool)
	result := []interface{}{}
	for _, v := range arr {
		k := valueKey(v)
		if !seen[k] {
			seen[k] = true
			result = append(result, v)
		}
	}
	return result, nil
}

func nativeFlatten(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("flatten requires 1 argument")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("flatten requires an array argument")
	}
	result := []interface{}{}
	for _, v := range arr {
		if inner, ok := v.([]interface{}); ok {
			result = append(result, inner...)
		} else {
			result = append(result, v)
		}
	}
	return result, nil
}

func nativePush(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("push requires at least 2 arguments (array, item...)")
	}
	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("push requires an array as first argument")
	}
	result := make([]interface{}, len(arr), len(arr)+len(args)-1)
	copy(result, arr)
	return append(result, args[1:]...), nil
}

func nativeConcat(args ...interface{}) (interface{}, error) {
	result := []interface{}{}
	for _, a := range args {
		arr, ok := a.([]interface{})
		if !ok {
			return nil, fmt.Errorf("concat requires array arguments")
		}
		result = append(result, arr...)
	}
	return result, nil
}

// ============ Object functions ============

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func nativeKeys(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("keys requires 1 argument")
	}
	m, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("keys requires an object argument")
	}
	keys := sortedKeys(m)
	result := make([]interface{}, len(keys))
	for i, k := range keys {
		result[i] = k
	}
	return result, nil
}

func nativeValues(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("values requires 1 argument")
	}
	m, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("values requires an object argument")
	}
	result := make([]interface{}, 0, len(m))
	for _, k := range sortedKeys(m) {
		result = append(result, m[k])
	}
	return result, nil
}

func nativeEntries(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("entries requires 1 argument")
	}
	m, ok := args[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("entries requires an object argument")
	}
	result := make([]interface{}, 0, len(m))
	for _, k := range sortedKeys(m) {
		result = append(result, []interface{}{k, m[k]})
	}
	return result, nil
}

func nativeHas(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("has requires 2 arguments")
	}
	switch coll := args[0].(type) {
	case map[string]interface{}:
		key, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("has requires a string key for objects")
		}
		_, present := coll[key]
		return present, nil
	case []interface{}:
		target := valueKey(args[1])
		for _, item := range coll {
			if valueKey(item) == target {
				return true, nil
			}
		}
		return false, nil
	default:
		return nil, fmt.Errorf("has requires an object or array as first argument")
	}
}

// ============ Number parsing ============

func nativeParseInt(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("parseInt requires 1 argument")
	}
	switch v := args[0].(type) {
	case float64:
		return math.Trunc(v), nil
	case int:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse '%s' as integer", v)
		}
		return math.Trunc(f), nil
	default:
		return nil, fmt.Errorf("parseInt requires a string or number")
	}
}

func nativeParseFloat(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("parseFloat requires 1 argument")
	}
	switch v := args[0].(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse '%s' as number", v)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("parseFloat requires a string or number")
	}
}

// ============ Regex ============

func nativeRegexMatch(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("regexMatch requires 2 arguments (string, pattern)")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("regexMatch requires a string as first argument")
	}
	pat, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("regexMatch requires a string pattern")
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %v", err)
	}
	return re.MatchString(s), nil
}

func nativeRegexReplace(args ...interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("regexReplace requires 3 arguments (string, pattern, replacement)")
	}
	s, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("regexReplace requires a string as first argument")
	}
	pat, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("regexReplace requires a string pattern")
	}
	repl, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("regexReplace requires a string replacement")
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %v", err)
	}
	return re.ReplaceAllString(s, repl), nil
}

// ============ Date/Time functions ============

func nativeNow(args ...interface{}) (interface{}, error) {
	return float64(time.Now().UnixMilli()), nil
}

func nativeDate(args ...interface{}) (interface{}, error) {
	return time.Now().Format("2006-01-02"), nil
}

func nativeTime(args ...interface{}) (interface{}, error) {
	return time.Now().Format("15:04:05"), nil
}

func nativeDatetime(args ...interface{}) (interface{}, error) {
	return time.Now().Format(time.RFC3339), nil
}

func nativeTimestamp(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return float64(time.Now().UnixMilli()), nil
	}
	dateStr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("timestamp requires a string argument")
	}
	// Try common formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return float64(t.UnixMilli()), nil
		}
	}
	return nil, fmt.Errorf("cannot parse date: %s", dateStr)
}

func nativeFormatDate(args ...interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("formatDate requires 1 or 2 arguments")
	}
	ts, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("formatDate requires a number as first argument")
	}
	t := time.UnixMilli(int64(ts))

	format := "YYYY-MM-DD"
	if len(args) == 2 {
		format, ok = args[1].(string)
		if !ok {
			return nil, fmt.Errorf("formatDate requires a string as second argument")
		}
	}

	// Convert simple format to Go layout
	result := format
	result = strings.ReplaceAll(result, "YYYY", t.Format("2006"))
	result = strings.ReplaceAll(result, "MM", t.Format("01"))
	result = strings.ReplaceAll(result, "DD", t.Format("02"))
	result = strings.ReplaceAll(result, "HH", t.Format("15"))
	result = strings.ReplaceAll(result, "mm", t.Format("04"))
	result = strings.ReplaceAll(result, "ss", t.Format("05"))
	return result, nil
}

func nativeYear(args ...interface{}) (interface{}, error) {
	t := time.Now()
	if len(args) > 0 {
		ts, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("year requires a number argument")
		}
		t = time.UnixMilli(int64(ts))
	}
	return float64(t.Year()), nil
}

func nativeMonth(args ...interface{}) (interface{}, error) {
	t := time.Now()
	if len(args) > 0 {
		ts, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("month requires a number argument")
		}
		t = time.UnixMilli(int64(ts))
	}
	return float64(t.Month()), nil
}

func nativeDay(args ...interface{}) (interface{}, error) {
	t := time.Now()
	if len(args) > 0 {
		ts, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("day requires a number argument")
		}
		t = time.UnixMilli(int64(ts))
	}
	return float64(t.Day()), nil
}

func nativeHour(args ...interface{}) (interface{}, error) {
	t := time.Now()
	if len(args) > 0 {
		ts, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("hour requires a number argument")
		}
		t = time.UnixMilli(int64(ts))
	}
	return float64(t.Hour()), nil
}

func nativeMinute(args ...interface{}) (interface{}, error) {
	t := time.Now()
	if len(args) > 0 {
		ts, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("minute requires a number argument")
		}
		t = time.UnixMilli(int64(ts))
	}
	return float64(t.Minute()), nil
}

func nativeSecond(args ...interface{}) (interface{}, error) {
	t := time.Now()
	if len(args) > 0 {
		ts, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("second requires a number argument")
		}
		t = time.UnixMilli(int64(ts))
	}
	return float64(t.Second()), nil
}

func nativeDayOfWeek(args ...interface{}) (interface{}, error) {
	t := time.Now()
	if len(args) > 0 {
		ts, ok := toFloat(args[0])
		if !ok {
			return nil, fmt.Errorf("dayOfWeek requires a number argument")
		}
		t = time.UnixMilli(int64(ts))
	}
	return float64(t.Weekday()), nil
}

func nativeAddDays(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("addDays requires 2 arguments (timestamp, days)")
	}
	ts, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("addDays requires a number as first argument")
	}
	days, ok := toFloat(args[1])
	if !ok {
		return nil, fmt.Errorf("addDays requires a number as second argument")
	}
	t := time.UnixMilli(int64(ts))
	t = t.AddDate(0, 0, int(days))
	return float64(t.UnixMilli()), nil
}

func nativeAddHours(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("addHours requires 2 arguments (timestamp, hours)")
	}
	ts, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("addHours requires a number as first argument")
	}
	hours, ok := toFloat(args[1])
	if !ok {
		return nil, fmt.Errorf("addHours requires a number as second argument")
	}
	t := time.UnixMilli(int64(ts))
	t = t.Add(time.Duration(hours) * time.Hour)
	return float64(t.UnixMilli()), nil
}

func nativeDiffDays(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("diffDays requires 2 arguments (timestamp1, timestamp2)")
	}
	ts1, ok := toFloat(args[0])
	if !ok {
		return nil, fmt.Errorf("diffDays requires a number as first argument")
	}
	ts2, ok := toFloat(args[1])
	if !ok {
		return nil, fmt.Errorf("diffDays requires a number as second argument")
	}
	t1 := time.UnixMilli(int64(ts1))
	t2 := time.UnixMilli(int64(ts2))
	diff := t2.Sub(t1)
	return float64(int(diff.Hours() / 24)), nil
}
