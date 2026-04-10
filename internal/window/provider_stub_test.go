//go:build !windows
// +build !windows

package window

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestOSProviderUnsupportedOperationsReturnActionableErrors(t *testing.T) {
	p := NewOSProvider()
	_, err := p.WindowBounds(context.Background(), 1)
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("WindowBounds error = %v, want ErrUnsupportedPlatform", err)
	}
	if !strings.Contains(err.Error(), "requires Windows") {
		t.Fatalf("WindowBounds error = %q, want actionable guidance", err.Error())
	}

	if err := p.MinimizeWindow(context.Background(), 1); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("MinimizeWindow error = %v, want ErrUnsupportedPlatform", err)
	}

	if _, err := p.WorkAreaForWindow(context.Background(), 1); !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("WorkAreaForWindow error = %v, want ErrUnsupportedPlatform", err)
	}
}
