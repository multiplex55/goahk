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

func TestDispatchHotkeyEvents_BlockingActionDoesNotPreventStopHotkey(t *testing.T) {
	reg := actions.NewRegistry()
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	if err := reg.Register("test.block", func(ctx actions.ActionContext, _ actions.Step) error {
		started <- struct{}{}
		select {
		case <-release:
			return nil
		case <-ctx.Context.Done():
			return ctx.Context.Err()
		}
	}); err != nil {
		t.Fatalf("register test.block: %v", err)
	}
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdown := make(chan struct{})

	var stopCalls atomic.Int32
	results := DispatchHotkeyEvents(ctx, shutdown, events, map[string]actions.Plan{
		"work": {{Name: "test.block"}},
		"quit": {},
	}, map[string]RuntimeControlCommand{"quit": RuntimeControlStop}, executor, actions.ActionContext{}, nil, func(ev runtimeControlEvent) {
		if ev.Command == RuntimeControlStop {
			stopCalls.Add(1)
			cancel()
		}
	})

	events <- hotkey.TriggerEvent{BindingID: "work", Chord: hotkey.Chord{Key: "F9"}}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("blocking job did not start")
	}
	events <- hotkey.TriggerEvent{BindingID: "quit", Chord: hotkey.Chord{Key: "Escape"}}

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("stop hotkey did not cancel context while work job was blocked")
	}
	close(release)
	close(shutdown)
	for range results {
	}
	if stopCalls.Load() != 1 {
		t.Fatalf("stop calls = %d, want 1", stopCalls.Load())
	}
}

func TestDispatchHotkeyEvents_CompiledExplicitControlHardStopDispatchesControlEvent(t *testing.T) {
	p := program.Program{
		Bindings: []program.BindingSpec{
			{ID: "hard", Hotkey: "ctrl+f12", Steps: []program.StepSpec{{Action: "runtime.control_hard_stop"}}},
		},
	}
	bindings, err := CompileRuntimeBindings(p, actions.NewRegistry())
	if err != nil {
		t.Fatalf("CompileRuntimeBindings() error = %v", err)
	}
	control := map[string]RuntimeControlCommand{}
	for _, b := range bindings {
		if b.ControlCommand != "" {
			control[b.ID] = RuntimeControlCommand(b.ControlCommand)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan hotkey.TriggerEvent, 1)
	var hardStops atomic.Int32
	results := DispatchHotkeyEvents(ctx, ctx.Done(), events, map[string]actions.Plan{"hard": {}}, control, actions.NewExecutor(actions.NewRegistry()), actions.ActionContext{}, nil, func(ev runtimeControlEvent) {
		if ev.Command == RuntimeControlHardStop {
			hardStops.Add(1)
			cancel()
		}
	})
	events <- hotkey.TriggerEvent{BindingID: "hard", Chord: hotkey.Chord{Modifiers: hotkey.ModCtrl, Key: "F12"}}
	for range results {
	}
	if got := hardStops.Load(); got != 1 {
		t.Fatalf("hard stop events = %d, want 1", got)
	}
}

func TestDispatchHotkeyEvents_ControlStopBypassesActionPlanExecution(t *testing.T) {
	reg := actions.NewRegistry()
	var ran atomic.Int32
	if err := reg.Register("test.should_not_run", func(ctx actions.ActionContext, _ actions.Step) error {
		ran.Add(1)
		return nil
	}); err != nil {
		t.Fatalf("register test.should_not_run: %v", err)
	}
	executor := actions.NewExecutor(reg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan hotkey.TriggerEvent, 1)
	shutdown := make(chan struct{})

	var stopCalls atomic.Int32
	results := DispatchHotkeyEvents(ctx, shutdown, events, map[string]actions.Plan{
		"quit": {{Name: "test.should_not_run"}},
	}, map[string]RuntimeControlCommand{"quit": RuntimeControlStop}, executor, actions.ActionContext{}, nil, func(ev runtimeControlEvent) {
		if ev.Command == RuntimeControlStop {
			stopCalls.Add(1)
		}
	})

	events <- hotkey.TriggerEvent{BindingID: "quit", Chord: hotkey.Chord{Key: "Escape"}}
	deadline := time.After(time.Second)
	for stopCalls.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for control stop dispatch")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
	close(shutdown)
	cancel()
	for range results {
	}
	if got := stopCalls.Load(); got != 1 {
		t.Fatalf("control stop calls = %d, want 1", got)
	}
	if got := ran.Load(); got != 0 {
		t.Fatalf("control-stop binding executed %d action steps, want 0", got)
	}
}

func TestDispatchHotkeyEvents_StopActionFollowsActionSequenceSemantics(t *testing.T) {
	reg := actions.NewRegistry()
	var firstRan atomic.Int32
	if err := reg.Register("test.first", func(ctx actions.ActionContext, _ actions.Step) error {
		firstRan.Add(1)
		return nil
	}); err != nil {
		t.Fatalf("register test.first: %v", err)
	}
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdown := make(chan struct{})
	var stopCalls atomic.Int32
	base := actions.ActionContext{Stop: func(string) {
		stopCalls.Add(1)
		cancel()
	}}

	results := DispatchHotkeyEvents(ctx, shutdown, events, map[string]actions.Plan{
		"work": {
			{Name: "test.first"},
			{Name: "runtime.stop"},
			{Name: "system.log", Params: map[string]string{"message": "must_skip"}},
		},
	}, nil, executor, base, nil, nil)

	events <- hotkey.TriggerEvent{BindingID: "work", Chord: hotkey.Chord{Key: "F8"}}
	select {
	case got := <-results:
		if len(got.Execution.Steps) != 3 {
			t.Fatalf("steps len = %d, want 3", len(got.Execution.Steps))
		}
		if got.Execution.Steps[0].Status != actions.StepStatusSuccess {
			t.Fatalf("first step status = %q, want success", got.Execution.Steps[0].Status)
		}
		if got.Execution.Steps[1].Status != actions.StepStatusSuccess {
			t.Fatalf("runtime.stop step status = %q, want success", got.Execution.Steps[1].Status)
		}
		if got.Execution.Steps[2].Status != actions.StepStatusSkipped {
			t.Fatalf("post-stop step status = %q, want skipped", got.Execution.Steps[2].Status)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for action-stop result")
	}
	if got := firstRan.Load(); got != 1 {
		t.Fatalf("first step calls = %d, want 1", got)
	}
	if got := stopCalls.Load(); got != 1 {
		t.Fatalf("stop callback calls = %d, want 1", got)
	}
}
