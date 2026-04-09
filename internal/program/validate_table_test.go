package program

import (
	"errors"
	"testing"
)

func TestValidateTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		program   Program
		wantCodes []string
	}{
		{
			name: "valid normalized steps",
			program: Program{Bindings: []BindingSpec{{
				ID:     " paste ",
				Hotkey: " shift + control + v ",
				Steps:  []StepSpec{{Action: "system.log"}},
			}}},
		},
		{
			name: "duplicate id",
			program: Program{Bindings: []BindingSpec{
				{ID: "copy", Hotkey: "ctrl+c", Steps: []StepSpec{{Action: "system.log"}}},
				{ID: " COPY ", Hotkey: "ctrl+shift+c", Steps: []StepSpec{{Action: "system.log"}}},
			}},
			wantCodes: []string{ErrCodeDuplicateBindingID},
		},
		{
			name:      "invalid hotkey",
			program:   Program{Bindings: []BindingSpec{{ID: "copy", Hotkey: "ctrl+badkey", Steps: []StepSpec{{Action: "system.log"}}}}},
			wantCodes: []string{ErrCodeInvalidHotkey},
		},
		{
			name:      "empty steps without flow",
			program:   Program{Bindings: []BindingSpec{{ID: "copy", Hotkey: "ctrl+c"}}},
			wantCodes: []string{ErrCodeStepsRequired},
		},
		{
			name:      "empty action",
			program:   Program{Bindings: []BindingSpec{{ID: "copy", Hotkey: "ctrl+c", Steps: []StepSpec{{Action: " "}}}}},
			wantCodes: []string{ErrCodeStepActionRequired},
		},
		{
			name:      "unknown flow",
			program:   Program{Bindings: []BindingSpec{{ID: "copy", Hotkey: "ctrl+c", Flow: "missing"}}, Options: Options{Flows: []FlowSpec{{ID: "known", Steps: []FlowStepSpec{{Action: "system.log"}}}}}},
			wantCodes: []string{ErrCodeUnknownFlow},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.program)
			if len(tt.wantCodes) == 0 {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() error = nil, want codes %v", tt.wantCodes)
			}
			var verr *ValidationError
			if !errors.As(err, &verr) {
				t.Fatalf("Validate() error = %T, want *ValidationError", err)
			}
			for _, code := range tt.wantCodes {
				if !verr.HasCode(code) {
					t.Fatalf("Validate() missing code %q in %#v", code, verr.Issues)
				}
			}
		})
	}
}
