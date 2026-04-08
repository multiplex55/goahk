//go:build !windows
// +build !windows

package folders

import (
	"context"
	"errors"
)

var ErrUnsupportedPlatform = errors.New("folders: unsupported platform")

type unsupportedService struct{}

func newPlatformService() Service {
	return unsupportedService{}
}

func (unsupportedService) ListOpenFolders(context.Context) ([]FolderInfo, error) {
	return nil, ErrUnsupportedPlatform
}
