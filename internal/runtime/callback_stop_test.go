package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/config"
	"goahk/internal/program"
)

func TestRuntime_CallbackCancellationAndGracePolicy(t *testing.T) {
	listener := newFakeListener()
	cfg := config.Config{Hotkeys: []config.HotkeyBinding{{
		ID:     "cb",
		Hotkey: "ctrl+1",
		Steps:  []config.Step{{Action: actions.CallbackActionName, Params: map[string]string{"callback_ref": "slow"}}},
	}}}
	var canceledSeen atomic.Bool

	b := NewBootstrap()
	b.StopGrace = 80 * time.Millisecond
	b.HardStopAfterGrace = true
	b.LoadProgram = func(context.Context, string) (program.Program, error) { return mustProgram(t, cfg), nil }
	b.NewListener = func(context.Context) (Listener, error) { return listener, nil }
	b.BuildRegistry = func(context.Context, program.Program) (*actions.Registry, error) {
		reg := actions.NewRegistry()
		reg.MustRegisterCallback("slow", func(ctx actions.CallbackContext) error {
			<-ctx.Done()
			if ctx.IsCancelled() {
				canceledSeen.Store(true)
			}
			return ctx.Err()
		})
		return reg, nil
	}
	listener.runFn = func(ctx context.Context) error {
		go func() {
			time.Sleep(10 * time.Millisecond)
			listener.Emit(1)
		}()
		<-ctx.Done()
		return ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.Run(ctx, "ignored") }()
	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run() did not return within grace policy")
	}
	if !canceledSeen.Load() {
		t.Fatal("callback did not observe cancellation")
	}
}
