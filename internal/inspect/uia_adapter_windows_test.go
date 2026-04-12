//go:build windows
// +build windows

package inspect

import (
	"errors"
	"testing"
)

func TestMapUIAError_TypedErrors(t *testing.T) {
	if got := mapUIAError(&UIAComUnavailableError{Op: "GetFocusedElement", Err: errors.New("rpc unavailable")}); !errors.Is(got, ErrProviderTransientFailure) {
		t.Fatalf("com unavailable should map to transient failure, got %v", got)
	}
	if got := mapUIAError(&UIAElementStaleError{Op: "ElementByRuntimeID", Err: errors.New("stale")}); !errors.Is(got, errStaleElementReference) {
		t.Fatalf("stale should map to stale element, got %v", got)
	}
	unsup := &UIAUnsupportedPropertyError{Property: "HelpText", Err: errors.New("property unsupported")}
	if got := mapUIAError(unsup); got != unsup {
		t.Fatalf("unsupported property should pass through, got %v", got)
	}
}
