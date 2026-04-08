//go:build windows
// +build windows

package messagebox

import "testing"

func TestParseIcon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "default fallback empty", in: "", want: mbIconInfo},
		{name: "default fallback unknown", in: "toast", want: mbIconInfo},
		{name: "trimmed warning alias", in: "  warn\t", want: mbIconWarning},
		{name: "case insensitive warning", in: "WaRnInG", want: mbIconWarning},
		{name: "stop alias", in: " STOP ", want: mbIconError},
		{name: "question", in: "Question", want: mbIconQuestion},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := parseIcon(tt.in); got != tt.want {
				t.Fatalf("parseIcon(%q) = %#x, want %#x", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "default fallback empty", in: "", want: mbOK},
		{name: "default fallback unknown", in: "retry", want: mbOK},
		{name: "trimmed snake case", in: "  ok_cancel\n", want: mbOKCancel},
		{name: "alias no underscore", in: "okcancel", want: mbOKCancel},
		{name: "case insensitive alias", in: "OkCaNcEl", want: mbOKCancel},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := parseOptions(tt.in); got != tt.want {
				t.Fatalf("parseOptions(%q) = %#x, want %#x", tt.in, got, tt.want)
			}
		})
	}
}
