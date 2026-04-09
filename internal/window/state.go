package window

import (
	"context"
	"fmt"
)

type StateProvider interface {
	EnumerateWindows(context.Context) ([]Info, error)
	MinimizeWindow(context.Context, HWND) error
	MaximizeWindow(context.Context, HWND) error
	RestoreWindow(context.Context, HWND) error
}

type ActiveStateProvider interface {
	ActiveWindow(context.Context) (Info, error)
	MinimizeWindow(context.Context, HWND) error
	MaximizeWindow(context.Context, HWND) error
	RestoreWindow(context.Context, HWND) error
}

func MinimizeForMatcher(ctx context.Context, provider StateProvider, matcher Matcher, policy ActivationPolicy) (Info, error) {
	return applyStateForMatcher(ctx, provider, matcher, policy, "minimize", provider.MinimizeWindow)
}

func MaximizeForMatcher(ctx context.Context, provider StateProvider, matcher Matcher, policy ActivationPolicy) (Info, error) {
	return applyStateForMatcher(ctx, provider, matcher, policy, "maximize", provider.MaximizeWindow)
}

func RestoreForMatcher(ctx context.Context, provider StateProvider, matcher Matcher, policy ActivationPolicy) (Info, error) {
	return applyStateForMatcher(ctx, provider, matcher, policy, "restore", provider.RestoreWindow)
}

func MinimizeActive(ctx context.Context, provider ActiveStateProvider) (Info, error) {
	return applyStateToActive(ctx, provider, "minimize", provider.MinimizeWindow)
}

func MaximizeActive(ctx context.Context, provider ActiveStateProvider) (Info, error) {
	return applyStateToActive(ctx, provider, "maximize", provider.MaximizeWindow)
}

func RestoreActive(ctx context.Context, provider ActiveStateProvider) (Info, error) {
	return applyStateToActive(ctx, provider, "restore", provider.RestoreWindow)
}

func applyStateForMatcher(ctx context.Context, provider StateProvider, matcher Matcher, policy ActivationPolicy, op string, fn func(context.Context, HWND) error) (Info, error) {
	target, err := ResolveTargetWindow(ctx, provider, matcher, policy)
	if err != nil {
		return Info{}, err
	}
	if err := fn(ctx, target.HWND); err != nil {
		return Info{}, fmt.Errorf("%s window %s: %w", op, target.HWND, err)
	}
	return target, nil
}

func applyStateToActive(ctx context.Context, provider ActiveStateProvider, op string, fn func(context.Context, HWND) error) (Info, error) {
	active, err := provider.ActiveWindow(ctx)
	if err != nil {
		return Info{}, err
	}
	if err := fn(ctx, active.HWND); err != nil {
		return Info{}, fmt.Errorf("%s active window %s: %w", op, active.HWND, err)
	}
	return active, nil
}
