package hotkey

import (
	"strings"
	"testing"
)

func TestDetectConflictReport_UserDiagnostics(t *testing.T) {
	a, _ := ParseBinding("paste", "Ctrl+Shift+V")
	b, _ := ParseBinding("alt_paste", "shift+control+v")
	report := DetectConflictReport([]Binding{a, b})
	if report == nil || len(report.Conflicts) != 1 {
		t.Fatalf("report = %#v", report)
	}
	msg := report.UserMessage()
	for _, want := range []string{"Ctrl+Shift+V", "paste", "alt_paste", "choose unique chords"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message missing %q: %s", want, msg)
		}
	}
}
