//go:build windows
// +build windows

package clipboard

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	cfUnicodeText = 13
	gmEMoveable   = 0x0002
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procGetClipboardData = user32.NewProc("GetClipboardData")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")
	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
)

type windowsBackend struct{}

func NewPlatformBackend() Backend {
	return windowsBackend{}
}

func (windowsBackend) ReadText(ctx context.Context) (string, error) {
	if err := openClipboard(ctx); err != nil {
		return "", err
	}
	defer procCloseClipboard.Call()

	h, _, _ := procGetClipboardData.Call(cfUnicodeText)
	if h == 0 {
		return "", nil
	}
	ptr, _, _ := procGlobalLock.Call(h)
	if ptr == 0 {
		return "", fmt.Errorf("clipboard: GlobalLock failed")
	}
	defer procGlobalUnlock.Call(h)

	units := make([]uint16, 0, 128)
	for i := 0; ; i++ {
		u := *(*uint16)(unsafe.Pointer(ptr + uintptr(i*2)))
		units = append(units, u)
		if u == 0 {
			break
		}
	}
	text, err := DecodeUTF16(units)
	if err != nil {
		return "", err
	}
	return NormalizeReadText(text), nil
}

func (windowsBackend) WriteText(ctx context.Context, text string) error {
	if err := openClipboard(ctx); err != nil {
		return err
	}
	defer procCloseClipboard.Call()
	if r, _, _ := procEmptyClipboard.Call(); r == 0 {
		return fmt.Errorf("clipboard: EmptyClipboard failed")
	}

	units := EncodeUTF16(text)
	size := uintptr(len(units) * 2)
	hMem, _, _ := procGlobalAlloc.Call(gmEMoveable, size)
	if hMem == 0 {
		return fmt.Errorf("clipboard: GlobalAlloc failed")
	}
	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		return fmt.Errorf("clipboard: GlobalLock failed")
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(ptr)), len(units)), units)
	procGlobalUnlock.Call(hMem)
	if r, _, _ := procSetClipboardData.Call(cfUnicodeText, hMem); r == 0 {
		return fmt.Errorf("clipboard: SetClipboardData failed")
	}
	return nil
}

func openClipboard(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if r, _, _ := procOpenClipboard.Call(0); r == 0 {
		return fmt.Errorf("clipboard: OpenClipboard failed")
	}
	return nil
}
