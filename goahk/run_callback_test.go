package goahk

import (
	"context"
	"testing"

	"goahk/internal/actions"
	internalapp "goahk/internal/app"
)

func TestRunCallback_RegistersAndExecutesThroughExecutor(t *testing.T) {
	t.Parallel()

	called := false
	a := NewApp()
	a.Bind("Ctrl+H", Func(func(*Context) error {
		called = true
		return nil
	}))

	bindings, executor := compileTestBindings(t, a)
	res := executor.Execute(actions.ActionContext{Context: context.Background(), BindingID: bindings[0].ID}, bindings[0].Plan)
	if !res.Success {
		t.Fatalf("execution failed: %#v", res)
	}
	if !called {
		t.Fatal("callback was not executed")
	}
}

func compileTestBindings(t *testing.T, a *App) ([]internalapp.RuntimeBinding, *actions.Executor) {
	t.Helper()

	program, _, callbacks := a.runtimeArtifacts()
	registry := buildRegistryWithCallbacks(a.state, callbacks)
	bindings, err := internalapp.CompileRuntimeBindingsFromProgram(program, registry)
	if err != nil {
		t.Fatalf("CompileRuntimeBindingsFromProgram() error = %v", err)
	}
	return bindings, actions.NewExecutor(registry)
}
