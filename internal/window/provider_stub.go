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
