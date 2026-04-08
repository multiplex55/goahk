//go:build !windows
// +build !windows

package folders

import (
	"context"
	"errors"
	"testing"
)

func TestUnsupportedService_ListOpenFolders(t *testing.T) {
	svc := newPlatformService()
	_, err := svc.ListOpenFolders(context.Background())
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("err = %v, want %v", err, ErrUnsupportedPlatform)
	}
}
