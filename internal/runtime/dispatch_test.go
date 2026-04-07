package runtime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

func TestDispatchHotkeyEvents_KnownBindingExecutesExpectedPlan(t *testing.T) {
	reg := actions.NewRegistry()
	executed := make(chan string, 1)
	if err := reg.Register("test.mark", func(ctx actions.ActionContext, _ actions.Step) error {
		executed <- ctx.BindingID
		return nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}

	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 1)
	results := DispatchHotkeyEvents(context.Background(), make(chan struct{}), events, map[string]actions.Plan{
		"binding.paste": {{Name: "test.mark"}},
	}, executor, actions.ActionContext{}, nil)
	events <- hotkey.TriggerEvent{BindingID: "binding.paste", Chord: hotkey.Chord{Key: "V"}}

	select {
	case got := <-executed:
		if got != "binding.paste" {
			t.Fatalf("executed binding context = %q, want %q", got, "binding.paste")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for executor call")
	}

	select {
	case result := <-results:
		if result.BindingID != "binding.paste" || !result.Execution.Success {
			t.Fatalf("unexpected result: %#v", result)
		}
		if len(result.Actions) != 1 || result.Actions[0] != "test.mark" {
			t.Fatalf("actions=%v, want [test.mark]", result.Actions)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for dispatch result")
	}
}

func TestDispatchHotkeyEvents_UnknownBindingLoggedAndSkipped(t *testing.T) {
	reg := actions.NewRegistry()
	_ = reg.Register("test.mark", func(actions.ActionContext, actions.Step) error { return nil })
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 1)
	shutdown := make(chan struct{})

	var mu sync.Mutex
	logs := make([]DispatchLogEntry, 0, 2)
	sink := func(_ context.Context, entry DispatchLogEntry) {
		mu.Lock()
		defer mu.Unlock()
		logs = append(logs, entry)
	}

	results := DispatchHotkeyEvents(context.Background(), shutdown, events, map[string]actions.Plan{"known": {{Name: "test.mark"}}}, executor, actions.ActionContext{}, sink)
	events <- hotkey.TriggerEvent{BindingID: "missing"}
	time.Sleep(20 * time.Millisecond)
	close(shutdown)

	for range results {
	}

	mu.Lock()
	defer mu.Unlock()
	foundUnknown := false
	for _, entry := range logs {
		if entry.Event == "dispatch_unknown_binding" {
			foundUnknown = true
			if entry.BindingID != "missing" {
				t.Fatalf("unknown-binding log binding = %q, want %q", entry.BindingID, "missing")
			}
		}
	}
	if !foundUnknown {
		t.Fatalf("expected unknown-binding log entry, got %#v", logs)
	}
}

func TestDispatchHotkeyEvents_ExecutorErrorCapturedAndLogged(t *testing.T) {
	reg := actions.NewRegistry()
	if err := reg.Register("test.fail", func(actions.ActionContext, actions.Step) error {
		return errors.New("boom")
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 1)
	shutdown := make(chan struct{})

	var mu sync.Mutex
	logs := []DispatchLogEntry{}
	sink := func(_ context.Context, entry DispatchLogEntry) {
		mu.Lock()
		defer mu.Unlock()
		logs = append(logs, entry)
	}

	results := DispatchHotkeyEvents(context.Background(), shutdown, events, map[string]actions.Plan{
		"broken": {{Name: "test.fail"}},
	}, executor, actions.ActionContext{}, sink)
	events <- hotkey.TriggerEvent{BindingID: "broken"}

	var result DispatchResult
	select {
	case result = <-results:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for error result")
	}
	close(shutdown)
	for range results {
	}

	if result.Error == "" {
		t.Fatalf("expected dispatch result error, got %#v", result)
	}
	if result.Error != "boom" {
		t.Fatalf("result.Error = %q, want %q", result.Error, "boom")
	}

	mu.Lock()
	defer mu.Unlock()
	foundFailure := false
	for _, entry := range logs {
		if entry.Event == "dispatch_failure_detail" {
			foundFailure = true
			if entry.Error != "boom" {
				t.Fatalf("failure log error = %q, want %q", entry.Error, "boom")
			}
		}
	}
	if !foundFailure {
		t.Fatalf("expected dispatch_failure_detail log, got %#v", logs)
	}
}

func TestDispatchHotkeyEvents_StopsImmediatelyAfterShutdownSignal(t *testing.T) {
	reg := actions.NewRegistry()
	executions := make(chan struct{}, 4)
	if err := reg.Register("test.mark", func(actions.ActionContext, actions.Step) error {
		executions <- struct{}{}
		return nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 8)
	shutdown := make(chan struct{})

	results := DispatchHotkeyEvents(context.Background(), shutdown, events, map[string]actions.Plan{
		"hk": {{Name: "test.mark"}},
	}, executor, actions.ActionContext{}, nil)

	events <- hotkey.TriggerEvent{BindingID: "hk"}
	select {
	case <-results:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first result")
	}

	close(shutdown)
	events <- hotkey.TriggerEvent{BindingID: "hk"}
	for range results {
	}

	if got := len(executions); got != 1 {
		t.Fatalf("execution count = %d, want 1", got)
	}
}
