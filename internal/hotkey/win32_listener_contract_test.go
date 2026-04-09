package hotkey

import (
	"errors"
	"testing"
	"time"
)

func TestWin32ListenerContract_RegisterUnregisterEventSemantics(t *testing.T) {
	backend := newFakeWin32Backend()
	listener := newWin32ListenerWithBackend(backend)
	defer func() { _ = listener.Close() }()

	if err := listener.Register(10, Chord{Modifiers: ModCtrl | ModShift, Key: "K"}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := listener.Register(10, Chord{Modifiers: ModCtrl | ModShift, Key: "K"}); !errors.Is(err, ErrDuplicateRegistration) {
		t.Fatalf("duplicate Register() error = %v, want ErrDuplicateRegistration", err)
	}

	backend.messages <- win32Message{Message: win32WMHotkey, WParam: 10}
	select {
	case got := <-listener.Events():
		if got.RegistrationID != 10 {
			t.Fatalf("RegistrationID = %d, want 10", got.RegistrationID)
		}
		if got.TriggeredAt.IsZero() {
			t.Fatal("TriggeredAt must be set")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for hotkey event")
	}

	if err := listener.Unregister(10); err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}
	if err := listener.Unregister(10); !errors.Is(err, ErrRegistrationNotFound) {
		t.Fatalf("second Unregister() error = %v, want ErrRegistrationNotFound", err)
	}

	backend.messages <- win32Message{Message: win32WMHotkey, WParam: 10}
	select {
	case <-listener.Events():
		t.Fatal("did not expect event after unregistration")
	case <-time.After(100 * time.Millisecond):
	}
}
