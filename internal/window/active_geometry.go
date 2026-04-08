package window

import (
	"context"
	"fmt"
)

type ActiveGeometryProvider interface {
	ActiveWindow(context.Context) (Info, error)
	WindowBounds(context.Context, HWND) (Rect, error)
	MoveWindow(context.Context, HWND, int, int) error
	ResizeWindow(context.Context, HWND, int, int) error
}

func ActiveBounds(ctx context.Context, provider ActiveGeometryProvider) (Info, Rect, error) {
	active, err := provider.ActiveWindow(ctx)
	if err != nil {
		return Info{}, Rect{}, err
	}
	bounds, err := provider.WindowBounds(ctx, active.HWND)
	if err != nil {
		return Info{}, Rect{}, fmt.Errorf("bounds for active window %s: %w", active.HWND, err)
	}
	return active, bounds, nil
}

func MoveActive(ctx context.Context, provider ActiveGeometryProvider, x, y int) (Info, error) {
	active, err := provider.ActiveWindow(ctx)
	if err != nil {
		return Info{}, err
	}
	if err := provider.MoveWindow(ctx, active.HWND, x, y); err != nil {
		return Info{}, fmt.Errorf("move active window %s: %w", active.HWND, err)
	}
	return active, nil
}

func ResizeActive(ctx context.Context, provider ActiveGeometryProvider, width, height int) (Info, error) {
	active, err := provider.ActiveWindow(ctx)
	if err != nil {
		return Info{}, err
	}
	if err := provider.ResizeWindow(ctx, active.HWND, width, height); err != nil {
		return Info{}, fmt.Errorf("resize active window %s: %w", active.HWND, err)
	}
	return active, nil
}
