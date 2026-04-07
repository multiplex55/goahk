package hotkey

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type managerTestListener struct {
	events chan ListenerEvent

	mu          sync.Mutex
	registerIDs []int
	unregisters []int
	closeCount  int
	closeErr    error
	closeOnce   sync.Once
}

func newManagerTestListener() *managerTestListener {
	return &managerTestListener{events: make(chan ListenerEvent, 16)}
}

func (l *managerTestListener) Register(registrationID int, _ Chord) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.registerIDs = append(l.registerIDs, registrationID)
	return nil
}

func (l *managerTestListener) Unregister(registrationID int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.unregisters = append(l.unregisters, registrationID)
	return nil
}

func (l *managerTestListener) Events() <-chan ListenerEvent { return l.events }

func (l *managerTestListener) Close() error {
	l.mu.Lock()
	l.closeCount++
	l.mu.Unlock()
	l.closeOnce.Do(func() {
		close(l.events)
	})
	return l.closeErr
}

func (l *managerTestListener) closeCalls() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.closeCount
}

func (l *managerTestListener) registeredIDs() []int {
	l.mu.Lock()
	defer l.mu.Unlock()
	ids := make([]int, len(l.registerIDs))
	copy(ids, l.registerIDs)
	return ids
}

func (l *managerTestListener) unregisteredIDs() []int {
	l.mu.Lock()
	defer l.mu.Unlock()
	ids := make([]int, len(l.unregisters))
	copy(ids, l.unregisters)
	return ids
}

func TestManagerCloseWhileRunningDoesNotPanic(t *testing.T) {
	listener := newManagerTestListener()
	manager := NewManager(listener)
	if err := manager.Register("binding", Chord{Modifiers: ModCtrl, Key: "K"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := make(chan error, 1)
	go func() {
		runErr <- manager.Run(ctx)
	}()

	listener.events <- ListenerEvent{RegistrationID: 1, TriggeredAt: time.Now().UTC()}
	if err := manager.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}

	if err := <-runErr; err != nil {
		t.Fatalf("run error: %v", err)
	}
	if got := listener.closeCalls(); got != 1 {
		t.Fatalf("listener close calls = %d, want 1", got)
	}
}

func TestManagerRunExitsWhenListenerEventChannelCloses(t *testing.T) {
	listener := newManagerTestListener()
	manager := NewManager(listener)

	runErr := make(chan error, 1)
	go func() {
		runErr <- manager.Run(context.Background())
	}()

	close(listener.events)

	if err := <-runErr; err != nil {
		t.Fatalf("run error: %v", err)
	}
	if _, ok := <-manager.Events(); ok {
		t.Fatal("expected trigger event stream to be closed")
	}
	if _, ok := <-manager.Events(); ok {
		t.Fatal("expected closed trigger event stream on subsequent read")
	}
}

func TestManagerRunExitsOnContextCancellationAndClosesStreams(t *testing.T) {
	listener := newManagerTestListener()
	manager := NewManager(listener)

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		runErr <- manager.Run(ctx)
	}()

	cancel()

	if err := <-runErr; !errors.Is(err, context.Canceled) {
		t.Fatalf("run error = %v, want %v", err, context.Canceled)
	}
	if _, ok := <-manager.Events(); ok {
		t.Fatal("expected trigger event stream to be closed")
	}
}

func TestManagerRegisterUnregisterLifecycle(t *testing.T) {
	listener := newManagerTestListener()
	manager := NewManager(listener)

	if err := manager.Register("binding", Chord{Modifiers: ModAlt, Key: "X"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := manager.Unregister("binding"); err != nil {
		t.Fatalf("unregister: %v", err)
	}
	if err := manager.Unregister("binding"); err == nil {
		t.Fatal("expected error unregistering missing binding")
	}

	registered := listener.registeredIDs()
	if len(registered) != 1 || registered[0] != 1 {
		t.Fatalf("registered ids = %v, want [1]", registered)
	}
	unregistered := listener.unregisteredIDs()
	if len(unregistered) != 1 || unregistered[0] != 1 {
		t.Fatalf("unregistered ids = %v, want [1]", unregistered)
	}
}

func TestManagerDuplicateBindingRegistration(t *testing.T) {
	listener := newManagerTestListener()
	manager := NewManager(listener)

	if err := manager.Register("binding", Chord{Modifiers: ModCtrl, Key: "Z"}); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := manager.Register("binding", Chord{Modifiers: ModCtrl, Key: "Z"}); err == nil {
		t.Fatal("expected duplicate binding registration error")
	}

	registered := listener.registeredIDs()
	if len(registered) != 1 {
		t.Fatalf("register call count = %d, want 1", len(registered))
	}
}

func TestManagerEventMappingRegistrationIDToBindingID(t *testing.T) {
	listener := newManagerTestListener()
	manager := NewManager(listener)
	chordA := Chord{Modifiers: ModCtrl, Key: "A"}
	chordB := Chord{Modifiers: ModShift, Key: "B"}
	if err := manager.Register("first", chordA); err != nil {
		t.Fatalf("register first: %v", err)
	}
	if err := manager.Register("second", chordB); err != nil {
		t.Fatalf("register second: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runErr := make(chan error, 1)
	go func() {
		runErr <- manager.Run(ctx)
	}()

	triggeredAt := time.Now().UTC()
	listener.events <- ListenerEvent{RegistrationID: 2, TriggeredAt: triggeredAt}

	select {
	case ev := <-manager.Events():
		if ev.BindingID != "second" {
			t.Fatalf("binding id = %q, want %q", ev.BindingID, "second")
		}
		if ev.Chord != chordB {
			t.Fatalf("chord = %#v, want %#v", ev.Chord, chordB)
		}
		if !ev.TriggeredAt.Equal(triggeredAt) {
			t.Fatalf("triggered at = %v, want %v", ev.TriggeredAt, triggeredAt)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for mapped trigger event")
	}

	cancel()
	if err := <-runErr; !errors.Is(err, context.Canceled) {
		t.Fatalf("run error = %v, want %v", err, context.Canceled)
	}
}
