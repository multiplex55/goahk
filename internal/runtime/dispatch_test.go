package runtime

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
	"goahk/internal/services/messagebox"
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
	}, nil, executor, actions.ActionContext{}, nil, nil)
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

	results := DispatchHotkeyEvents(context.Background(), shutdown, events, map[string]actions.Plan{"known": {{Name: "test.mark"}}}, nil, executor, actions.ActionContext{}, sink, nil)
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
	}, nil, executor, actions.ActionContext{}, sink, nil)
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
	}, nil, executor, actions.ActionContext{}, nil, nil)

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

func TestDispatchHotkeyEvents_CallbackOrderingAndArguments(t *testing.T) {
	reg := actions.NewRegistry()
	var mu sync.Mutex
	var calls []string
	mark := func(name string) func(actions.ActionContext, actions.Step) error {
		return func(ctx actions.ActionContext, step actions.Step) error {
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, name+":"+ctx.BindingID+":"+ctx.TriggerText+":"+step.Params["tag"])
			return nil
		}
	}
	if err := reg.Register("test.first", mark("first")); err != nil {
		t.Fatalf("register first action: %v", err)
	}
	if err := reg.Register("test.second", mark("second")); err != nil {
		t.Fatalf("register second action: %v", err)
	}
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 1)
	shutdown := make(chan struct{})

	results := DispatchHotkeyEvents(context.Background(), shutdown, events, map[string]actions.Plan{
		"binding.sequence": {
			{Name: "test.first", Params: map[string]string{"tag": "one"}},
			{Name: "test.second", Params: map[string]string{"tag": "two"}},
		},
	}, nil, executor, actions.ActionContext{}, nil, nil)

	events <- hotkey.TriggerEvent{
		BindingID: "binding.sequence",
		Chord:     hotkey.Chord{Modifiers: hotkey.ModCtrl | hotkey.ModShift, Key: "S"},
	}

	select {
	case res := <-results:
		if !res.Execution.Success {
			t.Fatalf("dispatch should succeed, got %#v", res)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for dispatch result")
	}
	close(shutdown)
	for range results {
	}

	mu.Lock()
	gotCalls := append([]string(nil), calls...)
	mu.Unlock()
	wantCalls := []string{
		"first:binding.sequence:Ctrl+Shift+S:one",
		"second:binding.sequence:Ctrl+Shift+S:two",
	}
	if !reflect.DeepEqual(gotCalls, wantCalls) {
		t.Fatalf("calls = %v, want %v", gotCalls, wantCalls)
	}
}

type dispatchRecordingMessageBox struct {
	mu    sync.Mutex
	calls []messagebox.Request
}

func (d *dispatchRecordingMessageBox) Show(_ context.Context, req messagebox.Request) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls = append(d.calls, req)
	return nil
}

func (d *dispatchRecordingMessageBox) requests() []messagebox.Request {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]messagebox.Request, len(d.calls))
	copy(out, d.calls)
	return out
}

func TestDispatchHotkeyEvents_MessageBoxActionReceivesExpectedDialogPayload(t *testing.T) {
	reg := actions.NewRegistry()
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 1)
	shutdown := make(chan struct{})
	box := &dispatchRecordingMessageBox{}

	results := DispatchHotkeyEvents(
		context.Background(),
		shutdown,
		events,
		map[string]actions.Plan{
			"hk.dialog": {
				{Name: "system.message_box", Params: map[string]string{"title": "Hotkey Test", "body": "Pressed Ctrl+Alt+D", "icon": "info", "options": "ok"}},
			},
		},
		nil,
		executor,
		actions.ActionContext{Services: actions.Services{MessageBox: box}},
		nil,
		nil,
	)

	events <- hotkey.TriggerEvent{BindingID: "hk.dialog", Chord: hotkey.Chord{Modifiers: hotkey.ModCtrl | hotkey.ModAlt, Key: "D"}}
	select {
	case res := <-results:
		if !res.Execution.Success {
			t.Fatalf("dialog dispatch should succeed, got %#v", res)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for dialog dispatch result")
	}
	close(shutdown)
	for range results {
	}

	requests := box.requests()
	if len(requests) != 1 {
		t.Fatalf("message box call count = %d, want 1", len(requests))
	}
	if requests[0].Title != "Hotkey Test" || requests[0].Body != "Pressed Ctrl+Alt+D" {
		t.Fatalf("message box request = %#v", requests[0])
	}
}
