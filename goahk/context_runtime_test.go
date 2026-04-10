package goahk

import (
	"context"
	"errors"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/testutil"
)

func TestContextRuntime_StopAndSleep(t *testing.T) {
	t.Parallel()

	stopped := ""
	actionCtx := actions.ActionContext{Context: context.Background(), Stop: func(reason string) { stopped = reason }}
	ctx := newContext(&actionCtx, newAppState())
	ctx.Runtime.Stop()
	if stopped != "runtime.stop" {
		t.Fatalf("stop reason = %q, want runtime.stop", stopped)
	}

	start := time.Now()
	if !ctx.Runtime.Sleep(5 * time.Millisecond) {
		t.Fatalf("Sleep() = false, want true")
	}
	if time.Since(start) < 4*time.Millisecond {
		t.Fatalf("Sleep() returned too quickly")
	}
}

func TestContext_ErrReflectsCancellation(t *testing.T) {
	t.Parallel()

	base, cancel := context.WithCancel(context.Background())
	actionCtx := actions.ActionContext{Context: base}
	ctx := newContext(&actionCtx, newAppState())

	if got := ctx.Err(); got != nil {
		t.Fatalf("Err() before cancel = %v, want nil", got)
	}

	cancel()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("Err() after cancel = %v, want context.Canceled", ctx.Err())
	}
}

func TestContext_SleepReturnsFalseWhenCanceledMidSleep(t *testing.T) {
	t.Parallel()

	base, cancel := context.WithCancel(context.Background())
	actionCtx := actions.ActionContext{Context: base}
	ctx := newContext(&actionCtx, newAppState())

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	if got := ctx.Sleep(50 * time.Millisecond); got {
		t.Fatalf("Sleep() = true, want false on cancel")
	}
	if time.Since(start) >= 45*time.Millisecond {
		t.Fatalf("Sleep() did not return promptly after cancellation")
	}
}

func TestContext_LogRoutesToConfiguredLoggerAndNoopsWithoutLogger(t *testing.T) {
	t.Parallel()

	fake := &testutil.FakeLogger{}
	actionCtx := actions.ActionContext{Context: context.Background(), Logger: fake}
	ctx := newContext(&actionCtx, newAppState())
	ctx.Log("hello", "binding", "Ctrl+H", "attempt", 1)

	if len(fake.Entries) != 1 {
		t.Fatalf("entries len = %d, want 1", len(fake.Entries))
	}
	if fake.Entries[0].Msg != "hello" {
		t.Fatalf("entry msg = %q, want hello", fake.Entries[0].Msg)
	}
	if fake.Entries[0].Fields["binding"] != "Ctrl+H" {
		t.Fatalf("binding field = %#v, want Ctrl+H", fake.Entries[0].Fields["binding"])
	}

	// No logger configured: should not panic.
	noLogger := newContext(&actions.ActionContext{Context: context.Background()}, newAppState())
	noLogger.Log("noop")
}

func TestContextRuntime_DefaultLoggerIsNoopWhenOptionOmitted(t *testing.T) {
	t.Parallel()

	a := NewApp()
	logger := a.actionLogger()
	logger.Info("noop", map[string]any{"k": "v"})
	logger.Error("noop", map[string]any{"k": "v"})
}
