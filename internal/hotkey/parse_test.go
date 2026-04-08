package hotkey

import "testing"

func TestParse_ValidAndNormalized(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"1", "1"},
		{"Ctrl+Shift+V", "Ctrl+Shift+V"},
		{"shift+control+v", "Ctrl+Shift+V"},
		{"option+cmd+esc", "Alt+Win+Escape"},
		{"ctrl+f12", "Ctrl+F12"},
	}
	for _, tt := range tests {
		chord, err := Parse(tt.in)
		if err != nil {
			t.Fatalf("Parse(%q) error = %v", tt.in, err)
		}
		if got := chord.String(); got != tt.want {
			t.Fatalf("Parse(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParse_Invalid(t *testing.T) {
	for _, in := range []string{"", "Ctrl+", "Ctrl+Shift", "Ctrl+Nope"} {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) expected error", in)
		}
	}
}

func TestParse_InvalidModifierOnlyAndMalformedCombinations(t *testing.T) {
	tests := []string{
		"Ctrl+Alt",
		"Shift+Win",
		"Ctrl++A",
		"Ctrl+Shift+",
		"Ctrl+Ctrl+K",
	}
	for _, in := range tests {
		if _, err := Parse(in); err == nil {
			t.Fatalf("Parse(%q) expected invalid combination error", in)
		}
	}
}
