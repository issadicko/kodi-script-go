package kodi

import (
	"testing"

	"github.com/issadicko/kodi-script-go/cache"
)

// BenchmarkWithCache tests performance with AST caching enabled
func BenchmarkWithCache_Arithmetic(b *testing.B) {
	code := `1 + 2 * 3 - 4 / 2`
	cache.DefaultCache.Clear()
	
	// Warm up cache
	Run(code, nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(code).WithCache(true).Execute()
	}
}

// BenchmarkWithoutCache tests performance without caching
func BenchmarkWithoutCache_Arithmetic(b *testing.B) {
	code := `1 + 2 * 3 - 4 / 2`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(code).WithCache(false).Execute()
	}
}

func BenchmarkWithCache_Function(b *testing.B) {
	code := `
		let add = fn(a, b) { return a + b }
		add(10, 20)
	`
	cache.DefaultCache.Clear()
	Run(code, nil) // Warm up
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(code).WithCache(true).Execute()
	}
}

func BenchmarkWithoutCache_Function(b *testing.B) {
	code := `
		let add = fn(a, b) { return a + b }
		add(10, 20)
	`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(code).WithCache(false).Execute()
	}
}

func BenchmarkWithCache_Recursion(b *testing.B) {
	code := `
		let factorial = fn(n) {
			if (n <= 1) { return 1 }
			return n * factorial(n - 1)
		}
		factorial(10)
	`
	cache.DefaultCache.Clear()
	Run(code, nil) // Warm up
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(code).WithCache(true).Execute()
	}
}

func BenchmarkWithoutCache_Recursion(b *testing.B) {
	code := `
		let factorial = fn(n) {
			if (n <= 1) { return 1 }
			return n * factorial(n - 1)
		}
		factorial(10)
	`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(code).WithCache(false).Execute()
	}
}

func BenchmarkWithCache_ComplexScenario(b *testing.B) {
	code := `
		let users = [
			{ name: "Alice", age: 25 },
			{ name: "Bob", age: 30 },
			{ name: "Charlie", age: 35 }
		]
		let names = map(users, fn(u) { return u.name })
		let adults = filter(users, fn(u) { return u.age >= 30 })
		let totalAge = reduce(users, fn(acc, u) { return acc + u.age }, 0)
	`
	cache.DefaultCache.Clear()
	Run(code, nil) // Warm up
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(code).WithCache(true).Execute()
	}
}

func BenchmarkWithoutCache_ComplexScenario(b *testing.B) {
	code := `
		let users = [
			{ name: "Alice", age: 25 },
			{ name: "Bob", age: 30 },
			{ name: "Charlie", age: 35 }
		]
		let names = map(users, fn(u) { return u.name })
		let adults = filter(users, fn(u) { return u.age >= 30 })
		let totalAge = reduce(users, fn(acc, u) { return acc + u.age }, 0)
	`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(code).WithCache(false).Execute()
	}
}
