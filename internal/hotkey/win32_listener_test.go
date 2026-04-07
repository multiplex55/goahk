package hotkey

import (
	"errors"
	"testing"
	"time"
)

type fakeWin32Backend struct {
	messages     chan win32Message
	registered   map[int]win32Registration
	unregistered []int
	threadID     uint32
	failRegister map[int]error
}

func newFakeWin32Backend() *fakeWin32Backend {
	return &fakeWin32Backend{
		messages:     make(chan win32Message, 64),
		registered:   map[int]win32Registration{},
		threadID:     42,
		failRegister: map[int]error{},
	}
}

func (f *fakeWin32Backend) registerHotKey(id int, modifiers uint32, vk uint32) error {
	if err, ok := f.failRegister[id]; ok {
		return err
	}
	f.registered[id] = win32Registration{modifiers: modifiers, vk: vk}
	return nil
}

func (f *fakeWin32Backend) unregisterHotKey(id int) error {
	f.unregistered = append(f.unregistered, id)
	delete(f.registered, id)
	return nil
}

func (f *fakeWin32Backend) getMessage() (win32Message, bool, error) {
	msg := <-f.messages
	if msg.Message == win32WMQuit {
		return msg, false, nil
	}
	return msg, true, nil
}

func (f *fakeWin32Backend) postThreadMessage(_ uint32, message uint32, wParam uintptr, _ uintptr) error {
	f.messages <- win32Message{Message: message, WParam: wParam}
	return nil
}

func (f *fakeWin32Backend) postQuitMessage(_ int32) {
	f.messages <- win32Message{Message: win32WMQuit}
}

func (f *fakeWin32Backend) currentThreadID() uint32 { return f.threadID }

func TestChordToWin32Mapping(t *testing.T) {
	tests := []struct {
		name string
		in   Chord
		mod  uint32
		vk   uint32
	}{
		{name: "ctrl_shift_letter", in: Chord{Modifiers: ModCtrl | ModShift, Key: "K"}, mod: win32ModControl | win32ModShift, vk: 'K'},
		{name: "alt_win_digit", in: Chord{Modifiers: ModAlt | ModWin, Key: "7"}, mod: win32ModAlt | win32ModWin, vk: '7'},
		{name: "ctrl_function", in: Chord{Modifiers: ModCtrl, Key: "F12"}, mod: win32ModControl, vk: 0x7B},
		{name: "ctrl_special", in: Chord{Modifiers: ModCtrl, Key: "Escape"}, mod: win32ModControl, vk: win32VKEscape},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := chordToWin32(tt.in)
			if err != nil {
				t.Fatalf("chordToWin32() err = %v", err)
			}
			if got.modifiers != tt.mod || got.vk != tt.vk {
				t.Fatalf("chordToWin32() = (%#x,%#x), want (%#x,%#x)", got.modifiers, got.vk, tt.mod, tt.vk)
			}
		})
	}
}

func TestWin32ListenerDuplicateRegistrationRejected(t *testing.T) {
	backend := newFakeWin32Backend()
	l := newWin32ListenerWithBackend(backend)
	defer func() { _ = l.Close() }()

	if err := l.Register(11, Chord{Modifiers: ModCtrl, Key: "K"}); err != nil {
		t.Fatalf("first register err = %v", err)
	}
	if err := l.Register(11, Chord{Modifiers: ModCtrl, Key: "K"}); !errors.Is(err, ErrDuplicateRegistration) {
		t.Fatalf("duplicate register err = %v, want ErrDuplicateRegistration", err)
	}
}

func TestWin32ListenerUnregisterAfterRegisterSucceeds(t *testing.T) {
	backend := newFakeWin32Backend()
	l := newWin32ListenerWithBackend(backend)
	defer func() { _ = l.Close() }()

	if err := l.Register(7, Chord{Modifiers: ModCtrl | ModShift, Key: "A"}); err != nil {
		t.Fatalf("register err = %v", err)
	}
	if err := l.Unregister(7); err != nil {
		t.Fatalf("unregister err = %v", err)
	}
	if _, exists := backend.registered[7]; exists {
		t.Fatalf("registration 7 should be removed")
	}
}

func TestWin32ListenerWMHotkeyMapsToRegistrationID(t *testing.T) {
	backend := newFakeWin32Backend()
	l := newWin32ListenerWithBackend(backend)
	defer func() { _ = l.Close() }()

	if err := l.Register(99, Chord{Modifiers: ModCtrl, Key: "V"}); err != nil {
		t.Fatalf("register err = %v", err)
	}
	backend.messages <- win32Message{Message: win32WMHotkey, WParam: 99}

	select {
	case ev := <-l.Events():
		if ev.RegistrationID != 99 {
			t.Fatalf("event registration id = %d, want 99", ev.RegistrationID)
		}
		if ev.TriggeredAt.IsZero() {
			t.Fatal("event TriggeredAt should be set")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for listener event")
	}
}
