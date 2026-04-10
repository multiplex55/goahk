package runtime

import (
	"testing"

	"goahk/internal/actions"
	"goahk/internal/program"
)

func TestCompileRuntimeBindingsFromProgram(t *testing.T) {
	p := program.Program{
		Bindings: []program.BindingSpec{{
			ID:                "paste",
			Hotkey:            "ctrl+shift+v",
			Steps:             []program.StepSpec{{Action: "system.log"}},
			ConcurrencyPolicy: program.ConcurrencyPolicyQueueOne,
		}},
	}

	bindings, err := CompileRuntimeBindings(p, actions.NewRegistry())
	if err != nil {
		t.Fatalf("CompileRuntimeBindings() error = %v", err)
	}
	if got := len(bindings); got != 1 {
		t.Fatalf("len(bindings) = %d, want 1", got)
	}
	if got := bindings[0].ID; got != "paste" {
		t.Fatalf("binding id = %q, want %q", got, "paste")
	}
	if got := bindings[0].Chord.String(); got != "Ctrl+Shift+V" {
		t.Fatalf("chord = %q, want %q", got, "Ctrl+Shift+V")
	}
	if got := len(bindings[0].Plan); got != 1 {
		t.Fatalf("plan length = %d, want 1", got)
	}
	if got := bindings[0].Policy; got != program.ConcurrencyPolicyQueueOne {
		t.Fatalf("policy = %q, want %q", got, program.ConcurrencyPolicyQueueOne)
	}
}

func TestCompileRuntimeBindings_EscapeWithNormalActionIsNotControlCommand(t *testing.T) {
	p := program.Program{
		Bindings: []program.BindingSpec{{
			ID:     "escape_log",
			Hotkey: "escape",
			Steps:  []program.StepSpec{{Action: "system.log"}},
		}},
	}
	bindings, err := CompileRuntimeBindings(p, actions.NewRegistry())
	if err != nil {
		t.Fatalf("CompileRuntimeBindings() error = %v", err)
	}
	if got := bindings[0].ControlCommand; got != "" {
		t.Fatalf("control command = %q, want empty", got)
	}
}

func TestCompileRuntimeBindings_ExplicitControlActionsCreateControlCommands(t *testing.T) {
	p := program.Program{
		Bindings: []program.BindingSpec{
			{ID: "graceful", Hotkey: "f12", Steps: []program.StepSpec{{Action: "runtime.control_stop"}}},
			{ID: "hard", Hotkey: "ctrl+f12", Steps: []program.StepSpec{{Action: "runtime.control_hard_stop"}}},
		},
	}
	bindings, err := CompileRuntimeBindings(p, actions.NewRegistry())
	if err != nil {
		t.Fatalf("CompileRuntimeBindings() error = %v", err)
	}
	if got := bindings[0].ControlCommand; got != "stop" {
		t.Fatalf("graceful control command = %q, want stop", got)
	}
	if got := bindings[1].ControlCommand; got != "hard_stop" {
		t.Fatalf("hard control command = %q, want hard_stop", got)
	}
}

func TestCompileRuntimeBindings_ImplicitEscapeCompatibilityFlag(t *testing.T) {
	p := program.Program{
		Bindings: []program.BindingSpec{
			{ID: "esc", Hotkey: "escape", Steps: []program.StepSpec{{Action: "system.log"}}},
			{ID: "hard", Hotkey: "shift+escape", Steps: []program.StepSpec{{Action: "system.log"}}},
		},
		Options: program.Options{EnableImplicitEscapeControls: true},
	}
	bindings, err := CompileRuntimeBindings(p, actions.NewRegistry())
	if err != nil {
		t.Fatalf("CompileRuntimeBindings() error = %v", err)
	}
	if got := bindings[0].ControlCommand; got != "stop" {
		t.Fatalf("escape control command = %q, want stop", got)
	}
	if got := bindings[1].ControlCommand; got != "hard_stop" {
		t.Fatalf("shift+escape control command = %q, want hard_stop", got)
	}
}
