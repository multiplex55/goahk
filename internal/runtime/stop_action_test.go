package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
	"goahk/internal/program"
)

func TestStopActionTerminatesRunLoopPredictably(t *testing.T) {
	events := make(chan hotkey.TriggerEvent, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdown := make(chan struct{})

	var stopCalls atomic.Int32
	base := actions.ActionContext{Stop: func(string) {
		stopCalls.Add(1)
		cancel()
	}}

	results := DispatchHotkeyEvents(ctx, shutdown, events, map[string]actions.Plan{
		"quit": {
			{Name: "runtime.stop"},
			{Name: "system.log", Params: map[string]string{"message": "must_skip"}},
		},
	}, nil, actions.NewExecutor(actions.NewRegistry()), base, nil, nil)

	events <- hotkey.TriggerEvent{BindingID: "quit", Chord: hotkey.Chord{Key: "Escape"}}

	<-ctx.Done()
	close(shutdown)
	for range results {
	}

	if got := stopCalls.Load(); got != 1 {
		t.Fatalf("stop calls = %d, want 1", got)
	}
}

func TestStopActionAvoidsDuplicateStopSideEffects(t *testing.T) {
	events := make(chan hotkey.TriggerEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdown := make(chan struct{})

	var stopCalls atomic.Int32
	base := actions.ActionContext{Stop: func(string) { stopCalls.Add(1) }}

	results := DispatchHotkeyEvents(ctx, shutdown, events, map[string]actions.Plan{
		"quit": {{Name: "runtime.stop"}, {Name: "runtime.stop"}},
	}, nil, actions.NewExecutor(actions.NewRegistry()), base, nil, nil)

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
	close(shutdown)
	for range results {
	}
}

func TestRuntimeControlHotkeys_EscapeGracefulVsShiftEscapeHardStop(t *testing.T) {
	p := program.Program{Bindings: []program.BindingSpec{
		{ID: "esc", Hotkey: "esc", Steps: []program.StepSpec{{Action: "system.log"}}},
		{ID: "hard", Hotkey: "shift+esc", Steps: []program.StepSpec{{Action: "system.log"}}},
	}}
	compiled, err := CompileRuntimeBindings(p, actions.NewRegistry())
	if err != nil {
		t.Fatalf("CompileRuntimeBindings() error = %v", err)
	}
	control := map[string]RuntimeControlCommand{}
	for _, b := range compiled {
		if b.ControlCommand != "" {
			control[b.ID] = RuntimeControlCommand(b.ControlCommand)
		}
	}

	events := make(chan hotkey.TriggerEvent, 2)
	shutdown := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var graceful, hard atomic.Int32
	results := DispatchHotkeyEvents(ctx, shutdown, events, map[string]actions.Plan{"esc": {}, "hard": {}}, control, actions.NewExecutor(actions.NewRegistry()), actions.ActionContext{}, nil, func(ev runtimeControlEvent) {
		switch ev.Command {
		case RuntimeControlStop:
			graceful.Add(1)
		case RuntimeControlHardStop:
			hard.Add(1)
		}
	})
	events <- hotkey.TriggerEvent{BindingID: "esc", Chord: hotkey.Chord{Key: "Escape"}}
	events <- hotkey.TriggerEvent{BindingID: "hard", Chord: hotkey.Chord{Modifiers: hotkey.ModShift, Key: "Escape"}}
	time.Sleep(30 * time.Millisecond)
	close(shutdown)
	for range results {
	}
	if graceful.Load() != 1 {
		t.Fatalf("graceful calls=%d want 1", graceful.Load())
	}
	if hard.Load() != 1 {
		t.Fatalf("hard-stop calls=%d want 1", hard.Load())
	}
}
