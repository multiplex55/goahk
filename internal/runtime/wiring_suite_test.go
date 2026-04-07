package runtime

import "testing"

func TestRuntimeWiringSuite(t *testing.T) {
	t.Run("compile_to_plan_mapping", func(t *testing.T) {
		testRuntimeWiringCompileToPlanMapping(t)
	})
	t.Run("dispatch_behavior", func(t *testing.T) {
		testRuntimeWiringDispatchBehavior(t)
	})
	t.Run("early_failure", func(t *testing.T) {
		testRuntimeWiringEarlyFailure(t)
	})
	t.Run("clean_stop", func(t *testing.T) {
		testRuntimeWiringCleanStop(t)
	})
}

func testRuntimeWiringCompileToPlanMapping(t *testing.T) {
	t.Helper()
	TestBootstrap_RegistersCompiledBindings(t)
}

func testRuntimeWiringDispatchBehavior(t *testing.T) {
	t.Helper()
	TestBootstrap_DispatchExecutesCorrectBindingPlan(t)
}

func testRuntimeWiringEarlyFailure(t *testing.T) {
	t.Helper()
	TestBootstrap_FailsFastOnUnknownAction(t)
}

func testRuntimeWiringCleanStop(t *testing.T) {
	t.Helper()
	TestBootstrap_ShutdownCancellationStopsCleanly(t)
}
