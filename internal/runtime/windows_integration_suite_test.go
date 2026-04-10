//go:build windows && integration

package runtime

import (
	"os"
	"testing"
)

func TestWindowsIntegrationSuite(t *testing.T) {
	t.Run("runtime_start", func(t *testing.T) {
		t.Skip("post-listener: implement runtime start integration assertion")
	})
	t.Run("single_hotkey_registration", func(t *testing.T) {
		t.Skip("post-listener: verify exactly one hotkey registration")
	})
	t.Run("trigger_path_verification", func(t *testing.T) {
		t.Skip("post-listener: verify listener trigger reaches dispatch/action path")
	})
	t.Run("clean_shutdown_unregister", func(t *testing.T) {
		t.Skip("post-listener: verify shutdown unregisters hotkeys and exits cleanly")
	})
	t.Run("input_backend_e2e", func(t *testing.T) {
		if os.Getenv("GOAHK_ENABLE_WINDOWS_INPUT_ITEST") != "1" {
			t.Skip("set GOAHK_ENABLE_WINDOWS_INPUT_ITEST=1 to enable real SendInput integration checks")
		}
		t.Skip("TODO: assert SendInput keyboard/mouse end-to-end behavior against foreground app")
	})
}
