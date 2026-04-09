package hotkey

import (
	"context"
	"sync"
	"testing"
	"time"
)

type shutdownRaceListener struct {
	mu     sync.Mutex
	events chan ListenerEvent
	closed bool
	closes int
}

func newShutdownRaceListener() *shutdownRaceListener {
	return &shutdownRaceListener{events: make(chan ListenerEvent, 16)}
}

func (l *shutdownRaceListener) Register(int, Chord) error { return nil }
func (l *shutdownRaceListener) Unregister(int) error      { return nil }
func (l *shutdownRaceListener) Events() <-chan ListenerEvent {
	return l.events
}

func (l *shutdownRaceListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closes++
	if l.closed {
		return nil
	}
	l.closed = true
	close(l.events)
	return nil
}

func (l *shutdownRaceListener) emit(id int) {
	l.mu.Lock()
	closed := l.closed
	l.mu.Unlock()
	if closed {
		return
	}
	select {
	case l.events <- ListenerEvent{RegistrationID: id, TriggeredAt: time.Now().UTC()}:
	default:
	}
}

func TestManagerShutdownRace_RunCloseFuzzLike(t *testing.T) {
	for i := 0; i < 200; i++ {
		listener := newShutdownRaceListener()
		manager := NewManager(listener)
		if err := manager.Register("hk", Chord{Modifiers: ModCtrl, Key: "K"}); err != nil {
			t.Fatalf("iteration %d register: %v", i, err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		runDone := make(chan error, 1)
		go func() {
			runDone <- manager.Run(ctx)
		}()

		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			listener.emit(1)
		}()
		go func() {
			defer wg.Done()
			_ = manager.Close()
		}()
		go func() {
			defer wg.Done()
			cancel()
			_ = manager.Close()
		}()
		wg.Wait()

		select {
		case <-manager.Events():
		default:
		}
		if _, ok := <-manager.Events(); ok {
			t.Fatalf("iteration %d: manager events should be closed", i)
		}

		select {
		case err := <-runDone:
			if err != nil && err != context.Canceled {
				t.Fatalf("iteration %d run error: %v", i, err)
			}
		case <-time.After(time.Second):
			t.Fatalf("iteration %d: timed out waiting for Run to exit", i)
		}

		listener.mu.Lock()
		closeCalls := listener.closes
		listener.mu.Unlock()
		if closeCalls != 1 {
			t.Fatalf("iteration %d: listener close calls = %d, want 1", i, closeCalls)
		}
	}
}

func TestManagerRunAfterCloseReturnsWithoutPanic(t *testing.T) {
	listener := newShutdownRaceListener()
	manager := NewManager(listener)
	if err := manager.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if err := manager.Run(context.Background()); err != nil {
		t.Fatalf("run after close error = %v, want nil", err)
	}
	if _, ok := <-manager.Events(); ok {
		t.Fatal("expected closed events channel after close")
	}
}

func TestManagerListenerClosesBeforeManagerClose(t *testing.T) {
	listener := newShutdownRaceListener()
	manager := NewManager(listener)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runDone := make(chan error, 1)
	go func() { runDone <- manager.Run(ctx) }()

	if err := listener.Close(); err != nil {
		t.Fatalf("listener close: %v", err)
	}
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("run error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for run to exit after listener close")
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("manager close: %v", err)
	}
}

func TestManagerCloseBeforeListenerNaturallyStops(t *testing.T) {
	listener := newShutdownRaceListener()
	manager := NewManager(listener)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runDone := make(chan error, 1)
	go func() { runDone <- manager.Run(ctx) }()

	if err := manager.Close(); err != nil {
		t.Fatalf("manager close: %v", err)
	}
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("run error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for run to exit after manager close")
	}
}
