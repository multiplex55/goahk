package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/hotkey"
)

type fakeContractListener struct {
	events chan hotkey.ListenerEvent
}

func (f *fakeContractListener) Register(int, hotkey.Chord) error { return nil }
func (f *fakeContractListener) Unregister(int) error             { return nil }
func (f *fakeContractListener) Events() <-chan hotkey.ListenerEvent {
	return f.events
}
func (f *fakeContractListener) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
func (f *fakeContractListener) Close() error { return nil }

func TestListenerContractAcceptsNarrowInterface(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b := Bootstrap{}
	if err := b.runLoop(ctx, &fakeContractListener{events: make(chan hotkey.ListenerEvent)}, make(chan error, 1)); err != nil {
		t.Fatalf("runLoop err = %v", err)
	}
}

func TestListenerContract_CallbackExecutionCannotBlockEventIngestion(t *testing.T) {
	reg := actions.NewRegistry()
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	reg.MustRegisterCallback("slow", func(ctx actions.CallbackContext) error {
		started <- struct{}{}
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	executor := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 8)
	shutdown := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var controlSeen atomic.Int32
	results := DispatchHotkeyEvents(ctx, shutdown, events, map[string]actions.Plan{
		"work": {{Name: actions.CallbackActionName, Params: map[string]string{"callback_ref": "slow"}}},
		"quit": {},
	}, map[string]RuntimeControlCommand{"quit": RuntimeControlStop}, executor, actions.ActionContext{}, nil, func(ev runtimeControlEvent) {
		if ev.Command == RuntimeControlStop {
			controlSeen.Add(1)
			cancel()
		}
	})
	events <- hotkey.TriggerEvent{BindingID: "work", Chord: hotkey.Chord{Key: "F9"}}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("callback job did not start")
	}
	events <- hotkey.TriggerEvent{BindingID: "quit", Chord: hotkey.Chord{Key: "Escape"}}
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("listener ingestion stalled while callback was running")
	}
	close(release)
	close(shutdown)
	for range results {
	}
	if controlSeen.Load() != 1 {
		t.Fatalf("control events handled = %d, want 1", controlSeen.Load())
	}
}
