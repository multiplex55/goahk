package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFile_ValidMinimal(t *testing.T) {
	cfg, err := LoadFile(filepath.Join("..", "..", "testdata", "config", "valid_minimal.json"))
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if got, want := len(cfg.Hotkeys), 1; got != want {
		t.Fatalf("len(cfg.Hotkeys) = %d, want %d", got, want)
	}
	if cfg.Logging.Level != "info" {
		t.Fatalf("default logging level not applied, got %q", cfg.Logging.Level)
	}
}

func TestLoadFile_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "broken.json")
	if err := os.WriteFile(path, []byte(`{"hotkeys": [`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("LoadFile() error = nil, want decode error")
	}
}

func TestLoadFile_UnknownFieldsRejected(t *testing.T) {
	_, err := LoadFile(filepath.Join("..", "..", "testdata", "config", "malformed_schema.json"))
	if err == nil {
		t.Fatal("LoadFile() error = nil, want unknown field error")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("LoadFile() error = %q, want unknown field", err)
	}
}
