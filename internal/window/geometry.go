package window

import (
	"context"
	"fmt"
)

// GeometryProvider exposes OS-backed window geometry operations.
type GeometryProvider interface {
	EnumerateWindows(context.Context) ([]Info, error)
	WindowBounds(context.Context, HWND) (Rect, error)
	MoveWindow(context.Context, HWND, int, int) error
	ResizeWindow(context.Context, HWND, int, int) error
}

func BoundsForMatcher(ctx context.Context, provider GeometryProvider, matcher Matcher, policy ActivationPolicy) (Info, Rect, error) {
	target, err := ResolveTargetWindow(ctx, provider, matcher, policy)
	if err != nil {
		return Info{}, Rect{}, err
	}
	bounds, err := provider.WindowBounds(ctx, target.HWND)
	if err != nil {
		return Info{}, Rect{}, fmt.Errorf("bounds for window %s: %w", target.HWND, err)
	}
	return target, bounds, nil
}

func MoveForMatcher(ctx context.Context, provider GeometryProvider, matcher Matcher, x, y int, policy ActivationPolicy) (Info, error) {
	target, err := ResolveTargetWindow(ctx, provider, matcher, policy)
	if err != nil {
		return Info{}, err
	}
	if err := provider.MoveWindow(ctx, target.HWND, x, y); err != nil {
		return Info{}, fmt.Errorf("move window %s: %w", target.HWND, err)
	}
	return target, nil
}

func ResizeForMatcher(ctx context.Context, provider GeometryProvider, matcher Matcher, width, height int, policy ActivationPolicy) (Info, error) {
	target, err := ResolveTargetWindow(ctx, provider, matcher, policy)
	if err != nil {
		return Info{}, err
	}
	if err := provider.ResizeWindow(ctx, target.HWND, width, height); err != nil {
		return Info{}, fmt.Errorf("resize window %s: %w", target.HWND, err)
	}
	return target, nil
}
