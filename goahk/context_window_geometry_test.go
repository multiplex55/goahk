package goahk

import (
	"context"
	"testing"

	"goahk/internal/actions"
	"goahk/internal/window"
)

func TestContextWindow_GeometryUsesServices(t *testing.T) {
	t.Parallel()

	var moved, resized window.HWND
	ctx := newContext(&actions.ActionContext{
		Context: context.Background(),
		Services: actions.Services{
			WindowActive: func(context.Context) (window.Info, error) { return window.Info{HWND: 77, Title: "Active"}, nil },
			WindowBounds: func(context.Context, window.HWND) (window.Rect, error) {
				return window.Rect{Left: 1, Top: 2, Right: 301, Bottom: 202}, nil
			},
			WindowMove: func(_ context.Context, hwnd window.HWND, x, y int) error {
				moved = hwnd
				_, _ = x, y
				return nil
			},
			WindowResize: func(_ context.Context, hwnd window.HWND, width, height int) error {
				resized = hwnd
				_, _ = width, height
				return nil
			},
		},
	}, newAppState())

	active, err := ctx.Window.Active()
	if err != nil {
		t.Fatalf("Active() error = %v", err)
	}
	if active.HWND != 77 {
		t.Fatalf("active hwnd = %v, want 77", active.HWND)
	}
	bounds, err := ctx.Window.Bounds(active.HWND)
	if err != nil {
		t.Fatalf("Bounds() error = %v", err)
	}
	if bounds.Width() != 300 || bounds.Height() != 200 {
		t.Fatalf("unexpected bounds = %#v", bounds)
	}
	if err := ctx.Window.Move(active.HWND, 25, 30); err != nil {
		t.Fatalf("Move() error = %v", err)
	}
	if err := ctx.Window.Resize(active.HWND, 900, 700); err != nil {
		t.Fatalf("Resize() error = %v", err)
	}
	if moved != 77 || resized != 77 {
		t.Fatalf("move/resize hwnd mismatch: moved=%v resized=%v", moved, resized)
	}
}
