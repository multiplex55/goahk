package actions

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCallbackContext_SleepExitsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ac := ActionContext{Context: ctx}
	cb := NewCallbackContext(&ac)
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	completed := cb.Sleep(time.Second)
	if completed {
		t.Fatal("Sleep should return false when canceled")
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("Sleep canceled too late: %s", elapsed)
	}
}

func TestCallbackContext_DoneAndErrSemantics(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	ac := ActionContext{Context: ctx}
	cb := NewCallbackContext(&ac)
	if cb.IsCancelled() {
		t.Fatal("context should not be canceled yet")
	}
	cause := errors.New("stop requested")
	cancel(cause)
	<-cb.Done()
	if !cb.IsCancelled() {
		t.Fatal("context should be canceled")
	}
	if !errors.Is(cb.Err(), context.Canceled) {
		t.Fatalf("Err() = %v, want context.Canceled", cb.Err())
	}
}
