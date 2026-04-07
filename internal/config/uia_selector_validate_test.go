package config

import (
	"strings"
	"testing"
)

func TestValidate_UIASelectorSchema(t *testing.T) {
	cfg := Config{
		UIASelectors: map[string]UIASelector{
			"bad": {},
		},
	}
	err := Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "must specify at least one") {
		t.Fatalf("error=%v", err)
	}
}

func TestValidate_UIASelectorReferenceResolution(t *testing.T) {
	cfg := Config{
		UIASelectors: map[string]UIASelector{
			"ok": {AutomationID: "submit"},
		},
		Hotkeys: []HotkeyBinding{
			{
				ID:     "hk",
				Hotkey: "Ctrl+1",
				Steps: []Step{
					{Action: "uia.find", Params: map[string]string{"selector": "missing"}},
				},
			},
		},
	}
	err := Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "unknown uia selector") {
		t.Fatalf("error=%v", err)
	}
}

func TestParseUIASelector_Reference(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]string
		defs    map[string]UIASelector
		wantID  string
		wantErr string
	}{
		{
			name:   "reference",
			params: map[string]string{"selector": "login"},
			defs: map[string]UIASelector{
				"login": {AutomationID: "username", Ancestors: []UIASelector{{Name: "Form"}}},
			},
			wantID: "username",
		},
		{
			name:    "missing reference",
			params:  map[string]string{"selector": "missing"},
			defs:    map[string]UIASelector{"ok": {Name: "x"}},
			wantErr: "unknown uia selector",
		},
		{
			name:    "missing selector content",
			params:  map[string]string{},
			defs:    map[string]UIASelector{},
			wantErr: "selector requires automationId, name, or controlType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := ParseUIASelector(tt.params, tt.defs)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err=%v want contains %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseUIASelector() error = %v", err)
			}
			if sel.AutomationID != tt.wantID || len(sel.Ancestors) != 1 {
				t.Fatalf("selector=%+v", sel)
			}
		})
	}
}
