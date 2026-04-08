package goahk

import (
	"context"
	"errors"
	"strings"
	"testing"

	"goahk/internal/actions"
)

func TestRunCallback_ErrorBubblesWithBindingStepContext(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	a := NewApp()
	a.Bind("Ctrl+H", Log("start"), Func(func(*Context) error { return boom }))

	bindings, executor := compileTestBindings(t, a)
	res := executor.Execute(actions.ActionContext{Context: context.Background(), BindingID: bindings[0].ID}, bindings[0].Plan)
	if res.Success {
		t.Fatalf("execution succeeded, want failure: %#v", res)
	}
	if len(res.Steps) != 2 {
		t.Fatalf("steps len = %d, want 2", len(res.Steps))
	}
	if got, want := res.Steps[1].Action, callbackActionName(0, 1); got != want {
		t.Fatalf("failed action = %q, want %q", got, want)
	}
	if !strings.Contains(res.Steps[1].Error, boom.Error()) {
		t.Fatalf("failed step error = %q, want to contain %q", res.Steps[1].Error, boom.Error())
	}
}
