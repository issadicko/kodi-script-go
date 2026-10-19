package natives

import (
	"fmt"
	"strconv"
	"strings"
)

// ============ List utilities ============
//
// A list argument may be null, which behaves as an empty list; any other
// non-list value is an error. Two values are equal when they have the same
// type and the same canonical form, so 1 equals 1.0 but not "1".

// ListArg reads a list argument for the function name; null is an empty list.
func ListArg(name string, v interface{}) ([]interface{}, error) {
	switch l := v.(type) {
	case nil:
		return []interface{}{}, nil
	case []interface{}:
		return l, nil
	}
	return nil, fmt.Errorf("%s requires an array as first argument", name)
}

// EqualityKey identifies a value for equality: its type and its canonical form.
func EqualityKey(v interface{}) string {
	kind, _ := nativeTypeOf(v)
	return kind.(string) + ":" + Stringify(v)
}

// CompareValues orders two values: numbers numerically, anything else by canonical form.
func CompareValues(a, b interface{}) int {
	return compareValues(a, b)
}

// AtPath reads a dotted path: object keys, or list indexes written in digits.
// The empty path is the value itself; a missing step gives nil.
func AtPath(v interface{}, path string) interface{} {
	if path == "" {
		return v
	}
	current := v
	for _, segment := range strings.Split(path, ".") {
		switch c := current.(type) {
		case map[string]interface{}:
			current = c[segment]
		case []interface{}:
			idx, err := strconv.Atoi(segment)
			if err != nil || idx < 0 || idx >= len(c) || segment != strconv.Itoa(idx) {
				return nil
			}
			current = c[idx]
		default:
			return nil
		}
	}
	return current
}

func nativeChunk(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("chunk requires 2 arguments (array, size)")
	}
	list, err := ListArg("chunk", args[0])
	if err != nil {
		return nil, err
	}
	n, ok := toFloat(args[1])
	if !ok || int(n) < 1 {
		return nil, fmt.Errorf("chunk requires a size of at least 1")
	}
	size := int(n)
	result := []interface{}{}
	for start := 0; start < len(list); start += size {
		end := start + size
		if end > len(list) {
			end = len(list)
		}
		part := make([]interface{}, end-start)
		copy(part, list[start:end])
		result = append(result, part)
	}
	return result, nil
}

func nativeCompact(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("compact requires 1 argument (array)")
	}
	list, err := ListArg("compact", args[0])
	if err != nil {
		return nil, err
	}
	result := []interface{}{}
	for _, item := range list {
		if item != nil {
			result = append(result, item)
		}
	}
	return result, nil
}

func nativePluck(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("pluck requires 2 arguments (array, path)")
	}
	list, err := ListArg("pluck", args[0])
	if err != nil {
		return nil, err
	}
	path, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("pluck requires a string path as second argument")
	}
	result := make([]interface{}, len(list))
	for i, item := range list {
		result[i] = AtPath(item, path)
	}
	return result, nil
}

func nativeZip(args ...interface{}) (interface{}, error) {
	lists := make([][]interface{}, len(args))
	shortest := -1
	for i, a := range args {
		list, err := ListArg("zip", a)
		if err != nil {
			return nil, err
		}
		lists[i] = list
		if shortest < 0 || len(list) < shortest {
			shortest = len(list)
		}
	}
	result := []interface{}{}
	for row := 0; row < shortest; row++ {
		tuple := make([]interface{}, len(lists))
		for i, list := range lists {
			tuple[i] = list[row]
		}
		result = append(result, tuple)
	}
	return result, nil
}

func nativeFromEntries(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("fromEntries requires 1 argument (array of pairs)")
	}
	list, err := ListArg("fromEntries", args[0])
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	for _, item := range list {
		pair, ok := item.([]interface{})
		if !ok || len(pair) < 2 {
			return nil, fmt.Errorf("fromEntries requires each element to be a [key, value] pair")
		}
		result[Stringify(pair[0])] = pair[1]
	}
	return result, nil
}

func keySet(list []interface{}) map[string]bool {
	set := make(map[string]bool, len(list))
	for _, item := range list {
		set[EqualityKey(item)] = true
	}
	return set
}

func twoLists(name string, args []interface{}) ([]interface{}, []interface{}, error) {
	if len(args) != 2 {
		return nil, nil, fmt.Errorf("%s requires 2 arguments (array, array)", name)
	}
	a, err := ListArg(name, args[0])
	if err != nil {
		return nil, nil, err
	}
	b, err := ListArg(name, args[1])
	if err != nil {
		return nil, nil, fmt.Errorf("%s requires an array as second argument", name)
	}
	return a, b, nil
}

func nativeDifference(args ...interface{}) (interface{}, error) {
	a, b, err := twoLists("difference", args)
	if err != nil {
		return nil, err
	}
	excluded := keySet(b)
	result := []interface{}{}
	for _, item := range a {
		if !excluded[EqualityKey(item)] {
			result = append(result, item)
		}
	}
	return result, nil
}

func nativeIntersection(args ...interface{}) (interface{}, error) {
	a, b, err := twoLists("intersection", args)
	if err != nil {
		return nil, err
	}
	kept := keySet(b)
	seen := map[string]bool{}
	result := []interface{}{}
	for _, item := range a {
		k := EqualityKey(item)
		if kept[k] && !seen[k] {
			seen[k] = true
			result = append(result, item)
		}
	}
	return result, nil
}

func nativeUnion(args ...interface{}) (interface{}, error) {
	seen := map[string]bool{}
	result := []interface{}{}
	for i, a := range args {
		list, err := ListArg("union", a)
		if err != nil && i > 0 {
			return nil, fmt.Errorf("union requires array arguments")
		} else if err != nil {
			return nil, err
		}
		for _, item := range list {
			if k := EqualityKey(item); !seen[k] {
				seen[k] = true
				result = append(result, item)
			}
		}
	}
	return result, nil
}
