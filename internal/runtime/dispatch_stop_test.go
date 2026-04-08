package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

func TestDispatchHotkeyEvents_RuntimeStopCausesCleanShutdown(t *testing.T) {
	reg := actions.NewRegistry()
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stopCalls atomic.Int32
	base := actions.ActionContext{
		Stop: func(string) {
			stopCalls.Add(1)
			cancel()
		},
	}

	results := DispatchHotkeyEvents(ctx, ctx.Done(), events, map[string]actions.Plan{
		"quit": {{Name: "runtime.stop"}},
	}, executor, base, nil)
	events <- hotkey.TriggerEvent{BindingID: "quit", Chord: hotkey.Chord{Key: "Esc"}}

	select {
	case got, ok := <-results:
		if !ok {
			t.Fatal("results closed before stop dispatch result was emitted")
		}
		if !got.Execution.Success {
			t.Fatalf("runtime.stop dispatch should succeed: %#v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for runtime.stop dispatch result")
	}

	select {
	case _, ok := <-results:
		if ok {
			t.Fatal("results should close after runtime.stop cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("results channel did not close after runtime.stop")
	}

	if stopCalls.Load() != 1 {
		t.Fatalf("stop callback count = %d, want 1", stopCalls.Load())
	}
}

func TestDispatchHotkeyEvents_RuntimeStopDuringDispatchNoDeadlock(t *testing.T) {
	reg := actions.NewRegistry()
	started := make(chan struct{}, 1)
	if err := reg.Register("test.slow", func(_ actions.ActionContext, _ actions.Step) error {
		started <- struct{}{}
		time.Sleep(20 * time.Millisecond)
		return nil
	}); err != nil {
		t.Fatalf("register test.slow: %v", err)
	}
	if err := reg.Register("test.never", func(_ actions.ActionContext, _ actions.Step) error {
		t.Fatal("test.never should be skipped after runtime.stop")
		return nil
	}); err != nil {
		t.Fatalf("register test.never: %v", err)
	}
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	results := DispatchHotkeyEvents(ctx, ctx.Done(), events, map[string]actions.Plan{
		"quit": {{Name: "test.slow"}, {Name: "runtime.stop"}, {Name: "test.never"}},
	}, executor, actions.ActionContext{
		Stop: func(string) {
			cancel()
		},
	}, nil)

	events <- hotkey.TriggerEvent{BindingID: "quit", Chord: hotkey.Chord{Key: "Esc"}}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("slow action did not start")
	}

	select {
	case got := <-results:
		if len(got.Execution.Steps) != 3 {
			t.Fatalf("step count = %d, want 3", len(got.Execution.Steps))
		}
		if got.Execution.Steps[2].Status != actions.StepStatusSkipped {
			t.Fatalf("last step status = %s, want %s", got.Execution.Steps[2].Status, actions.StepStatusSkipped)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for stop result")
	}

	select {
	case _, ok := <-results:
		if ok {
			t.Fatal("results channel should be closed after stop-triggered cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("results channel close timed out (possible deadlock)")
	}
}
