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
