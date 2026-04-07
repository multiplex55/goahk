package hotkey

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	win32ModAlt     = 0x0001
	win32ModControl = 0x0002
	win32ModShift   = 0x0004
	win32ModWin     = 0x0008

	win32WMHotkey  = 0x0312
	win32WMQuit    = 0x0012
	win32WMApp     = 0x8000
	win32WMCommand = win32WMApp + 1

	win32VKBackspace   = 0x08
	win32VKTab         = 0x09
	win32VKReturn      = 0x0D
	win32VKEscape      = 0x1B
	win32VKSpace       = 0x20
	win32VKPageUp      = 0x21
	win32VKPageDown    = 0x22
	win32VKEnd         = 0x23
	win32VKHome        = 0x24
	win32VKLeft        = 0x25
	win32VKUp          = 0x26
	win32VKRight       = 0x27
	win32VKDown        = 0x28
	win32VKInsert      = 0x2D
	win32VKDelete      = 0x2E
	win32VKPrintScreen = 0x2C
)

var (
	ErrDuplicateRegistration = errors.New("duplicate registration")
	ErrRegistrationNotFound  = errors.New("registration not found")
	ErrListenerClosed        = errors.New("listener is closed")
)

type win32Message struct {
	Message uint32
	WParam  uintptr
}

type win32Backend interface {
	registerHotKey(id int, modifiers uint32, vk uint32) error
	unregisterHotKey(id int) error
	getMessage() (win32Message, bool, error)
	postThreadMessage(threadID uint32, message uint32, wParam uintptr, lParam uintptr) error
	postQuitMessage(exitCode int32)
	currentThreadID() uint32
}

type win32Registration struct {
	modifiers uint32
	vk        uint32
}

type listenerCommand struct {
	register *registerCommand
	id       int
	resp     chan error
}

type registerCommand struct {
	id  int
	reg win32Registration
}

type Win32Listener struct {
	backend win32Backend
	events  chan ListenerEvent

	mu       sync.RWMutex
	closed   bool
	knownIDs map[int]struct{}

	commands chan listenerCommand
	done     chan struct{}
	errMu    sync.Mutex
	loopErr  error

	ready    chan struct{}
	threadID uint32
}

func NewWin32Listener() (*Win32Listener, error) {
	backend, err := newSystemWin32Backend()
	if err != nil {
		return nil, err
	}
	return newWin32ListenerWithBackend(backend), nil
}

func newWin32ListenerWithBackend(backend win32Backend) *Win32Listener {
	l := &Win32Listener{
		backend:  backend,
		events:   make(chan ListenerEvent, 64),
		knownIDs: map[int]struct{}{},
		commands: make(chan listenerCommand, 64),
		done:     make(chan struct{}),
		ready:    make(chan struct{}),
	}
	go l.runLoop()
	<-l.ready
	return l
}

func (l *Win32Listener) Events() <-chan ListenerEvent { return l.events }

func (l *Win32Listener) Register(registrationID int, chord Chord) error {
	reg, err := chordToWin32(chord)
	if err != nil {
		return err
	}

	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return ErrListenerClosed
	}
	if _, exists := l.knownIDs[registrationID]; exists {
		l.mu.Unlock()
		return fmt.Errorf("registration id %d: %w", registrationID, ErrDuplicateRegistration)
	}
	l.mu.Unlock()

	resp := make(chan error, 1)
	l.commands <- listenerCommand{register: &registerCommand{id: registrationID, reg: reg}, resp: resp}
	if err := l.backend.postThreadMessage(l.threadID, win32WMCommand, 0, 0); err != nil {
		return err
	}
	if err := <-resp; err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return ErrListenerClosed
	}
	l.knownIDs[registrationID] = struct{}{}
	return nil
}

func (l *Win32Listener) Unregister(registrationID int) error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return ErrListenerClosed
	}
	if _, exists := l.knownIDs[registrationID]; !exists {
		l.mu.Unlock()
		return fmt.Errorf("registration id %d: %w", registrationID, ErrRegistrationNotFound)
	}
	l.mu.Unlock()

	resp := make(chan error, 1)
	l.commands <- listenerCommand{id: registrationID, resp: resp}
	if err := l.backend.postThreadMessage(l.threadID, win32WMCommand, 0, 0); err != nil {
		return err
	}
	if err := <-resp; err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.knownIDs, registrationID)
	return nil
}

func (l *Win32Listener) Close() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	l.mu.Unlock()

	resp := make(chan error, 1)
	l.commands <- listenerCommand{id: -1, resp: resp}
	if err := l.backend.postThreadMessage(l.threadID, win32WMCommand, 0, 0); err != nil {
		return err
	}
	<-resp
	<-l.done

	close(l.events)
	return l.getLoopErr()
}

func (l *Win32Listener) runLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(l.done)

	l.threadID = l.backend.currentThreadID()
	close(l.ready)

	registered := map[int]win32Registration{}
	for {
		msg, ok, err := l.backend.getMessage()
		if err != nil {
			l.setLoopErr(err)
			return
		}
		if !ok || msg.Message == win32WMQuit {
			return
		}

		switch msg.Message {
		case win32WMHotkey:
			id := int(msg.WParam)
			if _, exists := registered[id]; !exists {
				continue
			}
			select {
			case l.events <- ListenerEvent{RegistrationID: id, TriggeredAt: time.Now().UTC()}:
			default:
			}
		case win32WMCommand:
			for {
				select {
				case cmd := <-l.commands:
					if cmd.register != nil {
						err := l.backend.registerHotKey(cmd.register.id, cmd.register.reg.modifiers, cmd.register.reg.vk)
						if err == nil {
							registered[cmd.register.id] = cmd.register.reg
						}
						cmd.resp <- err
						continue
					}
					if cmd.id == -1 {
						for id := range registered {
							_ = l.backend.unregisterHotKey(id)
							delete(registered, id)
						}
						cmd.resp <- nil
						l.backend.postQuitMessage(0)
						continue
					}
					err := l.backend.unregisterHotKey(cmd.id)
					if err == nil {
						delete(registered, cmd.id)
					}
					cmd.resp <- err
				default:
					goto drained
				}
			}
		}
	drained:
	}
}

func (l *Win32Listener) setLoopErr(err error) {
	l.errMu.Lock()
	defer l.errMu.Unlock()
	l.loopErr = err
}

func (l *Win32Listener) getLoopErr() error {
	l.errMu.Lock()
	defer l.errMu.Unlock()
	return l.loopErr
}

func chordToWin32(chord Chord) (win32Registration, error) {
	mods := uint32(0)
	if chord.Modifiers&ModAlt != 0 {
		mods |= win32ModAlt
	}
	if chord.Modifiers&ModCtrl != 0 {
		mods |= win32ModControl
	}
	if chord.Modifiers&ModShift != 0 {
		mods |= win32ModShift
	}
	if chord.Modifiers&ModWin != 0 {
		mods |= win32ModWin
	}
	vk, err := keyToVirtualKey(chord.Key)
	if err != nil {
		return win32Registration{}, err
	}
	return win32Registration{modifiers: mods, vk: vk}, nil
}

func keyToVirtualKey(key string) (uint32, error) {
	if len(key) == 1 {
		c := key[0]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			return uint32(c), nil
		}
	}

	switch key {
	case "Backspace":
		return win32VKBackspace, nil
	case "Tab":
		return win32VKTab, nil
	case "Enter":
		return win32VKReturn, nil
	case "Escape":
		return win32VKEscape, nil
	case "Space":
		return win32VKSpace, nil
	case "PageUp":
		return win32VKPageUp, nil
	case "PageDown":
		return win32VKPageDown, nil
	case "End":
		return win32VKEnd, nil
	case "Home":
		return win32VKHome, nil
	case "Left":
		return win32VKLeft, nil
	case "Up":
		return win32VKUp, nil
	case "Right":
		return win32VKRight, nil
	case "Down":
		return win32VKDown, nil
	case "Insert":
		return win32VKInsert, nil
	case "Delete":
		return win32VKDelete, nil
	case "PrintScreen":
		return win32VKPrintScreen, nil
	case "F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12":
		return 0x70 + parseFunctionOffset(key), nil
	default:
		return 0, fmt.Errorf("unsupported key for win32 listener: %q", key)
	}
}

func parseFunctionOffset(key string) uint32 {
	if len(key) == 2 {
		return uint32(key[1] - '1')
	}
	return 9 + uint32(key[2]-'0')
}
