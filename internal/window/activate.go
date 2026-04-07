package window

import (
	"context"
	"errors"
	"fmt"
)

var ErrNoMatchingWindow = errors.New("no matching window")

// ActivatorProvider provides the dependencies needed for activation.
type ActivatorProvider interface {
	EnumerateWindows(ctx context.Context) ([]Info, error)
	ActivateWindow(ctx context.Context, hwnd HWND) error
}

// ActivateForeground selects the first matching window and activates it.
func ActivateForeground(ctx context.Context, provider ActivatorProvider, matcher Matcher) (Info, error) {
	windows, err := provider.EnumerateWindows(ctx)
	if err != nil {
		return Info{}, fmt.Errorf("enumerate windows: %w", err)
	}
	matches, err := Filter(windows, matcher)
	if err != nil {
		return Info{}, err
	}
	if len(matches) == 0 {
		return Info{}, ErrNoMatchingWindow
	}
	selected := matches[0]
	if err := provider.ActivateWindow(ctx, selected.HWND); err != nil {
		return Info{}, fmt.Errorf("activate window %s: %w", selected.HWND, err)
	}
	return selected, nil
}
