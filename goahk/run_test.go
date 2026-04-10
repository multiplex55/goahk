package goahk

import (
	"context"
	"errors"
	stdruntime "runtime"
	"strings"
	"testing"
)

func TestRunSurfacesCompileFailuresWithActionableError(t *testing.T) {
	a := NewApp()
	a.On("Ctrl+Nope").Do(Log("x"))

	err := a.Run(context.Background())
	if err == nil {
		t.Fatal("Run() error = nil, want failure")
	}
	msg := err.Error()
	for _, token := range []string{"compile app program", "binding_1", "unsupported key"} {
		if !strings.Contains(msg, token) {
			t.Fatalf("Run() error = %q, missing %q", msg, token)
		}
	}
}

func TestRunReturnsExplicitUnsupportedPlatformErrorOnNonWindows(t *testing.T) {
	if stdruntime.GOOS == "windows" {
		t.Skip("non-windows only")
	}
	a := NewApp()
	a.On("Ctrl+H").Do(Log("x"))

	err := a.Run(context.Background())
	if err == nil {
		t.Fatal("Run() error = nil, want unsupported platform error")
	}
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("Run() error = %v, want ErrUnsupportedPlatform", err)
	}
	if !strings.Contains(err.Error(), "require=windows") {
		t.Fatalf("Run() error = %q, want explicit windows requirement", err.Error())
	}
}
