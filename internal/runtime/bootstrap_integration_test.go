//go:build windows && integration

package runtime

import (
	"context"
	"reflect"
	"testing"
	"time"

	"goahk/internal/config"
)

func TestBootstrap_WindowsLifecycle_RegisterRunShutdown(t *testing.T) {
	listener := newFakeListener()
	listener.runFn = func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}
	cfg := config.Config{Hotkeys: []config.HotkeyBinding{
		{ID: "one", Hotkey: "ctrl+1", Steps: []config.Step{{Action: "system.log"}}},
		{ID: "two", Hotkey: "ctrl+2", Steps: []config.Step{{Action: "system.log"}}},
	}}

	b := NewBootstrap()
	b.LoadConfig = func(context.Context, string) (config.Config, error) { return cfg, nil }
	b.NewListener = func(context.Context) (Listener, error) { return listener, nil }

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
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for bootstrap shutdown")
	}

	listener.mu.Lock()
	defer listener.mu.Unlock()
	if got := len(listener.registered); got != 2 {
		t.Fatalf("registered = %d, want 2", got)
	}
	if !reflect.DeepEqual(listener.unregistered, []int{2, 1}) {
		t.Fatalf("unregistered = %#v, want []int{2,1}", listener.unregistered)
	}
	if !listener.closed {
		t.Fatal("listener was not closed")
	}
}
