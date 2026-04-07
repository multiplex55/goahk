//go:build !windows

package hotkey

import "fmt"

func newSystemWin32Backend() (win32Backend, error) {
	return nil, fmt.Errorf("win32 listener is only available on windows")
}
