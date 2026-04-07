//go:build !windows

package runtime

import (
	"context"
	"fmt"
)

func NewWindowsListener(context.Context) (Listener, error) {
	return nil, fmt.Errorf("windows hotkey listener is only available on windows")
}
