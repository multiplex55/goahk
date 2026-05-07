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
		{name: "missing localized control type", elementName: "Search", controlType: "Pane", wantLabel: `pane "Search"`},
		{name: "fallback controltype prefix", elementName: "Search", controlType: "ControlType.Button", wantLabel: `button "Search"`},
		{name: "fallback empty type still quoted", elementName: "", controlType: "", wantLabel: `element ""`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatDisplayLabel(tc.elementName, tc.localized, tc.controlType); got != tc.wantLabel {
				t.Fatalf("formatDisplayLabel() = %q, want %q", got, tc.wantLabel)
			}
		})
	}
}

func TestFormatDisplayLabel_QuotedNameFormatting(t *testing.T) {
	got := formatDisplayLabel(`He said "hello"`, "edit", "Edit")
	want := `edit "He said \"hello\""`
	if got != want {
		t.Fatalf("formatDisplayLabel() = %q, want %q", got, want)
	}
}

func TestFormatDisplayLabel_ControlTypeFallbacks(t *testing.T) {
	tests := []struct {
		name      string
		localized string
		control   string
		want      string
	}{
		{name: "controltype prefix stripped", control: "ControlType.Button", want: `button ""`},
		{name: "plain fallback lowercased", control: "Pane", want: `pane ""`},
		{name: "empty fallback uses element", control: "", want: `element ""`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatDisplayLabel("", tc.localized, tc.control); got != tc.want {
				t.Fatalf("formatDisplayLabel() = %q, want %q", got, tc.want)
			}
		})
	}
}
