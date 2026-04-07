package app

import (
	"strings"
	"testing"

	"goahk/internal/actions"
	"goahk/internal/program"
)

func TestCompileRuntimeBindingsFromProgram_ValidProgram(t *testing.T) {
	registry := actions.NewRegistry()
	p := program.Program{Bindings: []program.BindingSpec{{
		ID:     "paste",
		Hotkey: "shift+ctrl+v",
		Steps:  []program.StepSpec{{Action: "system.log"}},
	}}}
	bindings, err := CompileRuntimeBindingsFromProgram(p, registry)
	if err != nil {
		t.Fatalf("CompileRuntimeBindingsFromProgram() error = %v", err)
	}
	if len(bindings) != 1 || bindings[0].Chord.String() != "Ctrl+Shift+V" {
		t.Fatalf("compiled bindings unexpected: %#v", bindings)
	}
}

func TestCompileRuntimeBindingsFromProgram_UnknownActionIncludesPathAndDiagnostics(t *testing.T) {
	p := program.Program{Bindings: []program.BindingSpec{{
		ID:     "open-terminal",
		Hotkey: "ctrl+alt+t",
		Steps:  []program.StepSpec{{Action: "sendKeys"}},
	}}}
	_, err := CompileRuntimeBindingsFromProgram(p, actions.NewRegistry())
	if err == nil {
		t.Fatal("CompileRuntimeBindingsFromProgram() error = nil, want failure")
	}
	msg := err.Error()
	for _, token := range []string{`binding "open-terminal"`, `binding/actions[0]/name`, `"sendKeys"`} {
		if !strings.Contains(msg, token) {
			t.Fatalf("CompileRuntimeBindingsFromProgram() error = %q, missing %q", msg, token)
		}
	}
}

func TestCompileRuntimeBindingsFromProgram_FlowReferenceBehavior(t *testing.T) {
	registry := actions.NewRegistry()
	t.Run("known flow", func(t *testing.T) {
		p := program.Program{
			Bindings: []program.BindingSpec{{ID: "paste", Hotkey: "shift+ctrl+v", Flow: "f1"}},
			Options:  program.Options{Flows: []program.FlowSpec{{ID: "f1", Steps: []program.FlowStepSpec{{Action: "system.log"}}}}},
		}
		bindings, err := CompileRuntimeBindingsFromProgram(p, registry)
		if err != nil {
			t.Fatalf("CompileRuntimeBindingsFromProgram() error = %v", err)
		}
		if len(bindings) != 1 || bindings[0].Flow == nil || len(bindings[0].Flow.Steps) != 1 {
			t.Fatalf("compiled flow binding unexpected: %#v", bindings)
		}
	})

	t.Run("unknown flow", func(t *testing.T) {
		p := program.Program{Bindings: []program.BindingSpec{{ID: "paste", Hotkey: "shift+ctrl+v", Flow: "missing"}}}
		_, err := CompileRuntimeBindingsFromProgram(p, registry)
		if err == nil || !strings.Contains(err.Error(), `binding "paste"`) {
			t.Fatalf("CompileRuntimeBindingsFromProgram() error = %v, want unknown flow with binding id", err)
		}
	})
}
