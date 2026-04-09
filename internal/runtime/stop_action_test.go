package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

func TestStopActionTerminatesRunLoopPredictably(t *testing.T) {
	events := make(chan hotkey.TriggerEvent, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stopCalls atomic.Int32
	base := actions.ActionContext{Stop: func(string) {
		stopCalls.Add(1)
		cancel()
	}}

	results := DispatchHotkeyEvents(ctx, ctx.Done(), events, map[string]actions.Plan{
		"quit": {
			{Name: "runtime.stop"},
			{Name: "system.log", Params: map[string]string{"message": "must_skip"}},
		},
	}, actions.NewExecutor(actions.NewRegistry()), base, nil)

	events <- hotkey.TriggerEvent{BindingID: "quit", Chord: hotkey.Chord{Key: "Escape"}}

	select {
	case got, ok := <-results:
		if !ok {
			t.Fatal("results channel closed before stop result")
		}
		if len(got.Execution.Steps) != 2 {
			t.Fatalf("steps len = %d, want 2", len(got.Execution.Steps))
		}
		if got.Execution.Steps[0].Status != actions.StepStatusSuccess {
			t.Fatalf("stop action status = %s, want success", got.Execution.Steps[0].Status)
		}
		if got.Execution.Steps[1].Status != actions.StepStatusSkipped {
			t.Fatalf("second action status = %s, want skipped", got.Execution.Steps[1].Status)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for dispatch result")
	}

	select {
	case _, ok := <-results:
		if ok {
			t.Fatal("results channel should be closed after stop")
		}
	case <-time.After(time.Second):
		t.Fatal("results channel did not close after stop")
	}

	if got := stopCalls.Load(); got != 1 {
		t.Fatalf("stop calls = %d, want 1", got)
	}
}

func TestStopActionAvoidsDuplicateStopSideEffects(t *testing.T) {
	events := make(chan hotkey.TriggerEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stopCalls atomic.Int32
	base := actions.ActionContext{Stop: func(string) { stopCalls.Add(1) }}

	results := DispatchHotkeyEvents(ctx, ctx.Done(), events, map[string]actions.Plan{
		"quit": {{Name: "runtime.stop"}, {Name: "runtime.stop"}},
	}, actions.NewExecutor(actions.NewRegistry()), base, nil)

	events <- hotkey.TriggerEvent{BindingID: "quit", Chord: hotkey.Chord{Key: "Escape"}}

	select {
	case got, ok := <-results:
		if !ok {
			t.Fatal("results channel closed unexpectedly")
		}
		if len(got.Execution.Steps) != 2 {
			t.Fatalf("steps len = %d, want 2", len(got.Execution.Steps))
		}
		if got.Execution.Steps[1].Status != actions.StepStatusSkipped {
			t.Fatalf("second runtime.stop status = %s, want skipped", got.Execution.Steps[1].Status)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for duplicate-stop result")
	}

	if got := stopCalls.Load(); got != 1 {
		t.Fatalf("stop callback calls = %d, want 1", got)
	}
}
