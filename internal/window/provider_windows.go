//go:build windows
// +build windows

package window

func NewOSProvider() *OSProvider {
	return &OSProvider{}
}
