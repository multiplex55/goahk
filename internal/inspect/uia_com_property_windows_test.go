//go:build windows
// +build windows

package inspect

import "testing"

func TestControlTypeNameForID(t *testing.T) {
	if got := controlTypeNameForID(50032); got != "Window" {
		t.Fatalf("controlTypeNameForID(50032)=%q", got)
	}
	if got := controlTypeNameForID(50033); got != "Pane" {
		t.Fatalf("controlTypeNameForID(50033)=%q", got)
	}
	if got := controlTypeNameForID(50030); got != "Document" {
		t.Fatalf("controlTypeNameForID(50030)=%q", got)
	}
	if got := controlTypeNameForID(59999); got != "ControlType(59999)" {
		t.Fatalf("unexpected fallback %q", got)
	}
}

func TestDecodeVariant_ComPointerAndArrays(t *testing.T) {
	if got := decodeVariant(comVariant{VT: vtUnknown, Val: 0x1234}); got.Status != propertyStatusOK || got.S == "" {
		t.Fatalf("vtUnknown pointer decode failed: %+v", got)
	}
	if got := decodeVariant(comVariant{VT: vtDispatch, Val: 0x5678}); got.Status != propertyStatusOK || got.S == "" {
		t.Fatalf("vtDispatch pointer decode failed: %+v", got)
	}
	if got := decodeVariant(comVariant{VT: vtArray | vtI4, Val: 0}); got.Status != propertyStatusEmpty {
		t.Fatalf("expected empty VT_ARRAY|VT_I4 for nil safearray: %+v", got)
	}
	if got := decodeVariant(comVariant{VT: vtArray | vtR8, Val: 0}); got.Status != propertyStatusEmpty {
		t.Fatalf("expected empty VT_ARRAY|VT_R8 for nil safearray: %+v", got)
	}
}
