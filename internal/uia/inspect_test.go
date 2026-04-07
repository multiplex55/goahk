package uia

import (
	"strings"
	"testing"
)

func TestFormatElementText_MissingProperties(t *testing.T) {
	got := FormatElementText(Element{ID: ""})
	for _, want := range []string{
		"ID: (unknown)",
		"Name: (missing)",
		"ControlType: (missing)",
		"Patterns: (none)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q in %q", want, got)
		}
	}
}

func TestFormatElementText_WithProperties(t *testing.T) {
	name, control := "Submit", "Button"
	className, autoID, fw := "Button", "submit-btn", "Win32"
	el := Element{
		ID:           "elem-1",
		Name:         &name,
		ControlType:  &control,
		ClassName:    &className,
		AutomationID: &autoID,
		FrameworkID:  &fw,
		Patterns:     []string{"Invoke", "LegacyIAccessible"},
	}
	got := FormatElementText(el)
	if !strings.Contains(got, "ID: elem-1") || !strings.Contains(got, "Patterns: Invoke, LegacyIAccessible") {
		t.Fatalf("unexpected output: %q", got)
	}
}
