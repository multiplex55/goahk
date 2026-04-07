package uia

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type fakeClock struct {
	now    time.Time
	sleeps []time.Duration
}

func (f *fakeClock) Now() time.Time { return f.now }
func (f *fakeClock) Sleep(_ context.Context, d time.Duration) error {
	f.sleeps = append(f.sleeps, d)
	f.now = f.now.Add(d)
	return nil
}

type flappingNavigator struct {
	attempt int
	readyAt int
}

func (f *flappingNavigator) ElementByID(_ context.Context, id string) (Element, error) {
	if id == "root" {
		return Element{ID: "root"}, nil
	}
	if id == "target" {
		return Element{ID: "target", Name: strPtr("Ready")}, nil
	}
	return Element{}, fmt.Errorf("unknown id")
}
func (f *flappingNavigator) ChildrenIDs(_ context.Context, id string) ([]string, error) {
	if id != "root" {
		return nil, nil
	}
	f.attempt++
	if f.attempt >= f.readyAt {
		return []string{"target"}, nil
	}
	return nil, nil
}

func TestWaitUntilExists_RetriesUntilPresent(t *testing.T) {
	clk := &fakeClock{now: time.Unix(0, 0)}
	nav := &flappingNavigator{readyAt: 3}
	got, err := WaitUntilExists(context.Background(), nav, "root", Selector{Name: "Ready"}, 2*time.Second, RetryPolicy{Interval: 10 * time.Millisecond}, clk)
	if err != nil {
		t.Fatalf("WaitUntilExists() error = %v", err)
	}
	if got.Attempts != 3 || got.RetryCount != 2 {
		t.Fatalf("attempts=%d retries=%d", got.Attempts, got.RetryCount)
	}
}

func TestWaitUntilExists_TimesOut(t *testing.T) {
	clk := &fakeClock{now: time.Unix(0, 0)}
	nav := &flappingNavigator{readyAt: 100}
	_, err := WaitUntilExists(context.Background(), nav, "root", Selector{Name: "Ready"}, 25*time.Millisecond, RetryPolicy{Interval: 10 * time.Millisecond}, clk)
	if err == nil || !errors.Is(err, ErrElementNotFound) {
		t.Fatalf("error=%v", err)
	}
}
