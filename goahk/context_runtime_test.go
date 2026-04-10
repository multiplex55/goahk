package goahk

import (
	"context"
	"testing"
	"time"

	"goahk/internal/actions"
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
	ctx.Runtime.Sleep(5 * time.Millisecond)
	if time.Since(start) < 4*time.Millisecond {
		t.Fatalf("Sleep() returned too quickly")
	}
}

func TestContextRuntime_DefaultLoggerIsNoopWhenOptionOmitted(t *testing.T) {
	t.Parallel()

	a := NewApp()
	logger := a.actionLogger()
	logger.Info("noop", map[string]any{"k": "v"})
	logger.Error("noop", map[string]any{"k": "v"})
}
