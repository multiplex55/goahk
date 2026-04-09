package runtime

import (
	"context"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/flow"
	"goahk/internal/hotkey"
)

func TestDispatchHotkeyEvents_BindingKindsParity(t *testing.T) {
	reg := actions.NewRegistry()
	calls := make(chan string, 3)
	if err := reg.Register("test.mark", func(ctx actions.ActionContext, _ actions.Step) error {
		calls <- ctx.BindingID
		return nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	if err := reg.RegisterCallback("named", func(ctx actions.CallbackContext) error {
		calls <- ctx.BindingID()
		return nil
	}); err != nil {
		t.Fatalf("register callback: %v", err)
	}

	events := make(chan hotkey.TriggerEvent, 3)
	shutdown := make(chan struct{})
	handle := DispatchHotkeyEventsWithBindingsHandle(
		context.Background(),
		shutdown,
		events,
		map[string]actions.ExecutableBinding{
			"direct":   {ID: "direct", Kind: actions.BindingKindPlan, Plan: actions.Plan{{Name: "test.mark"}}},
			"flow":     {ID: "flow", Kind: actions.BindingKindFlow, Flow: &flow.Definition{ID: "main", Steps: []flow.Step{{Action: "test.mark"}}}},
			"callback": {ID: "callback", Kind: actions.BindingKindCallback, Policy: actions.BindingExecutionPolicy{CallbackRef: "named"}},
		},
		nil,
		actions.NewExecutor(reg),
		actions.ActionContext{},
		nil,
		nil,
	)

	events <- hotkey.TriggerEvent{BindingID: "direct", Chord: hotkey.Chord{Key: "1"}}
	events <- hotkey.TriggerEvent{BindingID: "flow", Chord: hotkey.Chord{Key: "2"}}
	events <- hotkey.TriggerEvent{BindingID: "callback", Chord: hotkey.Chord{Key: "3"}}

	results := map[string]DispatchResult{}
	deadline := time.After(time.Second)
	for len(results) < 3 {
		select {
		case res := <-handle.Results:
			results[res.BindingID] = res
		case <-deadline:
			t.Fatalf("timed out waiting for dispatch results: %#v", results)
		}
	}
	close(shutdown)
	for range handle.Results {
	}

	if len(calls) != 3 {
		t.Fatalf("call count = %d, want 3", len(calls))
	}
	for _, id := range []string{"direct", "flow", "callback"} {
		res, ok := results[id]
		if !ok {
			t.Fatalf("missing result for %q", id)
		}
		if !res.Execution.Success {
			t.Fatalf("binding %q result should succeed: %#v", id, res)
		}
		if len(res.Actions) == 0 {
			t.Fatalf("binding %q actions should not be empty", id)
		}
	}
}
