package kodi

import (
	"testing"
)

// ============ Lexer Benchmarks ============

func BenchmarkLexer_SimpleExpression(b *testing.B) {
	code := `1 + 2 * 3`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkLexer_ComplexExpression(b *testing.B) {
	code := `(1 + 2) * (3 - 4) / 5 % 6 + 7 * 8 - 9`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

// ============ Parser Benchmarks ============

func BenchmarkParser_VariableDeclaration(b *testing.B) {
	code := `let x = 42`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkParser_MultipleStatements(b *testing.B) {
	code := `
		let a = 1
		let b = 2
		let c = 3
		let d = a + b + c
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkParser_FunctionDeclaration(b *testing.B) {
	code := `
		let add = fn(a, b) {
			return a + b
		}
		add(1, 2)
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

// ============ Interpreter Benchmarks ============

func BenchmarkInterpreter_Arithmetic(b *testing.B) {
	code := `1 + 2 * 3 - 4 / 2`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkInterpreter_StringConcat(b *testing.B) {
	code := `"hello" + " " + "world"`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkInterpreter_ArrayOperations(b *testing.B) {
	code := `
		let arr = [1, 2, 3, 4, 5]
		let first = arr[0]
		let last = arr[4]
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkInterpreter_ObjectOperations(b *testing.B) {
	code := `
		let obj = { name: "test", value: 42 }
		let n = obj.name
		let v = obj.value
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkInterpreter_FunctionCall(b *testing.B) {
	code := `
		let multiply = fn(a, b) { return a * b }
		multiply(6, 7)
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkInterpreter_Recursion(b *testing.B) {
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

func BenchmarkInterpreter_Loop(b *testing.B) {
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

func BenchmarkInterpreter_Conditionals(b *testing.B) {
	code := `
		let x = 42
		if (x > 10) {
			if (x > 20) {
				if (x > 30) {
					"large"
				} else {
					"medium-large"
				}
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

// ============ Higher-Order Functions Benchmarks ============

func BenchmarkHigherOrder_Map(b *testing.B) {
	code := `
		let arr = [1, 2, 3, 4, 5]
		map(arr, fn(x) { return x * 2 })
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkHigherOrder_Filter(b *testing.B) {
	code := `
		let arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
		filter(arr, fn(x) { return x > 5 })
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkHigherOrder_Reduce(b *testing.B) {
	code := `
		let arr = [1, 2, 3, 4, 5]
		reduce(arr, fn(acc, x) { return acc + x }, 0)
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

// ============ Native Functions Benchmarks ============

func BenchmarkNatives_StringFunctions(b *testing.B) {
	code := `
		let s = "Hello World"
		let upper = toUpperCase(s)
		let lower = toLowerCase(s)
		let len = length(s)
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkNatives_MathFunctions(b *testing.B) {
	code := `
		let a = abs(-42)
		let b = floor(3.7)
		let c = ceil(3.2)
		let d = round(3.5)
		let e = sqrt(16)
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkNatives_ArrayFunctions(b *testing.B) {
	code := `
		let arr = [5, 2, 8, 1, 9, 3]
		let s = size(arr)
		let f = first(arr)
		let l = last(arr)
		let sorted = sort(arr)
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkNatives_HashFunctions(b *testing.B) {
	code := `
		let h1 = md5("hello")
		let h2 = sha1("hello")
		let h3 = sha256("hello")
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

// ============ String Template Benchmarks ============

func BenchmarkStringTemplate_Simple(b *testing.B) {
	code := `
		let name = "World"
		"Hello ${name}!"
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkStringTemplate_Complex(b *testing.B) {
	code := `
		let first = "John"
		let last = "Doe"
		let age = 30
		"Name: ${first} ${last}, Age: ${age}"
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

// ============ Real-World Scenario Benchmarks ============

func BenchmarkScenario_DataTransformation(b *testing.B) {
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

func BenchmarkScenario_ConfigProcessing(b *testing.B) {
	code := `
		let config = {
			host: "localhost",
			port: 8080,
			debug: true,
			timeout: 30
		}
		let url = "http://${config.host}:${toString(config.port)}"
		let isDebug = config.debug == true
		let timeoutMs = config.timeout * 1000
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}

func BenchmarkScenario_BusinessLogic(b *testing.B) {
	code := `
		let calculateDiscount = fn(price, quantity) {
			let discount = 0
			if (quantity >= 100) {
				discount = 0.2
			} else if (quantity >= 50) {
				discount = 0.1
			} else if (quantity >= 10) {
				discount = 0.05
			}
			return price * quantity * (1 - discount)
		}
		
		let total1 = calculateDiscount(10, 5)
		let total2 = calculateDiscount(10, 25)
		let total3 = calculateDiscount(10, 75)
		let total4 = calculateDiscount(10, 150)
	`
	for i := 0; i < b.N; i++ {
		Run(code, nil)
	}
}
