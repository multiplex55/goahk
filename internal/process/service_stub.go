//go:build !windows
// +build !windows

package process

import (
	"context"
	"errors"
)

var ErrUnsupportedPlatform = errors.New("process: unsupported platform")

type unsupportedService struct{}

func newPlatformService() Service {
	return unsupportedService{}
}

func (unsupportedService) Launch(context.Context, Request) error {
	return ErrUnsupportedPlatform
}
