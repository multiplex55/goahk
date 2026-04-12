package inspect

import "testing"

func TestFormatDisplayLabel(t *testing.T) {
	tests := []struct {
		name        string
		elementName string
		localized   string
		controlType string
		wantLabel   string
	}{
		{name: "empty name", localized: "edit", controlType: "Edit", wantLabel: `edit ""`},
		{name: "localized control type primary", elementName: "Search", localized: "document", controlType: "Pane", wantLabel: `document "Search"`},
		{name: "special chars escaped", elementName: `A "quoted" value`, localized: "edit", controlType: "Edit", wantLabel: `edit "A \"quoted\" value"`},
		{name: "missing localized control type", elementName: "Search", controlType: "Pane", wantLabel: `element "Search"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatDisplayLabel(tc.elementName, tc.localized, tc.controlType); got != tc.wantLabel {
				t.Fatalf("formatDisplayLabel() = %q, want %q", got, tc.wantLabel)
			}
		})
	}
}
