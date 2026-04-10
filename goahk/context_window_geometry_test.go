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

func TestContextWindow_DeltaAndCenterHelpersTranslateArguments(t *testing.T) {
	t.Parallel()

	var movedX, movedY, resizedW, resizedH int
	ctx := newContext(&actions.ActionContext{
		Context: context.Background(),
		Services: actions.Services{
			WindowBounds: func(context.Context, window.HWND) (window.Rect, error) {
				return window.Rect{Left: 100, Top: 200, Right: 500, Bottom: 450}, nil
			},
			WindowMove: func(_ context.Context, _ window.HWND, x, y int) error {
				movedX, movedY = x, y
				return nil
			},
			WindowResize: func(_ context.Context, _ window.HWND, width, height int) error {
				resizedW, resizedH = width, height
				return nil
			},
			WindowWorkArea: func(context.Context, window.HWND) (window.Rect, error) {
				return window.Rect{Left: 0, Top: 0, Right: 1920, Bottom: 1040}, nil
			},
		},
	}, newAppState())

	if err := ctx.Window.MoveBy(77, 10, -50); err != nil {
		t.Fatalf("MoveBy() error = %v", err)
	}
	if movedX != 110 || movedY != 150 {
		t.Fatalf("MoveBy translated position = (%d,%d), want (110,150)", movedX, movedY)
	}

	if err := ctx.Window.ResizeBy(77, 20, -10); err != nil {
		t.Fatalf("ResizeBy() error = %v", err)
	}
	if resizedW != 420 || resizedH != 240 {
		t.Fatalf("ResizeBy translated size = (%d,%d), want (420,240)", resizedW, resizedH)
	}

	if err := ctx.Window.Center(77); err != nil {
		t.Fatalf("Center() error = %v", err)
	}
	if movedX != 760 || movedY != 395 {
		t.Fatalf("Center translated position = (%d,%d), want (760,395)", movedX, movedY)
	}
}
