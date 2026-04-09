package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
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
