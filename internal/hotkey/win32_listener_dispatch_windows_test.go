//go:build windows
// +build windows

package hotkey

import (
	"testing"
	"time"
)

func TestWin32ListenerWindowsDispatchesRegisteredEvent(t *testing.T) {
	backend := newFakeWin32Backend()
	l := newWin32ListenerWithBackend(backend)
	defer func() { _ = l.Close() }()

	if err := l.Register(777, Chord{Modifiers: ModCtrl, Key: "K"}); err != nil {
		t.Fatalf("register err = %v", err)
	}

	backend.messages <- win32Message{Message: win32WMHotkey, WParam: 777}

	select {
	case ev := <-l.Events():
		if ev.RegistrationID != 777 {
			t.Fatalf("registration id = %d, want 777", ev.RegistrationID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for dispatched hotkey event")
	}
}
