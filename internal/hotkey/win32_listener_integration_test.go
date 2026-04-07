//go:build windows
// +build windows

package hotkey

import "testing"

func TestWin32ListenerIntegrationRegisterUnregisterLifecycle(t *testing.T) {
	l, err := NewWin32Listener()
	if err != nil {
		t.Fatalf("NewWin32Listener() err = %v", err)
	}
	defer func() { _ = l.Close() }()

	if err := l.Register(301, Chord{Modifiers: ModCtrl | ModAlt, Key: "K"}); err != nil {
		t.Fatalf("register err = %v", err)
	}
	if err := l.Unregister(301); err != nil {
		t.Fatalf("unregister err = %v", err)
	}
}

func TestWin32ListenerIntegrationCloseCleansUpRegistrations(t *testing.T) {
	l, err := NewWin32Listener()
	if err != nil {
		t.Fatalf("NewWin32Listener() err = %v", err)
	}

	if err := l.Register(401, Chord{Modifiers: ModCtrl, Key: "P"}); err != nil {
		t.Fatalf("register #1 err = %v", err)
	}
	if err := l.Register(402, Chord{Modifiers: ModCtrl | ModShift, Key: "P"}); err != nil {
		t.Fatalf("register #2 err = %v", err)
	}

	if err := l.Close(); err != nil {
		t.Fatalf("close err = %v", err)
	}
}
