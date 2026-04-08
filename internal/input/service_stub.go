//go:build !windows
// +build !windows

package input

import (
	"context"
	"errors"
)

var ErrUnsupportedPlatform = errors.New("input: unsupported platform")

type unsupportedService struct{}

type unsupportedPlatformError struct{}

func (unsupportedPlatformError) Error() string     { return ErrUnsupportedPlatform.Error() }
func (unsupportedPlatformError) Unsupported() bool { return true }

func newPlatformService() Service {
	return unsupportedService{}
}

func (unsupportedService) SendText(context.Context, string, SendOptions) error {
	return unsupportedPlatformError{}
}

func (unsupportedService) SendKeys(context.Context, Sequence, SendOptions) error {
	return unsupportedPlatformError{}
}

func (unsupportedService) SendChord(context.Context, Chord, SendOptions) error {
	return unsupportedPlatformError{}
}
