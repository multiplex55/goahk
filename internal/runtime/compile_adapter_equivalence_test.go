package runtime

import (
	"reflect"
	"testing"

	"goahk/internal/actions"
	"goahk/internal/config"
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
