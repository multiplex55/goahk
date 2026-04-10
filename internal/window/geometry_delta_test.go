package window

import "testing"

func TestTranslateMoveBy(t *testing.T) {
	x, y := TranslateMoveBy(Rect{Left: 120, Top: 80, Right: 620, Bottom: 480}, -20, 15)
	if x != 100 || y != 95 {
		t.Fatalf("TranslateMoveBy() = (%d,%d), want (100,95)", x, y)
	}
}

func TestTranslateResizeBy(t *testing.T) {
	w, h, err := TranslateResizeBy(Rect{Left: 0, Top: 0, Right: 300, Bottom: 200}, 50, -20)
	if err != nil {
		t.Fatalf("TranslateResizeBy() error = %v", err)
	}
	if w != 350 || h != 180 {
		t.Fatalf("TranslateResizeBy() = (%d,%d), want (350,180)", w, h)
	}
}

func TestTranslateResizeBy_RejectsNonPositiveSize(t *testing.T) {
	_, _, err := TranslateResizeBy(Rect{Left: 0, Top: 0, Right: 10, Bottom: 10}, -10, -5)
	if err == nil {
		t.Fatal("TranslateResizeBy() error = nil, want error")
	}
}

func TestCenterPosition(t *testing.T) {
	x, y := CenterPosition(
		Rect{Left: 100, Top: 100, Right: 500, Bottom: 300},
		Rect{Left: 0, Top: 0, Right: 1920, Bottom: 1040},
	)
	if x != 760 || y != 420 {
		t.Fatalf("CenterPosition() = (%d,%d), want (760,420)", x, y)
	}
}
