package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidate_ConfigErrorsGolden(t *testing.T) {
	cfg := Config{Hotkeys: []HotkeyBinding{{}}}
	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error")
	}
	got := err.Error() + "\n"
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	path := filepath.Join(root, "testdata", "golden", "config", "validation_errors.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if string(want) != got {
		t.Fatalf("golden mismatch\nwant:\n%s\ngot:\n%s", string(want), got)
	}
}
