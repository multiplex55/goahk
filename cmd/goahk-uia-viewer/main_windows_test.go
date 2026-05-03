//go:build windows

package main

import "testing"

func TestRunWindows_SymbolAvailable(t *testing.T) {
	t.Parallel()
	_ = runWindows
}
