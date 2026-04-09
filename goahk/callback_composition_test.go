package goahk

import (
	"fmt"
	"testing"

	"goahk/internal/actions"
)

func TestCallbackComposition_DeclarativeAndCallbackOrderingAndSharedContext(t *testing.T) {
	t.Parallel()

	var callbackBindingID string
	app := NewApp().Bind(
		"Ctrl+K",
		Log("before"),
		Func(func(ctx *Context) error {
			callbackBindingID = ctx.BindingID()
			current := 0
			if raw, ok := ctx.AppState.Get("count"); ok {
				fmt.Sscanf(raw, "%d", &current)
			}
			current++
			ctx.AppState.Set("count", fmt.Sprintf("%d", current))
			ctx.Vars["count"] = fmt.Sprintf("%d", current)
			ctx.Vars["phase"] = "callback"
			return nil
		}),
		Log("after"),
	)

	p, _, callbacks := app.runtimeArtifacts()
	steps := p.Bindings[0].Steps
	if got, want := len(steps), 3; got != want {
		t.Fatalf("steps len = %d, want %d", got, want)
	}
	if got, want := steps[0].Action, "system.log"; got != want {
		t.Fatalf("step[0] action = %q, want %q", got, want)
	}
	if got, want := steps[1].Action, callbackActionName(0, 1); got != want {
		t.Fatalf("step[1] action = %q, want %q", got, want)
	}
	if got, want := steps[2].Action, "system.log"; got != want {
		t.Fatalf("step[2] action = %q, want %q", got, want)
	}
	if got, want := len(callbacks), 1; got != want {
		t.Fatalf("callbacks len = %d, want %d", got, want)
	}

	actionCtx := &actions.ActionContext{
		BindingID: "binding_1",
		Metadata:  map[string]string{"phase": "start"},
	}
	ctx := newContext(actionCtx, app.state)
	if err := callbacks[0].fn(ctx); err != nil {
		t.Fatalf("callback fn error = %v", err)
	}
	syncVarsToActionContext(ctx)

	if got, want := callbackBindingID, "binding_1"; got != want {
		t.Fatalf("callback BindingID = %q, want %q", got, want)
	}
	if got, want := actionCtx.Metadata["count"], "1"; got != want {
		t.Fatalf("metadata[count] = %q, want %q", got, want)
	}
	if got, want := actionCtx.Metadata["phase"], "callback"; got != want {
		t.Fatalf("metadata[phase] = %q, want %q", got, want)
	}
	if got, ok := app.state.Get("count"); !ok || got != "1" {
		t.Fatalf("app state count = %q, ok=%v, want 1,true", got, ok)
	}
}
