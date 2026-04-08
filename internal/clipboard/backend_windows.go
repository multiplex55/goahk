//go:build windows
// +build windows

package clipboard

import platformwindows "goahk/internal/platform/windows"

func NewPlatformBackend() Backend {
	return platformwindows.NewClipboardBackend()
}
