//go:build windows

package window

import (
	"context"
	"testing"
)

func TestWindowsProvider_CanMatchAndActivateKnownTarget(t *testing.T) {
	ctx := context.Background()
	provider := NewOSProvider()

	active, err := provider.ActiveWindow(ctx)
	if err != nil {
		t.Fatalf("ActiveWindow returned error: %v", err)
	}
	if active.HWND == 0 {
		t.Fatal("expected non-zero active HWND")
	}
	if active.PID == 0 {
		t.Fatal("expected active window PID to be populated")
	}

	matcher := Matcher{
		TitleExact: active.Title,
		ClassName:  active.Class,
		ExeName:    active.Exe,
	}
	selected, err := ActivateForegroundWithPolicy(ctx, provider, matcher, ActivationPolicy{RequireSingleMatch: true})
	if err != nil {
		t.Fatalf("ActivateForegroundWithPolicy returned error: %v", err)
	}
	if selected.HWND != active.HWND {
		t.Fatalf("selected HWND=%s want %s", selected.HWND, active.HWND)
	}
}
