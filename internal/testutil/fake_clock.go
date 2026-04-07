package testutil

import (
	"context"
	"sync"
	"time"
)

type FakeClock struct {
	mu    sync.Mutex
	now   time.Time
	Slept []time.Duration
}

func NewFakeClock(start time.Time) *FakeClock {
	return &FakeClock{now: start}
}

func (f *FakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *FakeClock) Sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Slept = append(f.Slept, d)
	f.now = f.now.Add(d)
	return nil
}
