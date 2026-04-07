//go:build !windows

package messagebox

import (
	"context"
	"errors"
)

var ErrUnsupportedPlatform = errors.New("messagebox: unsupported platform")

type unsupportedService struct{}

func newPlatformService() Service {
	return unsupportedService{}
}

func (unsupportedService) Show(context.Context, Request) error {
	return ErrUnsupportedPlatform
}
