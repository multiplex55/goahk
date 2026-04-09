package runtime

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

func TestRuntimeIntegration_CustomCallbackStopSafe(t *testing.T) {
	reg := actions.NewRegistry()
	started := make(chan struct{}, 1)
	var canceled atomic.Bool
	reg.MustRegisterCallback("slow", func(ctx actions.CallbackContext) error {
		started <- struct{}{}
		<-ctx.Done()
		canceled.Store(ctx.IsCancelled())
		return ctx.Err()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan hotkey.TriggerEvent, 4)
	shutdown := make(chan struct{})

	control := map[string]RuntimeControlCommand{"stop": RuntimeControlStop}
	results := DispatchHotkeyEvents(ctx, shutdown, events, map[string]actions.Plan{
		"work": {{Name: actions.CallbackActionName, Params: map[string]string{"callback_ref": "slow"}}},
		"stop": {},
	}, control, actions.NewExecutor(reg), actions.ActionContext{}, nil, func(ev runtimeControlEvent) {
		if ev.Command == RuntimeControlStop {
			cancel()
		}
	})

	events <- hotkey.TriggerEvent{BindingID: "work", Chord: hotkey.Chord{Modifiers: hotkey.ModCtrl | hotkey.ModShift, Key: "R"}}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("callback did not start")
	}
	events <- hotkey.TriggerEvent{BindingID: "stop", Chord: hotkey.Chord{Key: "Escape"}}

	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("context was not canceled by stop")
	}

	close(shutdown)
	for range results {
	}
	if !canceled.Load() {
		t.Fatal("callback did not observe cancellation")
	}
}

func TestRuntimeIntegration_WindowInspectAndMoveWindow(t *testing.T) {
	reg := actions.NewRegistry()
	var mu sync.Mutex
	calls := []string{}
	reg.MustRegister("test.window.inspect", func(_ actions.ActionContext, _ actions.Step) error {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, "inspect")
		return nil
	})
	reg.MustRegister("test.window.move", func(_ actions.ActionContext, step actions.Step) error {
		if step.Params["x"] == "" || step.Params["y"] == "" {
			return errors.New("missing coordinates")
		}
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, "move")
		return nil
	})

	events := make(chan hotkey.TriggerEvent, 2)
	shutdown := make(chan struct{})
	results := DispatchHotkeyEvents(context.Background(), shutdown, events, map[string]actions.Plan{
		"inspect": {{Name: "test.window.inspect"}},
		"move":    {{Name: "test.window.move", Params: map[string]string{"x": "960", "y": "0"}}},
	}, nil, actions.NewExecutor(reg), actions.ActionContext{}, nil, nil)

	events <- hotkey.TriggerEvent{BindingID: "inspect"}
	events <- hotkey.TriggerEvent{BindingID: "move"}

	for i := 0; i < 2; i++ {
		select {
		case r := <-results:
			if !r.Execution.Success {
				t.Fatalf("execution failed: %#v", r)
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for result")
		}
	}
	close(shutdown)
	for range results {
	}

	mu.Lock()
	defer mu.Unlock()
	if len(calls) != 2 {
		t.Fatalf("calls len = %d, want 2", len(calls))
	}
}

func TestRuntimeIntegration_MouseMoveAndClickPipeline(t *testing.T) {
	reg := actions.NewRegistry()
	var mu sync.Mutex
	ordered := []string{}
	reg.MustRegister("test.mouse.move", func(_ actions.ActionContext, _ actions.Step) error {
		mu.Lock()
		ordered = append(ordered, "move")
		mu.Unlock()
		return nil
	})
	reg.MustRegister("test.mouse.click", func(_ actions.ActionContext, _ actions.Step) error {
		mu.Lock()
		ordered = append(ordered, "click")
		mu.Unlock()
		return nil
	})

	events := make(chan hotkey.TriggerEvent, 1)
	shutdown := make(chan struct{})
	results := DispatchHotkeyEvents(context.Background(), shutdown, events, map[string]actions.Plan{
		"mouse": {
			{Name: "test.mouse.move", Params: map[string]string{"x": "1200", "y": "450"}},
			{Name: "test.mouse.click", Params: map[string]string{"button": "left"}},
		},
	}, nil, actions.NewExecutor(reg), actions.ActionContext{}, nil, nil)

	events <- hotkey.TriggerEvent{BindingID: "mouse"}
	select {
	case r := <-results:
		if !r.Execution.Success {
			t.Fatalf("execution failed: %#v", r)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for mouse pipeline result")
	}
	close(shutdown)
	for range results {
	}

	mu.Lock()
	defer mu.Unlock()
	if len(ordered) != 2 || ordered[0] != "move" || ordered[1] != "click" {
		t.Fatalf("ordered calls = %v, want [move click]", ordered)
	}
}

func TestRuntimeIntegration_LongRunningTaskWithReplacePolicy(t *testing.T) {
	reg := actions.NewRegistry()
	var starts atomic.Int32
	var firstStarted atomic.Bool
	releaseSecond := make(chan struct{})
	reg.MustRegister("test.blocking", func(ctx actions.ActionContext, _ actions.Step) error {
		run := starts.Add(1)
		if run == 1 {
			firstStarted.Store(true)
			<-ctx.Context.Done()
			return ctx.Context.Err()
		}
		select {
		case <-releaseSecond:
			return nil
		case <-ctx.Context.Done():
			return ctx.Context.Err()
		}
	})

	events := make(chan hotkey.TriggerEvent, 2)
	shutdown := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bindings := map[string]actions.ExecutableBinding{
		"replace": {
			ID:     "replace",
			Kind:   actions.BindingKindPlan,
			Plan:   actions.Plan{{Name: "test.blocking"}},
			Policy: actions.BindingExecutionPolicy{Concurrency: "replace"},
		},
	}
	results := DispatchHotkeyEventsWithBindingsHandle(ctx, shutdown, events, bindings, nil, actions.NewExecutor(reg), actions.ActionContext{}, nil, nil).Results

	events <- hotkey.TriggerEvent{BindingID: "replace"}
	for !firstStarted.Load() {
		time.Sleep(5 * time.Millisecond)
	}
	events <- hotkey.TriggerEvent{BindingID: "replace"}
	time.Sleep(40 * time.Millisecond)
	close(releaseSecond)
	close(shutdown)

	got := make([]DispatchResult, 0, 2)
	for r := range results {
		got = append(got, r)
	}
	if len(got) != 2 {
		t.Fatalf("results len = %d, want 2", len(got))
	}
	if starts.Load() != 2 {
		t.Fatalf("starts = %d, want 2", starts.Load())
	}
}

func TestRuntimeIntegration_EmergencyStopGracefulVsHard(t *testing.T) {
	events := make(chan hotkey.TriggerEvent, 4)
	shutdown := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var graceful atomic.Int32
	var hard atomic.Int32
	control := map[string]RuntimeControlCommand{
		"esc":  RuntimeControlStop,
		"hard": RuntimeControlHardStop,
	}
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
	_ = results
	time.Sleep(40 * time.Millisecond)

	if graceful.Load() != 1 {
		t.Fatalf("graceful calls = %d, want 1", graceful.Load())
	}
	if hard.Load() != 1 {
		t.Fatalf("hard calls = %d, want 1", hard.Load())
	}
}
