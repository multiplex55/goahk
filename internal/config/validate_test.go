package config

import (
	"strings"
	"testing"
)

func TestValidate_RequiredFields(t *testing.T) {
	cfg := Config{
		Hotkeys: []HotkeyBinding{{}},
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate() error = nil, want required field errors")
	}
	for _, token := range []string{"id is required", "hotkey is required", "steps is required"} {
		if !strings.Contains(err.Error(), token) {
			t.Fatalf("Validate() error = %q, missing %q", err, token)
		}
	}
}

func TestValidate_DuplicateIDs(t *testing.T) {
	cfg := Config{Hotkeys: []HotkeyBinding{
		{ID: "x", Hotkey: "Ctrl+1", Steps: []Step{{Action: "noop"}}},
		{ID: "X", Hotkey: "Ctrl+2", Steps: []Step{{Action: "noop"}}},
	}}

	err := Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "id duplicates") {
		t.Fatalf("Validate() error = %v, want duplicate id error", err)
	}
}

func TestValidate_DuplicateHotkeys(t *testing.T) {
	cfg := Config{Hotkeys: []HotkeyBinding{
		{ID: "a", Hotkey: "Ctrl+Alt+T", Steps: []Step{{Action: "noop"}}},
		{ID: "b", Hotkey: "ctrl + alt + t", Steps: []Step{{Action: "noop"}}},
	}}

	err := Validate(cfg)
	if err == nil || !strings.Contains(err.Error(), "hotkey duplicates") {
		t.Fatalf("Validate() error = %v, want duplicate hotkey error", err)
	}
}

func TestApplyDefaults(t *testing.T) {
	cfg := Config{}
	ApplyDefaults(&cfg)

	if cfg.App.Name == "" || cfg.Logging.Level == "" || cfg.Clipboard.HistorySize == 0 {
		t.Fatalf("defaults were not injected: %+v", cfg)
	}
}
