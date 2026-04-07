package app

import (
	"context"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/config"
	"goahk/internal/hotkey"
)

func TestCompileRuntimeBindingsAndDispatch(t *testing.T) {
	registry := actions.NewRegistry()
	called := false
	if err := registry.Register("test.mark", func(ctx actions.ActionContext, step actions.Step) error {
		called = ctx.BindingID == "paste"
		return nil
	}); err != nil {
		t.Fatalf("register custom action: %v", err)
	}

	cfg := config.Config{Hotkeys: []config.HotkeyBinding{{
		ID:     "paste",
		Hotkey: "shift+ctrl+v",
		Steps:  []config.Step{{Action: "test.mark"}},
	}}}

	bindings, err := CompileRuntimeBindings(cfg, registry)
	if err != nil {
		t.Fatalf("CompileRuntimeBindings() error = %v", err)
	}
	if len(bindings) != 1 || bindings[0].Chord.String() != "Ctrl+Shift+V" {
		t.Fatalf("compiled bindings unexpected: %#v", bindings)
	}

	executor := actions.NewExecutor(registry)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	trigger := make(chan hotkey.TriggerEvent, 1)
	results := DispatchHotkeyEvents(ctx, trigger, map[string]actions.Plan{"paste": bindings[0].Plan}, executor, actions.ActionContext{})

	trigger <- hotkey.TriggerEvent{BindingID: "paste", Chord: bindings[0].Chord, TriggeredAt: time.Now()}

	select {
	case res := <-results:
		if !res.Success {
			t.Fatalf("expected success result: %#v", res)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for dispatch result")
	}

	if !called {
		t.Fatal("expected handler to receive mapped binding context")
	}
}

func TestCompileRuntimeBindings_FlowReference(t *testing.T) {
	registry := actions.NewRegistry()
	cfg := config.Config{
		Flows:   []config.Flow{{ID: "f1", Steps: []config.FlowStep{{Action: "system.log"}}}},
		Hotkeys: []config.HotkeyBinding{{ID: "paste", Hotkey: "shift+ctrl+v", Flow: "f1"}},
	}
	bindings, err := CompileRuntimeBindings(cfg, registry)
	if err != nil {
		t.Fatalf("CompileRuntimeBindings() error = %v", err)
	}
	if len(bindings) != 1 || bindings[0].Flow == nil || len(bindings[0].Flow.Steps) != 1 {
		t.Fatalf("compiled flow binding unexpected: %#v", bindings)
	}
}
