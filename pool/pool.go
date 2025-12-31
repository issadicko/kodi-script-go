// Package pool provides object pooling for frequently allocated objects.
package pool

import (
	"sync"
)

// StringSlicePool pools string slices for output capture.
var StringSlicePool = sync.Pool{
	New: func() interface{} {
		s := make([]string, 0, 16)
		return &s
	},
}

// GetStringSlice gets a string slice from the pool.
func GetStringSlice() *[]string {
	return StringSlicePool.Get().(*[]string)
}

// PutStringSlice returns a string slice to the pool.
func PutStringSlice(s *[]string) {
	*s = (*s)[:0] // Reset length but keep capacity
	StringSlicePool.Put(s)
}

// MapPool pools map[string]interface{} for environments.
var MapPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]interface{}, 16)
	},
}

// GetMap gets a map from the pool.
func GetMap() map[string]interface{} {
	return MapPool.Get().(map[string]interface{})
}

// PutMap returns a map to the pool after clearing it.
func PutMap(m map[string]interface{}) {
	for k := range m {
		delete(m, k)
	}
	MapPool.Put(m)
}

// InterfaceSlicePool pools []interface{} for function arguments.
var InterfaceSlicePool = sync.Pool{
	New: func() interface{} {
		s := make([]interface{}, 0, 8)
		return &s
	},
}

// GetInterfaceSlice gets an interface slice from the pool.
func GetInterfaceSlice() *[]interface{} {
	return InterfaceSlicePool.Get().(*[]interface{})
}

// PutInterfaceSlice returns an interface slice to the pool.
func PutInterfaceSlice(s *[]interface{}) {
	*s = (*s)[:0]
	InterfaceSlicePool.Put(s)
}
