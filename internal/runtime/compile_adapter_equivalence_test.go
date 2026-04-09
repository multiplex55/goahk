package runtime

import (
	"context"
	"reflect"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/config"
	"goahk/internal/hotkey"
	"goahk/internal/program"
)

func TestCompileRuntimeBindings_AdapterAndDirectProgramEquivalent(t *testing.T) {
	cfg := config.Config{
		Flows: []config.Flow{{ID: "flow.main", Steps: []config.FlowStep{{Action: "system.log"}}}},
		Hotkeys: []config.HotkeyBinding{
			{ID: "one", Hotkey: "ctrl+1", Steps: []config.Step{{Action: "system.log"}}},
			{ID: "two", Hotkey: "ctrl+2", Flow: "flow.main"},
		},
	}

	adapted, err := config.ToProgram(cfg)
	if err != nil {
		t.Fatalf("ToProgram() error = %v", err)
	}

	direct := program.Program{
		Bindings: []program.BindingSpec{
			{ID: "one", Hotkey: "ctrl+1", Steps: []program.StepSpec{{Action: "system.log"}}},
			{ID: "two", Hotkey: "ctrl+2", Flow: "flow.main"},
		},
		Options: program.Options{Flows: []program.FlowSpec{{ID: "flow.main", Steps: []program.FlowStepSpec{{Action: "system.log"}}}}},
	}

	fromAdapter, err := CompileRuntimeBindings(adapted, actions.NewRegistry())
	if err != nil {
		t.Fatalf("CompileRuntimeBindings(adapted) error = %v", err)
	}
	fromDirect, err := CompileRuntimeBindings(direct, actions.NewRegistry())
	if err != nil {
		t.Fatalf("CompileRuntimeBindings(direct) error = %v", err)
	}

	if !reflect.DeepEqual(fromAdapter, fromDirect) {
		t.Fatalf("compiled bindings mismatch\nadapter=%#v\ndirect=%#v", fromAdapter, fromDirect)
	}
}

func TestCompileRuntimeBindings_CompiledFlowExecutableViaRuntimeDispatch(t *testing.T) {
	cfg := config.Config{
		Flows:   []config.Flow{{ID: "flow.main", Steps: []config.FlowStep{{Action: "test.mark"}}}},
		Hotkeys: []config.HotkeyBinding{{ID: "two", Hotkey: "ctrl+2", Flow: "flow.main"}},
	}
	p, err := config.ToProgram(cfg)
	if err != nil {
		t.Fatalf("ToProgram() error = %v", err)
	}

	reg := actions.NewRegistry()
	called := make(chan struct{}, 1)
	if err := reg.Register("test.mark", func(actions.ActionContext, actions.Step) error {
		called <- struct{}{}
		return nil
	}); err != nil {
		t.Fatalf("register action: %v", err)
	}
	compiled, err := CompileRuntimeBindings(p, reg)
	if err != nil {
		t.Fatalf("CompileRuntimeBindings() error = %v", err)
	}

	events := make(chan hotkey.TriggerEvent, 1)
	shutdown := make(chan struct{})
	handle := DispatchHotkeyEventsWithBindingsHandle(context.Background(), shutdown, events, buildExecutableBindings(compiled), nil, actions.NewExecutor(reg), actions.ActionContext{}, nil, nil)
	events <- hotkey.TriggerEvent{BindingID: "two", Chord: hotkey.Chord{Modifiers: hotkey.ModCtrl, Key: "2"}}

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("compiled flow did not execute")
	}
	select {
	case res := <-handle.Results:
		if !res.Execution.Success {
			t.Fatalf("dispatch result should succeed: %#v", res)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for dispatch result")
	}
	close(shutdown)
	for range handle.Results {
	}
}
