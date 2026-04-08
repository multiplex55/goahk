package goahk

import (
	"context"
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
