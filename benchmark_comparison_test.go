package kodi

import (
	"testing"

	"github.com/dop251/goja"
)

// ============ Arithmetic Comparison ============

func BenchmarkComparison_Arithmetic_KodiScript(b *testing.B) {
	code := `1 + 2 * 3 - 4 / 2`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_Arithmetic_Goja(b *testing.B) {
	code := `1 + 2 * 3 - 4 / 2`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ Variable Declaration Comparison ============

func BenchmarkComparison_Variables_KodiScript(b *testing.B) {
	code := `
		let a = 1
		let b = 2
		let c = a + b
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_Variables_Goja(b *testing.B) {
	code := `
		let a = 1
		let b = 2
		let c = a + b
	`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ Function Call Comparison ============

func BenchmarkComparison_Function_KodiScript(b *testing.B) {
	code := `
		let add = fn(a, b) { return a + b }
		add(10, 20)
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_Function_Goja(b *testing.B) {
	code := `
		function add(a, b) { return a + b }
		add(10, 20)
	`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ Recursion Comparison ============

func BenchmarkComparison_Recursion_KodiScript(b *testing.B) {
	code := `
		let factorial = fn(n) {
			if (n <= 1) { return 1 }
			return n * factorial(n - 1)
		}
		factorial(10)
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_Recursion_Goja(b *testing.B) {
	code := `
		function factorial(n) {
			if (n <= 1) { return 1 }
			return n * factorial(n - 1)
		}
		factorial(10)
	`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ Loop Comparison ============

func BenchmarkComparison_Loop_KodiScript(b *testing.B) {
	code := `
		let arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
		let sum = 0
		for (i in arr) {
			sum = sum + i
		}
		sum
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_Loop_Goja(b *testing.B) {
	code := `
		let arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
		let sum = 0
		for (let i of arr) {
			sum = sum + i
		}
		sum
	`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ Array Map Comparison ============

func BenchmarkComparison_ArrayMap_KodiScript(b *testing.B) {
	code := `
		let arr = [1, 2, 3, 4, 5]
		map(arr, fn(x) { return x * 2 })
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_ArrayMap_Goja(b *testing.B) {
	code := `
		let arr = [1, 2, 3, 4, 5]
		arr.map(x => x * 2)
	`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ Object Access Comparison ============

func BenchmarkComparison_Object_KodiScript(b *testing.B) {
	code := `
		let obj = { name: "test", value: 42, active: true }
		let n = obj.name
		let v = obj.value
		let a = obj.active
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_Object_Goja(b *testing.B) {
	code := `
		let obj = { name: "test", value: 42, active: true }
		let n = obj.name
		let v = obj.value
		let a = obj.active
	`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ String Operations Comparison ============

func BenchmarkComparison_String_KodiScript(b *testing.B) {
	code := `"hello" + " " + "world"`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_String_Goja(b *testing.B) {
	code := `"hello" + " " + "world"`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ Conditionals Comparison ============

func BenchmarkComparison_Conditionals_KodiScript(b *testing.B) {
	code := `
		let x = 42
		if (x > 10) {
			if (x > 20) {
				"large"
			} else {
				"medium"
			}
		} else {
			"small"
		}
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_Conditionals_Goja(b *testing.B) {
	code := `
		let x = 42
		if (x > 10) {
			if (x > 20) {
				"large"
			} else {
				"medium"
			}
		} else {
			"small"
		}
	`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ Complex Scenario Comparison ============

func BenchmarkComparison_ComplexScenario_KodiScript(b *testing.B) {
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
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkComparison_ComplexScenario_Goja(b *testing.B) {
	code := `
		let users = [
			{ name: "Alice", age: 25 },
			{ name: "Bob", age: 30 },
			{ name: "Charlie", age: 35 }
		]
		let names = users.map(u => u.name)
		let adults = users.filter(u => u.age >= 30)
		let totalAge = users.reduce((acc, u) => acc + u.age, 0)
	`
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(code)
	}
}

// ============ With VM Reuse (more realistic for Goja) ============

func BenchmarkComparison_Arithmetic_Goja_Reuse(b *testing.B) {
	code := `1 + 2 * 3 - 4 / 2`
	vm := goja.New()
	program, _ := goja.Compile("", code, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.RunProgram(program)
	}
}

func BenchmarkComparison_Recursion_Goja_Reuse(b *testing.B) {
	code := `
		function factorial(n) {
			if (n <= 1) { return 1 }
			return n * factorial(n - 1)
		}
		factorial(10)
	`
	vm := goja.New()
	program, _ := goja.Compile("", code, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.RunProgram(program)
	}
}

func BenchmarkComparison_ComplexScenario_Goja_Reuse(b *testing.B) {
	code := `
		let users = [
			{ name: "Alice", age: 25 },
			{ name: "Bob", age: 30 },
			{ name: "Charlie", age: 35 }
		]
		let names = users.map(u => u.name)
		let adults = users.filter(u => u.age >= 30)
		let totalAge = users.reduce((acc, u) => acc + u.age, 0)
	`
	vm := goja.New()
	program, _ := goja.Compile("", code, false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.RunProgram(program)
	}
}
