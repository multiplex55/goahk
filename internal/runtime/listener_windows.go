//go:build windows

package runtime

import (
	"context"
	"sync"

	"goahk/internal/hotkey"
)

type windowsListener struct {
	mu     sync.Mutex
	closed bool
	events chan hotkey.ListenerEvent
	regs   map[int]hotkey.Chord
}

func NewWindowsListener(context.Context) (Listener, error) {
	return &windowsListener{
		events: make(chan hotkey.ListenerEvent, 32),
		regs:   map[int]hotkey.Chord{},
	}, nil
}

func (l *windowsListener) Register(registrationID int, chord hotkey.Chord) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return context.Canceled
	}
	l.regs[registrationID] = chord
	return nil
}

func (l *windowsListener) Unregister(registrationID int) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return context.Canceled
	}
	delete(l.regs, registrationID)
	return nil
}

func (l *windowsListener) Events() <-chan hotkey.ListenerEvent { return l.events }

func (l *windowsListener) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (l *windowsListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	close(l.events)
	return nil
}
