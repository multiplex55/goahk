package hotkey

import "testing"

func TestHotkeyManagerSuite(t *testing.T) {
	t.Run("lifecycle", func(t *testing.T) {
		testManagerSuiteLifecycle(t)
	})
	t.Run("mapping", func(t *testing.T) {
		testManagerSuiteMapping(t)
	})
	t.Run("cancellation", func(t *testing.T) {
		testManagerSuiteCancellation(t)
	})
	t.Run("listener_close", func(t *testing.T) {
		testManagerSuiteListenerClose(t)
	})
	t.Run("safe_shutdown", func(t *testing.T) {
		testManagerSuiteSafeShutdown(t)
	})
}

func testManagerSuiteLifecycle(t *testing.T) {
	t.Helper()
	TestManagerRegisterUnregisterLifecycle(t)
}

func testManagerSuiteMapping(t *testing.T) {
	t.Helper()
	TestManagerEventMappingRegistrationIDToBindingID(t)
}

func testManagerSuiteCancellation(t *testing.T) {
	t.Helper()
	TestManagerRunExitsOnContextCancellationAndClosesStreams(t)
}

func testManagerSuiteListenerClose(t *testing.T) {
	t.Helper()
	TestManagerRunExitsWhenListenerEventChannelCloses(t)
}

func testManagerSuiteSafeShutdown(t *testing.T) {
	t.Helper()
	TestManagerCloseWhileRunningDoesNotPanic(t)
}
