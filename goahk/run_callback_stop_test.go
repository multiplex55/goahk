package goahk

import (
	"context"
	"testing"

	"goahk/internal/actions"
)

func TestRunCallback_StopSkipsRemainingSteps(t *testing.T) {
	t.Parallel()

	a := NewApp()
	a.Bind("Ctrl+H", Func(func(ctx *Context) error {
		ctx.Stop()
		return nil
	}), Log("never"))

	bindings, executor := compileTestBindings(t, a)
	res := executor.Execute(actions.ActionContext{Context: context.Background(), BindingID: bindings[0].ID}, bindings[0].Plan)
	if !res.Success {
		t.Fatalf("execution failed: %#v", res)
	}
	if len(res.Steps) != 2 {
		t.Fatalf("steps len = %d, want 2", len(res.Steps))
	}
	if got, want := res.Steps[1].Status, actions.StepStatusSkipped; got != want {
		t.Fatalf("step[1].status = %q, want %q", got, want)
	}
}
