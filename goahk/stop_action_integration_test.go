package goahk

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
	"goahk/internal/runtime"
)

func TestStopActionIntegration_EscapeBindingExitsWithoutDeadlock(t *testing.T) {
	t.Parallel()

	a := NewApp()
	a.Bind("Escape", Stop(), Log("never"))

	bindings, executor := compileTestBindings(t, a)
	plans := map[string]actions.Plan{bindings[0].ID: bindings[0].Plan}

	events := make(chan hotkey.TriggerEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stopCalls atomic.Int32
	base := actions.ActionContext{Stop: func(string) {
		stopCalls.Add(1)
		cancel()
	}}

	results := runtime.DispatchHotkeyEvents(ctx, ctx.Done(), events, plans, nil, executor, base, nil, nil)
	events <- hotkey.TriggerEvent{BindingID: bindings[0].ID, Chord: hotkey.Chord{Key: "Escape"}}

	select {
	case got, ok := <-results:
		if !ok {
			t.Fatal("results channel closed before stop result")
		}
		if !got.Execution.Success {
			t.Fatalf("execution failed: %#v", got.Execution)
		}
		if len(got.Execution.Steps) != 2 {
			t.Fatalf("steps len = %d, want 2", len(got.Execution.Steps))
		}
		if got.Execution.Steps[1].Status != actions.StepStatusSkipped {
			t.Fatalf("post-stop step status = %q, want skipped", got.Execution.Steps[1].Status)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stop integration result")
	}

	select {
	case _, ok := <-results:
		if ok {
			t.Fatal("results channel should close after stop")
		}
	case <-time.After(time.Second):
		t.Fatal("results channel did not close after stop")
	}

	if got := stopCalls.Load(); got != 1 {
		t.Fatalf("stop callback calls = %d, want 1", got)
	}
}
