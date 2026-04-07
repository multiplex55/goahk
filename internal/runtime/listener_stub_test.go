//go:build !windows
// +build !windows

package runtime

import (
	"context"
	"strings"
	"testing"
)

func TestNewWindowsListenerOnNonWindowsReturnsExplicitError(t *testing.T) {
	listener, err := NewWindowsListener(context.Background())
	if err == nil {
		t.Fatal("expected an unsupported-platform error")
	}
	if listener != nil {
		t.Fatal("expected nil listener")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "only available on windows") {
		t.Fatalf("expected error to mention windows-only support, got %q", err)
	}
}
