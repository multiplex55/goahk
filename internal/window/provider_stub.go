//go:build !windows
// +build !windows

package window

import (
	"context"
	"errors"
	"fmt"
	"runtime"
)

var ErrUnsupportedPlatform = errors.New("window provider is only supported on Windows")

type OSProvider struct{}

func NewOSProvider() *OSProvider { return &OSProvider{} }

func (p *OSProvider) EnumerateWindows(context.Context) ([]Info, error) {
	return nil, unsupportedPlatformError("enumerate windows")
}

func (p *OSProvider) ActiveWindow(context.Context) (Info, error) {
	return Info{}, unsupportedPlatformError("read active window")
}

func (p *OSProvider) ActivateWindow(context.Context, HWND) error {
	return unsupportedPlatformError("activate window")
}

func (p *OSProvider) WindowBounds(context.Context, HWND) (Rect, error) {
	return Rect{}, unsupportedPlatformError("read window bounds")
}

func (p *OSProvider) WorkAreaForWindow(context.Context, HWND) (Rect, error) {
	return Rect{}, unsupportedPlatformError("read monitor work area")
}

func (p *OSProvider) MoveWindow(context.Context, HWND, int, int) error {
	return unsupportedPlatformError("move window")
}

func (p *OSProvider) ResizeWindow(context.Context, HWND, int, int) error {
	return unsupportedPlatformError("resize window")
}

func (p *OSProvider) MinimizeWindow(context.Context, HWND) error {
	return unsupportedPlatformError("minimize window")
}

func (p *OSProvider) MaximizeWindow(context.Context, HWND) error {
	return unsupportedPlatformError("maximize window")
}

func (p *OSProvider) RestoreWindow(context.Context, HWND) error {
	return unsupportedPlatformError("restore window")
}

func unsupportedPlatformError(operation string) error {
	return fmt.Errorf("%w: %s is unavailable on %s (requires Windows)", ErrUnsupportedPlatform, operation, runtime.GOOS)
}
