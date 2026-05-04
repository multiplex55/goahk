//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
)

// walkBootstrapCheck creates and disposes a minimal main window.
// This isolates Walk startup issues (for example tooltip/common-controls init).
func walkBootstrapCheck() error {
	mw, err := walk.NewMainWindow()
	if err != nil {
		return fmt.Errorf("walk bootstrap failed: %w", err)
	}
	mw.Dispose()
	return nil
}
