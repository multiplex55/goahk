//go:build windows

package hotkey

import (
	"fmt"
	"syscall"
	"unsafe"
)

type systemWin32Backend struct {
	user32                 *syscall.LazyDLL
	procRegisterHotKey     *syscall.LazyProc
	procUnregisterHotKey   *syscall.LazyProc
	procGetMessageW        *syscall.LazyProc
	procPostThreadMessageW *syscall.LazyProc
	procPostQuitMessage    *syscall.LazyProc
	procGetCurrentThreadID *syscall.LazyProc
}

type point struct{ X, Y int32 }

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

func newSystemWin32Backend() (win32Backend, error) {
	dll := syscall.NewLazyDLL("user32.dll")
	if err := dll.Load(); err != nil {
		return nil, err
	}
	k32 := syscall.NewLazyDLL("kernel32.dll")
	if err := k32.Load(); err != nil {
		return nil, err
	}
	return &systemWin32Backend{
		user32:                 dll,
		procRegisterHotKey:     dll.NewProc("RegisterHotKey"),
		procUnregisterHotKey:   dll.NewProc("UnregisterHotKey"),
		procGetMessageW:        dll.NewProc("GetMessageW"),
		procPostThreadMessageW: dll.NewProc("PostThreadMessageW"),
		procPostQuitMessage:    dll.NewProc("PostQuitMessage"),
		procGetCurrentThreadID: k32.NewProc("GetCurrentThreadId"),
	}, nil
}

func (b *systemWin32Backend) registerHotKey(id int, modifiers uint32, vk uint32) error {
	r1, _, err := b.procRegisterHotKey.Call(0, uintptr(id), uintptr(modifiers), uintptr(vk))
	if r1 == 0 {
		return err
	}
	return nil
}

func (b *systemWin32Backend) unregisterHotKey(id int) error {
	r1, _, err := b.procUnregisterHotKey.Call(0, uintptr(id))
	if r1 == 0 {
		return err
	}
	return nil
}

func (b *systemWin32Backend) getMessage() (win32Message, bool, error) {
	var m msg
	r1, _, err := b.procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
	switch int32(r1) {
	case -1:
		return win32Message{}, false, err
	case 0:
		return win32Message{Message: win32WMQuit}, false, nil
	default:
		return win32Message{Message: m.Message, WParam: m.WParam}, true, nil
	}
}

func (b *systemWin32Backend) postThreadMessage(threadID uint32, message uint32, wParam uintptr, lParam uintptr) error {
	r1, _, err := b.procPostThreadMessageW.Call(uintptr(threadID), uintptr(message), wParam, lParam)
	if r1 == 0 {
		return fmt.Errorf("PostThreadMessageW: %w", err)
	}
	return nil
}

func (b *systemWin32Backend) postQuitMessage(exitCode int32) {
	b.procPostQuitMessage.Call(uintptr(exitCode))
}

func (b *systemWin32Backend) currentThreadID() uint32 {
	r1, _, _ := b.procGetCurrentThreadID.Call()
	return uint32(r1)
}
