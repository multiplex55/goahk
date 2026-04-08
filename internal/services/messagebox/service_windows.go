//go:build windows
// +build windows

package messagebox

import (
	"context"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	mbOK       = 0x00000000
	mbOKCancel = 0x00000001

	mbIconInfo     = 0x00000040
	mbIconWarning  = 0x00000030
	mbIconError    = 0x00000010
	mbIconQuestion = 0x00000020
)

var (
	messageBoxUser32 = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW  = messageBoxUser32.NewProc("MessageBoxW")
)

type nativeService struct{}

func newPlatformService() Service {
	return nativeService{}
}

func (nativeService) Show(ctx context.Context, req Request) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if strings.TrimSpace(req.Body) == "" {
		return fmt.Errorf("messagebox: body is required")
	}

	flags := uintptr(mbOK | parseIcon(req.Icon) | parseOptions(req.Options))
	bodyPtr, err := syscall.UTF16PtrFromString(req.Body)
	if err != nil {
		return fmt.Errorf("messagebox: body: %w", err)
	}
	titlePtr, err := syscall.UTF16PtrFromString(req.Title)
	if err != nil {
		return fmt.Errorf("messagebox: title: %w", err)
	}

	r, _, callErr := procMessageBoxW.Call(0, uintptr(unsafe.Pointer(bodyPtr)), uintptr(unsafe.Pointer(titlePtr)), flags)
	if r == 0 && callErr != syscall.Errno(0) {
		return fmt.Errorf("messagebox: MessageBoxW failed: %w", callErr)
	}
	return nil
}

func parseIcon(icon string) int {
	switch strings.ToLower(strings.TrimSpace(icon)) {
	case "error", "stop":
		return mbIconError
	case "warning", "warn":
		return mbIconWarning
	case "question":
		return mbIconQuestion
	default:
		return mbIconInfo
	}
}

func parseOptions(options string) int {
	switch strings.ToLower(strings.TrimSpace(options)) {
	case "ok_cancel", "okcancel":
		return mbOKCancel
	default:
		return mbOK
	}
}
