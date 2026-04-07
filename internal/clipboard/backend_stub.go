//go:build !windows
// +build !windows

package clipboard

import (
	"context"
	"errors"
)

var ErrUnsupportedPlatform = errors.New("clipboard: unsupported platform")

type unsupportedBackend struct{}

func NewPlatformBackend() Backend {
	return unsupportedBackend{}
}

func (unsupportedBackend) ReadText(context.Context) (string, error) {
	return "", ErrUnsupportedPlatform
}

func (unsupportedBackend) WriteText(context.Context, string) error {
	return ErrUnsupportedPlatform
}
