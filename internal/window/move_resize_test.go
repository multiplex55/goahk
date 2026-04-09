package window

import (
	"context"
	"testing"
)

func TestMoveAndResizeActive_DelegatesToProvider(t *testing.T) {
	p := &fakeGeometryProvider{
		active: Info{HWND: 15, Title: "Editor"},
	}

	if _, err := MoveActive(context.Background(), p, 40, 50); err != nil {
		t.Fatalf("MoveActive() error = %v", err)
	}
	if p.moved != 15 || p.movePosition != [2]int{40, 50} {
		t.Fatalf("move call mismatch: hwnd=%v pos=%v", p.moved, p.movePosition)
	}

	if _, err := ResizeActive(context.Background(), p, 1000, 720); err != nil {
		t.Fatalf("ResizeActive() error = %v", err)
	}
	if p.resized != 15 || p.resizeSize != [2]int{1000, 720} {
		t.Fatalf("resize call mismatch: hwnd=%v size=%v", p.resized, p.resizeSize)
	}
}
