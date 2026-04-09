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

func (unsupportedService) MoveAbsolute(context.Context, int, int) error {
	return unsupportedPlatformError{}
}

func (unsupportedService) MoveRelative(context.Context, int, int) error {
	return unsupportedPlatformError{}
}

func (unsupportedService) Position(context.Context) (MousePosition, error) {
	return MousePosition{}, unsupportedPlatformError{}
}

func (unsupportedService) ButtonDown(context.Context, string) error {
	return unsupportedPlatformError{}
}

func (unsupportedService) ButtonUp(context.Context, string) error {
	return unsupportedPlatformError{}
}

func (unsupportedService) Click(context.Context, string) error {
	return unsupportedPlatformError{}
}

func (unsupportedService) DoubleClick(context.Context, string) error {
	return unsupportedPlatformError{}
}

func (unsupportedService) Wheel(context.Context, int) error {
	return unsupportedPlatformError{}
}

func (unsupportedService) Drag(context.Context, string, int, int, int, int) error {
	return unsupportedPlatformError{}
}
