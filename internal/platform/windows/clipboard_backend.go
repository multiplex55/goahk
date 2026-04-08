//go:build windows
// +build windows

package windows

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"

	"goahk/internal/clipboard"
)

const (
	cfUnicodeText = 13
	gmEMoveable   = 0x0002
)

var (
	clipboardUser32      = syscall.NewLazyDLL("user32.dll")
	clipboardKernel32    = syscall.NewLazyDLL("kernel32.dll")
	procOpenClipboard    = clipboardUser32.NewProc("OpenClipboard")
	procCloseClipboard   = clipboardUser32.NewProc("CloseClipboard")
	procGetClipboardData = clipboardUser32.NewProc("GetClipboardData")
	procSetClipboardData = clipboardUser32.NewProc("SetClipboardData")
	procEmptyClipboard   = clipboardUser32.NewProc("EmptyClipboard")
	procGlobalAlloc      = clipboardKernel32.NewProc("GlobalAlloc")
	procGlobalLock       = clipboardKernel32.NewProc("GlobalLock")
	procGlobalUnlock     = clipboardKernel32.NewProc("GlobalUnlock")
)

type ClipboardBackend struct{}

func NewClipboardBackend() clipboard.Backend {
	return ClipboardBackend{}
}

func (ClipboardBackend) ReadText(ctx context.Context) (string, error) {
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
	text, err := clipboard.DecodeUTF16(units)
	if err != nil {
		return "", err
	}
	return clipboard.NormalizeReadText(text), nil
}

func (ClipboardBackend) WriteText(ctx context.Context, text string) error {
	if err := openClipboard(ctx); err != nil {
		return err
	}
	defer procCloseClipboard.Call()
	if r, _, _ := procEmptyClipboard.Call(); r == 0 {
		return fmt.Errorf("clipboard: EmptyClipboard failed")
	}

	units := clipboard.EncodeUTF16(text)
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
