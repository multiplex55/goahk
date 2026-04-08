//go:build !windows
// +build !windows

package window

import (
	"context"
	"errors"
)

var ErrUnsupportedPlatform = errors.New("window provider is only supported on Windows")

type OSProvider struct{}

func NewOSProvider() *OSProvider { return &OSProvider{} }

func (p *OSProvider) EnumerateWindows(context.Context) ([]Info, error) {
	return nil, ErrUnsupportedPlatform
}

func (p *OSProvider) ActiveWindow(context.Context) (Info, error) {
	return Info{}, ErrUnsupportedPlatform
}

func (p *OSProvider) ActivateWindow(context.Context, HWND) error {
	return ErrUnsupportedPlatform
}

func (p *OSProvider) WindowBounds(context.Context, HWND) (Rect, error) {
	return Rect{}, ErrUnsupportedPlatform
}

func (p *OSProvider) MoveWindow(context.Context, HWND, int, int) error {
	return ErrUnsupportedPlatform
}

func (p *OSProvider) ResizeWindow(context.Context, HWND, int, int) error {
	return ErrUnsupportedPlatform
}

func (p *OSProvider) MinimizeWindow(context.Context, HWND) error {
	return ErrUnsupportedPlatform
}

func (p *OSProvider) MaximizeWindow(context.Context, HWND) error {
	return ErrUnsupportedPlatform
}

func (p *OSProvider) RestoreWindow(context.Context, HWND) error {
	return ErrUnsupportedPlatform
}
