//go:build windows
// +build windows

package inspect

import "testing"

func TestOverlayWindowStyles(t *testing.T) {
	style := overlayWindowStyle()
	if style&wsPopup == 0 {
		t.Fatalf("expected popup style bit, got 0x%x", style)
	}

	exStyle := overlayWindowExStyle()
	if exStyle&wsExTopMost == 0 {
		t.Fatalf("expected topmost ex-style bit, got 0x%x", exStyle)
	}
	if exStyle&wsExLayered == 0 {
		t.Fatalf("expected layered ex-style bit, got 0x%x", exStyle)
	}
	if exStyle&wsExTransparent == 0 {
		t.Fatalf("expected transparent/click-through ex-style bit, got 0x%x", exStyle)
	}
}

func TestOverlayPaintBorderOnly(t *testing.T) {
	if !overlayPaintUsesBorderOnly() {
		t.Fatalf("expected border-only overlay paint path")
	}
}
