//go:build windows
// +build windows

package clipboard

func NewPlatformBackend() Backend {
	return windowsBackend{}
}
