package actions

import "testing"

func TestServiceAdapterSuites(t *testing.T) {
	t.Run("message_box", func(t *testing.T) {
		testServiceAdapterMessageBox(t)
	})
	t.Run("clipboard", func(t *testing.T) {
		testServiceAdapterClipboard(t)
	})
	t.Run("process", func(t *testing.T) {
		testServiceAdapterProcess(t)
	})
	t.Run("window_activation", func(t *testing.T) {
		testServiceAdapterWindowActivation(t)
	})
}

func testServiceAdapterMessageBox(t *testing.T) {
	t.Helper()
	TestMessageBoxAction_NormalizesParams(t)
	TestActionValidation_MissingRequiredFields(t)
}

func testServiceAdapterClipboard(t *testing.T) {
	t.Helper()
	TestClipboardActions_Semantics(t)
	TestClipboardActions_MissingServiceErrors(t)
}

func testServiceAdapterProcess(t *testing.T) {
	t.Helper()
	TestProcessLaunchAction_NormalizesParams(t)
	TestServiceErrorPropagation_PreservedInExecutionResult(t)
}

func testServiceAdapterWindowActivation(t *testing.T) {
	t.Helper()
	TestWindowActions_MissingServiceErrors(t)
}
