package window

import (
	"context"
	"errors"
	"fmt"
)

var ErrNoMatchingWindow = errors.New("no matching window")
var ErrAmbiguousWindow = errors.New("ambiguous window match")

// ActivationPolicy controls conflict handling when multiple windows match.
type ActivationPolicy struct {
	RequireSingleMatch bool
}

// ActivatorProvider provides the dependencies needed for activation.
type ActivatorProvider interface {
	EnumerateWindows(ctx context.Context) ([]Info, error)
	ActivateWindow(ctx context.Context, hwnd HWND) error
}

type ResolverProvider interface {
	EnumerateWindows(ctx context.Context) ([]Info, error)
}

// ActivateForeground selects the first matching window and activates it.
func ActivateForeground(ctx context.Context, provider ActivatorProvider, matcher Matcher) (Info, error) {
	return ActivateForegroundWithPolicy(ctx, provider, matcher, ActivationPolicy{})
}

// ResolveTargetWindow returns the window selected by matcher and policy.
func ResolveTargetWindow(ctx context.Context, provider ResolverProvider, matcher Matcher, policy ActivationPolicy) (Info, error) {
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
	if policy.RequireSingleMatch && len(matches) > 1 {
		return Info{}, fmt.Errorf("%w: %d windows matched", ErrAmbiguousWindow, len(matches))
	}
	return matches[0], nil
}

// ActivateForegroundWithPolicy resolves a target window and activates it.
func ActivateForegroundWithPolicy(ctx context.Context, provider ActivatorProvider, matcher Matcher, policy ActivationPolicy) (Info, error) {
	selected, err := ResolveTargetWindow(ctx, provider, matcher, policy)
	if err != nil {
		return Info{}, err
	}
	if err := provider.ActivateWindow(ctx, selected.HWND); err != nil {
		return Info{}, fmt.Errorf("activate window %s: %w", selected.HWND, err)
	}
	return selected, nil
}
