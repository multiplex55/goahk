//go:build windows
// +build windows

package messagebox

import platformwindows "goahk/internal/platform/windows"

func newPlatformService() Service {
	return platformwindows.NewMessageBoxService()
}
