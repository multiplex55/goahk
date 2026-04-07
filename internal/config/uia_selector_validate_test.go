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
	sel, err := ParseUIASelector(map[string]string{"selector": "login"}, map[string]UIASelector{
		"login": {AutomationID: "username", Ancestors: []UIASelector{{Name: "Form"}}},
	})
	if err != nil {
		t.Fatalf("ParseUIASelector() error = %v", err)
	}
	if sel.AutomationID != "username" || len(sel.Ancestors) != 1 {
		t.Fatalf("selector=%+v", sel)
	}
}
