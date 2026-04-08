package runtime

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/config"
	"goahk/internal/hotkey"
	"goahk/internal/program"
)

type fakeListener struct {
	mu           sync.Mutex
	events       chan hotkey.ListenerEvent
	registered   []hotkey.Chord
	unregistered []int
	closed       bool
	runFn        func(context.Context) error
}

func newFakeListener() *fakeListener {
	return &fakeListener{events: make(chan hotkey.ListenerEvent, 32)}
}

func (f *fakeListener) Register(registrationID int, chord hotkey.Chord) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.registered = append(f.registered, chord)
	return nil
}

func (f *fakeListener) Unregister(registrationID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unregistered = append(f.unregistered, registrationID)
	return nil
}

func (f *fakeListener) Events() <-chan hotkey.ListenerEvent { return f.events }

func (f *fakeListener) Run(ctx context.Context) error {
	if f.runFn != nil {
		return f.runFn(ctx)
	}
	<-ctx.Done()
	return ctx.Err()
}

func (f *fakeListener) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true
	close(f.events)
	return nil
}

func (f *fakeListener) Emit(registrationID int) {
	f.events <- hotkey.ListenerEvent{RegistrationID: registrationID, TriggeredAt: time.Now().UTC()}
}

func TestBootstrap_RegistersCompiledBindings(t *testing.T) {
	listener := newFakeListener()
	listener.runFn = func(context.Context) error { return nil }
	cfg := config.Config{Hotkeys: []config.HotkeyBinding{
		{ID: "one", Hotkey: "ctrl+1", Steps: []config.Step{{Action: "system.log"}}},
		{ID: "two", Hotkey: "ctrl+2", Steps: []config.Step{{Action: "system.log"}}},
	}}

	b := NewBootstrap()
	b.LoadConfig = func(context.Context, string) (config.Config, error) { return cfg, nil }
	b.NewListener = func(context.Context) (Listener, error) { return listener, nil }

	if err := b.Run(context.Background(), "ignored"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	listener.mu.Lock()
	defer listener.mu.Unlock()
	if len(listener.registered) != 2 {
		t.Fatalf("registered count = %d, want 2", len(listener.registered))
	}
	if listener.registered[0].String() != "Ctrl+1" || listener.registered[1].String() != "Ctrl+2" {
		t.Fatalf("registered chords = %#v", listener.registered)
	}
}

func TestBootstrap_DispatchExecutesCorrectBindingPlan(t *testing.T) {
	listener := newFakeListener()
	cfg := config.Config{Hotkeys: []config.HotkeyBinding{
		{ID: "one", Hotkey: "ctrl+1", Steps: []config.Step{{Action: "test.one"}}},
		{ID: "two", Hotkey: "ctrl+2", Steps: []config.Step{{Action: "test.two"}}},
	}}
	resultCh := make(chan string, 1)

	b := NewBootstrap()
	b.LoadConfig = func(context.Context, string) (config.Config, error) { return cfg, nil }
	b.BuildRegistry = func(context.Context, program.Program) (*actions.Registry, error) {
		reg := actions.NewRegistry()
		_ = reg.Register("test.one", func(actions.ActionContext, actions.Step) error { return nil })
		_ = reg.Register("test.two", func(_ actions.ActionContext, _ actions.Step) error { return nil })
		return reg, nil
	}
	b.NewListener = func(context.Context) (Listener, error) { return listener, nil }
	b.RecordResult = func(_ context.Context, bindingID string, _ actions.ExecutionResult) {
		select {
		case resultCh <- bindingID:
		default:
		}
	}
	listener.runFn = func(ctx context.Context) error {
		go func() {
			time.Sleep(10 * time.Millisecond)
			listener.Emit(2)
		}()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Millisecond):
			return nil
		}
	}

	if err := b.Run(context.Background(), "ignored"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	select {
	case got := <-resultCh:
		if got != "two" {
			t.Fatalf("executed binding = %q, want %q", got, "two")
		}
	default:
		t.Fatal("no dispatch result recorded")
	}
}

func TestBootstrap_ShutdownCancellationStopsCleanly(t *testing.T) {
	listener := newFakeListener()
	cfg := config.Config{Hotkeys: []config.HotkeyBinding{{ID: "one", Hotkey: "ctrl+1", Steps: []config.Step{{Action: "system.log"}}}}}
	b := NewBootstrap()
	b.LoadConfig = func(context.Context, string) (config.Config, error) { return cfg, nil }
	b.NewListener = func(context.Context) (Listener, error) { return listener, nil }
	listener.runFn = func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- b.Run(ctx, "ignored") }()
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not return after cancellation")
	}
}

func TestBootstrap_FailsFastOnUnknownAction(t *testing.T) {
	cfg := config.Config{Hotkeys: []config.HotkeyBinding{{ID: "one", Hotkey: "ctrl+1", Steps: []config.Step{{Action: "does.not.exist"}}}}}
	calledListener := false
	calledBuildRegistry := false
	b := NewBootstrap()
	b.LoadConfig = func(context.Context, string) (config.Config, error) { return cfg, nil }
	b.BuildRegistry = func(context.Context, program.Program) (*actions.Registry, error) {
		calledBuildRegistry = true
		return actions.NewRegistry(), nil
	}
	b.NewListener = func(context.Context) (Listener, error) {
		calledListener = true
		return newFakeListener(), nil
	}

	err := b.Run(context.Background(), "ignored")
	if err == nil {
		t.Fatal("Run() error = nil, want failure")
	}
	if calledListener {
		t.Fatal("listener should not be created when compile fails")
	}
	if !calledBuildRegistry {
		t.Fatal("registry should be built before compile-time validation")
	}
	msg := err.Error()
	for _, token := range []string{"compile runtime bindings", `binding "one"`, `binding/actions[0]/name`, `"does.not.exist"`} {
		if !strings.Contains(msg, token) {
			t.Fatalf("Run() error = %q, missing %q", msg, token)
		}
	}
}

func TestBootstrap_EscBindingWithRuntimeStopExits(t *testing.T) {
	listener := newFakeListener()
	cfg := config.Config{Hotkeys: []config.HotkeyBinding{
		{ID: "quit", Hotkey: "ctrl+esc", Steps: []config.Step{{Action: "runtime.stop"}}},
	}}
	b := NewBootstrap()
	b.LoadConfig = func(context.Context, string) (config.Config, error) { return cfg, nil }
	b.NewListener = func(context.Context) (Listener, error) { return listener, nil }
	listener.runFn = func(ctx context.Context) error {
		go func() {
			time.Sleep(10 * time.Millisecond)
			listener.Emit(1)
		}()
		<-ctx.Done()
		return ctx.Err()
	}

	done := make(chan error, 1)
	go func() { done <- b.Run(context.Background(), "ignored") }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not exit after Esc/runtime.stop")
	}
}
