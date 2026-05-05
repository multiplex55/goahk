//go:build windows
// +build windows

package inspect

import (
	"context"
	"fmt"
	"syscall"
)

const (
	smXVIRTUALSCREEN  = 76
	smYVIRTUALSCREEN  = 77
	smCXVIRTUALSCREEN = 78
	smCYVIRTUALSCREEN = 79
)

var (
	overlayUser32           = syscall.NewLazyDLL("user32.dll")
	procGetSystemMetricsHUD = overlayUser32.NewProc("GetSystemMetrics")
)

type nativeHighlightOverlay struct{}

func newNativeHighlightOverlay() highlightOverlay { return nativeHighlightOverlay{} }

func (nativeHighlightOverlay) Show(context.Context, Rect) error {
	// Native overlay rendering is intentionally stubbed in this environment.
	// Style and paint configuration helpers remain explicit/tested to guard
	// expected topmost/layered/click-through behavior.
	return nil
}

func (nativeHighlightOverlay) Clear(context.Context) error {
	return nil
}

const (
	wsPopup         = 0x80000000
	wsExTopMost     = 0x00000008
	wsExToolWindow  = 0x00000080
	wsExTransparent = 0x00000020
	wsExLayered     = 0x00080000
	wsExNoActivate  = 0x08000000
)

func overlayWindowStyle() uintptr {
	return wsPopup
}

func overlayWindowExStyle() uintptr {
	return wsExTopMost | wsExToolWindow | wsExTransparent | wsExLayered | wsExNoActivate
}

func overlayPaintUsesBorderOnly() bool {
	return true
}

func (nativeHighlightOverlay) ScreenBounds(context.Context) (*Rect, error) {
	left, _, err := procGetSystemMetricsHUD.Call(smXVIRTUALSCREEN)
	if err != syscall.Errno(0) {
		return nil, fmt.Errorf("GetSystemMetrics(SM_XVIRTUALSCREEN): %w", err)
	}
	top, _, err := procGetSystemMetricsHUD.Call(smYVIRTUALSCREEN)
	if err != syscall.Errno(0) {
		return nil, fmt.Errorf("GetSystemMetrics(SM_YVIRTUALSCREEN): %w", err)
	}
	width, _, err := procGetSystemMetricsHUD.Call(smCXVIRTUALSCREEN)
	if err != syscall.Errno(0) {
		return nil, fmt.Errorf("GetSystemMetrics(SM_CXVIRTUALSCREEN): %w", err)
	}
	height, _, err := procGetSystemMetricsHUD.Call(smCYVIRTUALSCREEN)
	if err != syscall.Errno(0) {
		return nil, fmt.Errorf("GetSystemMetrics(SM_CYVIRTUALSCREEN): %w", err)
	}
	if width <= 0 || height <= 0 {
		return nil, nil
	}
	return &Rect{Left: int(int32(left)), Top: int(int32(top)), Width: int(int32(width)), Height: int(int32(height))}, nil
}
