package kodi

import "testing"

// TestSandboxBlocksReflectionIntoEngine guards the sandbox escape where a script
// reached the interpreter's own internals by reflection. A user function is an
// *interpreter.Function whose Env field is exported, and *Environment exposes
// Get/Set/Release — so `f.Env.Release()` returned the running environment to the
// process-global pool, letting one execution inject code and leak variables into
// later, unrelated executions in the same process. Each probe below must now
// fail instead of succeeding.
func TestSandboxBlocksReflectionIntoEngine(t *testing.T) {
	const def = "let f = fn(x) { x }\n"

	cases := []struct {
		name   string
		source string
	}{
		{"read the environment of a function", def + "f.Env"},
		{"read a binding from that environment", def + `f.Env.Get("x")`},
		{"write a binding into that environment", def + `f.Env.Set("x", 1)`},
		{"release the environment into the shared pool", def + "f.Env.Release()"},
		{"reach the function body AST", def + "f.Body"},
		{"reach the function parameters AST", def + "f.Parameters"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := New(tc.source).Execute()
			if len(result.Errors) == 0 {
				t.Fatalf("sandbox escape: a script reached the engine internals and got %v with no error", result.Value)
			}
		})
	}
}

// TestSandboxPoolNotPoisonedAcrossExecutions replays the cross-execution attack
// through three separate scripts sharing one process (the nenga-connect motif):
// the install attempt must fail, and a later execution must compute from its own
// context only, never a value planted by an earlier one.
func TestSandboxPoolNotPoisonedAcrossExecutions(t *testing.T) {
	// Attacker tries to plant a binding that a later execution would pick up.
	install := New("let f = fn(x) { x }\nf.Env.Set(\"stolen\", \"leaked\")").Execute()
	if len(install.Errors) == 0 {
		t.Fatalf("sandbox escape: the install script mutated an engine environment without error")
	}

	// A later, unrelated execution reads `stolen`; it must be undefined, not the
	// planted value.
	victim := New(`stolen`).Execute()
	if len(victim.Errors) == 0 {
		t.Fatalf("pool poisoned: a planted binding leaked into a later execution as %v", victim.Value)
	}
}

// TestSandboxDoesNotBlockHostObjects pins the boundary: the barrier blocks the
// engine's own sub-package types, not a host object the caller bound — even one
// declared in the module root package, as the test fixtures are.
func TestSandboxDoesNotBlockHostObjects(t *testing.T) {
	user := &User{Name: "Alice", Age: 30}

	result := New(`user.Name + " / " + user.SayHello()`).Bind("user", user).Execute()
	if len(result.Errors) > 0 {
		t.Fatalf("host object access was wrongly blocked: %v", result.Errors)
	}
	if result.Value != "Alice / Hello, I'm Alice" {
		t.Errorf("unexpected value: %v", result.Value)
	}
}
