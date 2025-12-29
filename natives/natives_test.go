package natives

import (
	"strings"
	"testing"
)

func TestStringFunctions(t *testing.T) {
	t.Run("toString", func(t *testing.T) {
		result, err := nativeToString(42)
		if err != nil || result != "42" {
			t.Errorf("expected '42', got %v", result)
		}
		_, err = nativeToString()
		if err == nil {
			t.Error("expected error for no args")
		}
	})

	t.Run("toNumber", func(t *testing.T) {
		result, err := nativeToNumber("42.5")
		if err != nil || result != 42.5 {
			t.Errorf("expected 42.5, got %v", result)
		}
		result, err = nativeToNumber(float64(10))
		if err != nil || result != float64(10) {
			t.Errorf("expected 10, got %v", result)
		}
		result, err = nativeToNumber(int(5))
		if err != nil || result != float64(5) {
			t.Errorf("expected 5, got %v", result)
		}
		_, err = nativeToNumber("invalid")
		if err == nil {
			t.Error("expected error for invalid number")
		}
		_, err = nativeToNumber(true)
		if err == nil {
			t.Error("expected error for bool")
		}
	})

	t.Run("length", func(t *testing.T) {
		result, err := nativeLength("hello")
		if err != nil || result != float64(5) {
			t.Errorf("expected 5, got %v", result)
		}
		_, err = nativeLength(42)
		if err == nil {
			t.Error("expected error for non-string")
		}
	})

	t.Run("substring", func(t *testing.T) {
		result, err := nativeSubstring("hello", float64(1))
		if err != nil || result != "ello" {
			t.Errorf("expected 'ello', got %v", result)
		}
		result, err = nativeSubstring("hello", float64(1), float64(3))
		if err != nil || result != "el" {
			t.Errorf("expected 'el', got %v", result)
		}
		result, err = nativeSubstring("hello", float64(-1))
		if err != nil || result != "hello" {
			t.Errorf("expected 'hello', got %v", result)
		}
		result, err = nativeSubstring("hello", float64(10))
		if err != nil || result != "" {
			t.Errorf("expected '', got %v", result)
		}
		result, err = nativeSubstring("hello", float64(0), float64(100))
		if err != nil || result != "hello" {
			t.Errorf("expected 'hello', got %v", result)
		}
	})

	t.Run("toUpperCase", func(t *testing.T) {
		result, err := nativeToUpperCase("hello")
		if err != nil || result != "HELLO" {
			t.Errorf("expected 'HELLO', got %v", result)
		}
	})

	t.Run("toLowerCase", func(t *testing.T) {
		result, err := nativeToLowerCase("HELLO")
		if err != nil || result != "hello" {
			t.Errorf("expected 'hello', got %v", result)
		}
	})

	t.Run("trim", func(t *testing.T) {
		result, err := nativeTrim("  hello  ")
		if err != nil || result != "hello" {
			t.Errorf("expected 'hello', got %v", result)
		}
	})

	t.Run("replace", func(t *testing.T) {
		result, err := nativeReplace("hello world", "world", "kodi")
		if err != nil || result != "hello kodi" {
			t.Errorf("expected 'hello kodi', got %v", result)
		}
	})

	t.Run("contains", func(t *testing.T) {
		result, err := nativeContains("hello world", "world")
		if err != nil || result != true {
			t.Errorf("expected true, got %v", result)
		}
		result, err = nativeContains("hello world", "foo")
		if err != nil || result != false {
			t.Errorf("expected false, got %v", result)
		}
	})

	t.Run("startsWith", func(t *testing.T) {
		result, err := nativeStartsWith("hello world", "hello")
		if err != nil || result != true {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("endsWith", func(t *testing.T) {
		result, err := nativeEndsWith("hello world", "world")
		if err != nil || result != true {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("indexOf", func(t *testing.T) {
		result, err := nativeIndexOf("hello world", "world")
		if err != nil || result != float64(6) {
			t.Errorf("expected 6, got %v", result)
		}
	})

	t.Run("split", func(t *testing.T) {
		result, err := nativeSplit("a,b,c", ",")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		arr := result.([]interface{})
		if len(arr) != 3 || arr[0] != "a" {
			t.Errorf("expected ['a','b','c'], got %v", arr)
		}
	})

	t.Run("join", func(t *testing.T) {
		arr := []interface{}{"a", "b", "c"}
		result, err := nativeJoin(arr, ",")
		if err != nil || result != "a,b,c" {
			t.Errorf("expected 'a,b,c', got %v", result)
		}
	})
}

func TestMathFunctions(t *testing.T) {
	t.Run("abs", func(t *testing.T) {
		result, err := nativeAbs(float64(-5))
		if err != nil || result != float64(5) {
			t.Errorf("expected 5, got %v", result)
		}
	})

	t.Run("floor", func(t *testing.T) {
		result, err := nativeFloor(float64(3.7))
		if err != nil || result != float64(3) {
			t.Errorf("expected 3, got %v", result)
		}
	})

	t.Run("ceil", func(t *testing.T) {
		result, err := nativeCeil(float64(3.2))
		if err != nil || result != float64(4) {
			t.Errorf("expected 4, got %v", result)
		}
	})

	t.Run("round", func(t *testing.T) {
		result, err := nativeRound(float64(3.5))
		if err != nil || result != float64(4) {
			t.Errorf("expected 4, got %v", result)
		}
	})

	t.Run("min", func(t *testing.T) {
		result, err := nativeMin(float64(5), float64(3), float64(8))
		if err != nil || result != float64(3) {
			t.Errorf("expected 3, got %v", result)
		}
	})

	t.Run("max", func(t *testing.T) {
		result, err := nativeMax(float64(5), float64(3), float64(8))
		if err != nil || result != float64(8) {
			t.Errorf("expected 8, got %v", result)
		}
	})

	t.Run("pow", func(t *testing.T) {
		result, err := nativePow(float64(2), float64(3))
		if err != nil || result != float64(8) {
			t.Errorf("expected 8, got %v", result)
		}
	})

	t.Run("sqrt", func(t *testing.T) {
		result, err := nativeSqrt(float64(16))
		if err != nil || result != float64(4) {
			t.Errorf("expected 4, got %v", result)
		}
		_, err = nativeSqrt(float64(-1))
		if err == nil {
			t.Error("expected error for negative sqrt")
		}
	})

	t.Run("sin", func(t *testing.T) {
		_, err := nativeSin(float64(0))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("cos", func(t *testing.T) {
		_, err := nativeCos(float64(0))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("tan", func(t *testing.T) {
		_, err := nativeTan(float64(0))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("log", func(t *testing.T) {
		_, err := nativeLog(float64(10))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		_, err = nativeLog(float64(-1))
		if err == nil {
			t.Error("expected error for negative log")
		}
	})

	t.Run("log10", func(t *testing.T) {
		_, err := nativeLog10(float64(10))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		_, err = nativeLog10(float64(-1))
		if err == nil {
			t.Error("expected error for negative log10")
		}
	})

	t.Run("exp", func(t *testing.T) {
		_, err := nativeExp(float64(1))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRandomFunctions(t *testing.T) {
	t.Run("random", func(t *testing.T) {
		result, err := nativeRandom()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		v := result.(float64)
		if v < 0 || v >= 1 {
			t.Errorf("expected [0,1), got %v", v)
		}
	})

	t.Run("randomInt", func(t *testing.T) {
		result, err := nativeRandomInt(float64(1), float64(10))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		v := result.(float64)
		if v < 1 || v > 10 {
			t.Errorf("expected [1,10], got %v", v)
		}
		_, err = nativeRandomInt(float64(10), float64(1))
		if err == nil {
			t.Error("expected error for min >= max")
		}
	})

	t.Run("randomUUID", func(t *testing.T) {
		result, err := nativeRandomUUID()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		uuid := result.(string)
		if len(uuid) != 36 {
			t.Errorf("expected 36 chars, got %d", len(uuid))
		}
	})
}

func TestCryptoFunctions(t *testing.T) {
	t.Run("md5", func(t *testing.T) {
		result, err := nativeMd5("hello")
		if err != nil || result != "5d41402abc4b2a76b9719d911017c592" {
			t.Errorf("unexpected result: %v", result)
		}
	})

	t.Run("sha1", func(t *testing.T) {
		result, err := nativeSha1("hello")
		if err != nil || result != "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d" {
			t.Errorf("unexpected result: %v", result)
		}
	})

	t.Run("sha256", func(t *testing.T) {
		result, err := nativeSha256("hello")
		if err != nil || result != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
			t.Errorf("unexpected result: %v", result)
		}
	})
}

func TestJSONFunctions(t *testing.T) {
	t.Run("jsonParse", func(t *testing.T) {
		result, err := nativeJsonParse(`{"name":"test"}`)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		m := result.(map[string]interface{})
		if m["name"] != "test" {
			t.Errorf("expected 'test', got %v", m["name"])
		}
		_, err = nativeJsonParse("invalid")
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("jsonStringify", func(t *testing.T) {
		result, err := nativeJsonStringify(map[string]interface{}{"name": "test"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !strings.Contains(result.(string), "test") {
			t.Errorf("expected JSON string, got %v", result)
		}
	})
}

func TestBase64Functions(t *testing.T) {
	t.Run("base64Encode", func(t *testing.T) {
		result, err := nativeBase64Encode("hello")
		if err != nil || result != "aGVsbG8=" {
			t.Errorf("expected 'aGVsbG8=', got %v", result)
		}
	})

	t.Run("base64Decode", func(t *testing.T) {
		result, err := nativeBase64Decode("aGVsbG8=")
		if err != nil || result != "hello" {
			t.Errorf("expected 'hello', got %v", result)
		}
		_, err = nativeBase64Decode("!!invalid!!")
		if err == nil {
			t.Error("expected error for invalid base64")
		}
	})
}

func TestURLFunctions(t *testing.T) {
	t.Run("urlEncode", func(t *testing.T) {
		result, err := nativeUrlEncode("hello world")
		if err != nil || result != "hello+world" {
			t.Errorf("expected 'hello+world', got %v", result)
		}
	})

	t.Run("urlDecode", func(t *testing.T) {
		result, err := nativeUrlDecode("hello+world")
		if err != nil || result != "hello world" {
			t.Errorf("expected 'hello world', got %v", result)
		}
		_, err = nativeUrlDecode("%zz")
		if err == nil {
			t.Error("expected error for invalid URL encoding")
		}
	})
}

func TestTypeFunctions(t *testing.T) {
	t.Run("typeOf", func(t *testing.T) {
		result, _ := nativeTypeOf("hello")
		if result != "string" {
			t.Errorf("expected 'string', got %v", result)
		}
		result, _ = nativeTypeOf(float64(42))
		if result != "number" {
			t.Errorf("expected 'number', got %v", result)
		}
		result, _ = nativeTypeOf(true)
		if result != "boolean" {
			t.Errorf("expected 'boolean', got %v", result)
		}
		result, _ = nativeTypeOf(nil)
		if result != "null" {
			t.Errorf("expected 'null', got %v", result)
		}
		result, _ = nativeTypeOf(map[string]interface{}{})
		if result != "object" {
			t.Errorf("expected 'object', got %v", result)
		}
		result, _ = nativeTypeOf([]interface{}{})
		if result != "array" {
			t.Errorf("expected 'array', got %v", result)
		}
	})

	t.Run("isNull", func(t *testing.T) {
		result, _ := nativeIsNull(nil)
		if result != true {
			t.Errorf("expected true, got %v", result)
		}
		result, _ = nativeIsNull("not null")
		if result != false {
			t.Errorf("expected false, got %v", result)
		}
	})

	t.Run("isNumber", func(t *testing.T) {
		result, _ := nativeIsNumber(float64(42))
		if result != true {
			t.Errorf("expected true, got %v", result)
		}
		result, _ = nativeIsNumber("42")
		if result != false {
			t.Errorf("expected false, got %v", result)
		}
	})

	t.Run("isString", func(t *testing.T) {
		result, _ := nativeIsString("hello")
		if result != true {
			t.Errorf("expected true, got %v", result)
		}
	})

	t.Run("isBool", func(t *testing.T) {
		result, _ := nativeIsBool(true)
		if result != true {
			t.Errorf("expected true, got %v", result)
		}
	})
}

func TestArrayFunctions(t *testing.T) {
	t.Run("sort", func(t *testing.T) {
		arr := []interface{}{float64(3), float64(1), float64(2)}
		result, err := nativeSort(arr)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		sorted := result.([]interface{})
		if sorted[0] != float64(1) || sorted[2] != float64(3) {
			t.Errorf("sort failed: %v", sorted)
		}

		result, err = nativeSort(arr, "desc")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		sorted = result.([]interface{})
		if sorted[0] != float64(3) || sorted[2] != float64(1) {
			t.Errorf("sort desc failed: %v", sorted)
		}
	})

	t.Run("sortBy", func(t *testing.T) {
		arr := []interface{}{
			map[string]interface{}{"age": float64(30)},
			map[string]interface{}{"age": float64(20)},
		}
		result, err := nativeSortBy(arr, "age")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		sorted := result.([]interface{})
		first := sorted[0].(map[string]interface{})
		if first["age"] != float64(20) {
			t.Errorf("sortBy failed: %v", sorted)
		}
	})

	t.Run("reverse", func(t *testing.T) {
		arr := []interface{}{float64(1), float64(2), float64(3)}
		result, err := nativeReverse(arr)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		reversed := result.([]interface{})
		if reversed[0] != float64(3) {
			t.Errorf("reverse failed: %v", reversed)
		}
	})

	t.Run("size", func(t *testing.T) {
		result, err := nativeSize([]interface{}{1, 2, 3})
		if err != nil || result != float64(3) {
			t.Errorf("expected 3, got %v", result)
		}
		result, err = nativeSize("hello")
		if err != nil || result != float64(5) {
			t.Errorf("expected 5, got %v", result)
		}
		result, err = nativeSize(map[string]interface{}{"a": 1, "b": 2})
		if err != nil || result != float64(2) {
			t.Errorf("expected 2, got %v", result)
		}
	})

	t.Run("first", func(t *testing.T) {
		result, err := nativeFirst([]interface{}{float64(1), float64(2)})
		if err != nil || result != float64(1) {
			t.Errorf("expected 1, got %v", result)
		}
		result, err = nativeFirst([]interface{}{})
		if err != nil || result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})

	t.Run("last", func(t *testing.T) {
		result, err := nativeLast([]interface{}{float64(1), float64(2)})
		if err != nil || result != float64(2) {
			t.Errorf("expected 2, got %v", result)
		}
	})

	t.Run("slice", func(t *testing.T) {
		arr := []interface{}{float64(1), float64(2), float64(3), float64(4)}
		result, err := nativeSlice(arr, float64(1), float64(3))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		sliced := result.([]interface{})
		if len(sliced) != 2 || sliced[0] != float64(2) {
			t.Errorf("slice failed: %v", sliced)
		}

		result, err = nativeSlice(arr, float64(2))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		sliced = result.([]interface{})
		if len(sliced) != 2 {
			t.Errorf("slice without end failed: %v", sliced)
		}

		// Edge cases
		result, err = nativeSlice(arr, float64(-1))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		result, err = nativeSlice(arr, float64(100))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		result, err = nativeSlice(arr, float64(3), float64(1))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	fn := r.Get("toString")
	if fn == nil {
		t.Error("expected function, got nil")
	}

	r.Register("custom", func(args ...interface{}) (interface{}, error) {
		return "custom", nil
	})
	fn = r.Get("custom")
	if fn == nil {
		t.Error("expected custom function")
	}
	result, _ := fn()
	if result != "custom" {
		t.Errorf("expected 'custom', got %v", result)
	}
}

func TestCompareValues(t *testing.T) {
	// nil comparisons
	if compareValues(nil, nil) != 0 {
		t.Error("nil == nil should be 0")
	}
	if compareValues(nil, "a") >= 0 {
		t.Error("nil < anything")
	}
	if compareValues("a", nil) <= 0 {
		t.Error("anything > nil")
	}

	// Number comparisons
	if compareValues(float64(1), float64(2)) >= 0 {
		t.Error("1 < 2")
	}
	if compareValues(float64(2), float64(1)) <= 0 {
		t.Error("2 > 1")
	}
	if compareValues(float64(1), float64(1)) != 0 {
		t.Error("1 == 1")
	}

	// String comparisons
	if compareValues("a", "b") >= 0 {
		t.Error("a < b")
	}
}

func TestGetFieldValue(t *testing.T) {
	obj := map[string]interface{}{"name": "test"}
	if getFieldValue(obj, "name") != "test" {
		t.Error("expected 'test'")
	}
	if getFieldValue(obj, "missing") != nil {
		t.Error("expected nil for missing field")
	}
	if getFieldValue("not an object", "field") != nil {
		t.Error("expected nil for non-object")
	}
}
