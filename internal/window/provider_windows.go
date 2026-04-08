//go:build windows
// +build windows

package window

import (
	"context"

	platformwindows "goahk/internal/platform/windows"
)

type OSProvider struct {
	inner *platformwindows.WindowProvider
}

func NewOSProvider() *OSProvider { return &OSProvider{inner: platformwindows.NewWindowProvider()} }

func (p *OSProvider) EnumerateWindows(ctx context.Context) ([]Info, error) {
	return p.inner.EnumerateWindows(ctx)
}

func (p *OSProvider) ActiveWindow(ctx context.Context) (Info, error) {
	return p.inner.ActiveWindow(ctx)
}

func (p *OSProvider) ActivateWindow(ctx context.Context, hwnd HWND) error {
	return p.inner.ActivateWindow(ctx, hwnd)
}
