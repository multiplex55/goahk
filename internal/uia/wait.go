package uia

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type RetryPolicy struct {
	Interval    time.Duration
	MaxAttempts int
}

type Clock interface {
	Now() time.Time
	Sleep(context.Context, time.Duration) error
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }
func (realClock) Sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

type WaitResult struct {
	Element    Element
	Attempts   int
	RetryCount int
	Timeout    time.Duration
}

func WaitUntilExists(ctx context.Context, nav Navigator, rootID string, sel Selector, timeout time.Duration, policy RetryPolicy, clk Clock) (WaitResult, error) {
	if clk == nil {
		clk = realClock{}
	}
	if policy.Interval <= 0 {
		policy.Interval = 50 * time.Millisecond
	}
	start := clk.Now()
	deadline := start.Add(timeout)
	attempt := 0
	for {
		attempt++
		el, _, err := Find(ctx, nav, rootID, sel)
		if err == nil {
			return WaitResult{Element: el, Attempts: attempt, RetryCount: attempt - 1, Timeout: timeout}, nil
		}
		if !errors.Is(err, ErrElementNotFound) {
			return WaitResult{Attempts: attempt, RetryCount: attempt - 1, Timeout: timeout}, err
		}
		if policy.MaxAttempts > 0 && attempt >= policy.MaxAttempts {
			return WaitResult{Attempts: attempt, RetryCount: attempt - 1, Timeout: timeout}, fmt.Errorf("uia wait max attempts reached: %w", err)
		}
		if timeout > 0 && !clk.Now().Before(deadline) {
			return WaitResult{Attempts: attempt, RetryCount: attempt - 1, Timeout: timeout}, fmt.Errorf("uia wait timeout after %s: %w", timeout, err)
		}
		if err := clk.Sleep(ctx, policy.Interval); err != nil {
			return WaitResult{Attempts: attempt, RetryCount: attempt - 1, Timeout: timeout}, err
		}
	}
}
