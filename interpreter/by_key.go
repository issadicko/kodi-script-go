package interpreter

import (
	"fmt"

	"github.com/issadicko/kodi-script-go/natives"
)

// Builtins that group or aggregate a list by a key. The key is a dotted path
// read on each item, or a function called as fn(item, index); they live with
// map and filter because the function form calls back into the script.

// keyed is a list with the key of each item, computed once.
type keyed struct {
	items []interface{}
	keys  []Value
}

func (i *Interpreter) keyedList(name string, args []Value) (keyed, error) {
	if len(args) != 2 {
		return keyed{}, fmt.Errorf("%s requires 2 arguments: array and key (path or function)", name)
	}
	list, err := natives.ListArg(name, args[0])
	if err != nil {
		return keyed{}, err
	}
	keyOf, err := i.keyReader(name, args[1])
	if err != nil {
		return keyed{}, err
	}
	keys := make([]Value, len(list))
	for idx, item := range list {
		if keys[idx], err = keyOf(item, idx); err != nil {
			return keyed{}, err
		}
	}
	return keyed{items: list, keys: keys}, nil
}

func (i *Interpreter) keyReader(name string, key Value) (func(Value, int) (Value, error), error) {
	if path, ok := key.(string); ok {
		return func(item Value, _ int) (Value, error) { return natives.AtPath(item, path), nil }, nil
	}
	if isCallable(key) {
		return func(item Value, idx int) (Value, error) {
			return i.applyFunction(key, []Value{item, float64(idx)})
		}, nil
	}
	return nil, fmt.Errorf("%s requires a path string or a function as key", name)
}

func (i *Interpreter) builtinGroupBy(args []Value) (Value, error) {
	k, err := i.keyedList("groupBy", args)
	if err != nil {
		return nil, err
	}
	groups := map[string]interface{}{}
	for idx, item := range k.items {
		g := natives.Stringify(k.keys[idx])
		existing, _ := groups[g].([]interface{})
		groups[g] = append(existing, item)
	}
	return groups, nil
}

func (i *Interpreter) builtinCountBy(args []Value) (Value, error) {
	k, err := i.keyedList("countBy", args)
	if err != nil {
		return nil, err
	}
	counts := map[string]interface{}{}
	for idx := range k.items {
		g := natives.Stringify(k.keys[idx])
		n, _ := counts[g].(float64)
		counts[g] = n + 1
	}
	return counts, nil
}

func (i *Interpreter) builtinIndexBy(args []Value) (Value, error) {
	k, err := i.keyedList("indexBy", args)
	if err != nil {
		return nil, err
	}
	index := map[string]interface{}{}
	for idx, item := range k.items {
		index[natives.Stringify(k.keys[idx])] = item
	}
	return index, nil
}

func (i *Interpreter) builtinUniqueBy(args []Value) (Value, error) {
	k, err := i.keyedList("uniqueBy", args)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	result := []interface{}{}
	for idx, item := range k.items {
		if e := natives.EqualityKey(k.keys[idx]); !seen[e] {
			seen[e] = true
			result = append(result, item)
		}
	}
	return result, nil
}

func (i *Interpreter) builtinSumBy(args []Value) (Value, error) {
	k, err := i.keyedList("sumBy", args)
	if err != nil {
		return nil, err
	}
	total := 0.0
	for _, key := range k.keys {
		if key == nil {
			continue
		}
		n, ok := toNumber(key)
		if !ok {
			return nil, fmt.Errorf("sumBy requires numeric keys, got %s", natives.Stringify(key))
		}
		total += n
	}
	return total, nil
}

func (i *Interpreter) extremeBy(name string, args []Value, wanted int) (Value, error) {
	k, err := i.keyedList(name, args)
	if err != nil {
		return nil, err
	}
	best := -1
	for idx, key := range k.keys {
		if key == nil {
			continue
		}
		if best < 0 || natives.CompareValues(key, k.keys[best])*wanted > 0 {
			best = idx
		}
	}
	if best < 0 {
		return nil, nil
	}
	return k.items[best], nil
}

func (i *Interpreter) builtinMinBy(args []Value) (Value, error) {
	return i.extremeBy("minBy", args, -1)
}

func (i *Interpreter) builtinMaxBy(args []Value) (Value, error) {
	return i.extremeBy("maxBy", args, 1)
}

func (i *Interpreter) builtinPartition(args []Value) (Value, error) {
	k, err := i.keyedList("partition", args)
	if err != nil {
		return nil, err
	}
	yes, no := []interface{}{}, []interface{}{}
	for idx, item := range k.items {
		if isTruthy(k.keys[idx]) {
			yes = append(yes, item)
		} else {
			no = append(no, item)
		}
	}
	return []interface{}{yes, no}, nil
}
