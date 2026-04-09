package program

import (
	"strings"
	"testing"
)

func TestValidateDuplicateBindingID(t *testing.T) {
	p := Program{Bindings: []BindingSpec{{ID: "copy", Hotkey: "ctrl+c"}, {ID: "copy", Hotkey: "ctrl+shift+c"}}}
	err := Validate(p)
	if err == nil || !strings.Contains(err.Error(), "duplicate binding id") {
		t.Fatalf("Validate() error = %v, want duplicate id failure", err)
	}
}

func TestValidateDuplicateHotkey(t *testing.T) {
	p := Program{Bindings: []BindingSpec{{ID: "copy", Hotkey: "ctrl+c"}, {ID: "copy2", Hotkey: "ctrl+c"}}}
	err := Validate(p)
	if err == nil || !strings.Contains(err.Error(), "hotkey conflict") {
		t.Fatalf("Validate() error = %v, want hotkey conflict", err)
	}
}

func TestValidateInvalidHotkey(t *testing.T) {
	p := Program{Bindings: []BindingSpec{{ID: "copy", Hotkey: "ctrl+badkey"}}}
	err := Validate(p)
	if err == nil || !strings.Contains(err.Error(), `binding "copy"`) {
		t.Fatalf("Validate() error = %v, want binding id in parse error", err)
	}
}

func TestValidateUnknownFlowReference(t *testing.T) {
	p := Program{Bindings: []BindingSpec{{ID: "run", Hotkey: "ctrl+r", Flow: "missing"}}}
	err := Validate(p)
	if err == nil || !strings.Contains(err.Error(), `unknown flow`) {
		t.Fatalf("Validate() error = %v, want unknown flow", err)
	}
}

func TestValidateEmptyActionName(t *testing.T) {
	p := Program{Bindings: []BindingSpec{{
		ID:     "run",
		Hotkey: "ctrl+r",
		Steps:  []StepSpec{{Action: ""}},
	}}}
	err := Validate(p)
	if err == nil {
		t.Fatal("Validate() error = nil, want failure")
	}
	msg := err.Error()
	for _, token := range []string{ErrCodeStepActionRequired, `bindings[0].steps[0].action`} {
		if !strings.Contains(msg, token) {
			t.Fatalf("Validate() error = %q, missing %q", msg, token)
		}
	}
}
