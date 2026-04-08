//go:build windows
// +build windows

package runtime

import (
	"context"

	platformwindows "goahk/internal/platform/windows"
)

func NewWindowsListener(ctx context.Context) (Listener, error) {
	return platformwindows.NewHotkeyListener(ctx)
}
