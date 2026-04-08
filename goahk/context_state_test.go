package goahk

import (
	"context"
	"testing"

	"goahk/internal/actions"
	internalapp "goahk/internal/app"
)

func TestContextState_PerTriggerVarsIsolationAndSharedAppState(t *testing.T) {
	t.Parallel()

	a := NewApp()
	a.Bind("Ctrl+H", Func(func(ctx *Context) error {
		seed := ctx.Vars["seed"]
		ctx.AppState.Set("seen_seed_"+seed, seed)
		ctx.Vars["local"] = "first"
		ctx.AppState.Set("shared", "value")
		return nil
	}))
	program, _, callbacks := a.runtimeArtifacts()
	registry := buildRegistryWithCallbacks(a.state, callbacks)
	executor := actions.NewExecutor(registry)

	bindings, err := internalapp.CompileRuntimeBindingsFromProgram(program, registry)
	if err != nil {
		t.Fatalf("compile runtime bindings: %v", err)
	}

	res := executor.Execute(actions.ActionContext{
		Context: context.Background(),
		Metadata: map[string]string{
			"seed": "1",
		},
		BindingID: bindings[0].ID,
	}, bindings[0].Plan)
	if !res.Success {
		t.Fatalf("first execution failed: %#v", res)
	}

	res2 := executor.Execute(actions.ActionContext{
		Context: context.Background(),
		Metadata: map[string]string{
			"seed": "2",
		},
		BindingID: bindings[0].ID,
	}, bindings[0].Plan)
	if !res2.Success {
		t.Fatalf("second execution failed: %#v", res2)
	}
	if got, ok := a.state.Get("seen_seed_1"); !ok || got != "1" {
		t.Fatalf("seed 1 marker = (%q, %v), want (1, true)", got, ok)
	}
	if got, ok := a.state.Get("seen_seed_2"); !ok || got != "2" {
		t.Fatalf("seed 2 marker = (%q, %v), want (2, true)", got, ok)
	}
	if got, ok := a.state.Get("shared"); !ok || got != "value" {
		t.Fatalf("shared state = (%q, %v), want (value, true)", got, ok)
	}
}
