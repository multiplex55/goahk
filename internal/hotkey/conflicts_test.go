package hotkey

import "testing"

func TestDetectConflicts_DuplicateNormalized(t *testing.T) {
	a, _ := ParseBinding("one", "Ctrl+Shift+V")
	b, _ := ParseBinding("two", "shift+control+v")
	if err := DetectConflicts([]Binding{a, b}); err == nil {
		t.Fatal("DetectConflicts() expected conflict error")
	}
}

func TestDetectConflicts_NoDuplicates(t *testing.T) {
	a, _ := ParseBinding("one", "Ctrl+Shift+V")
	b, _ := ParseBinding("two", "Ctrl+Shift+C")
	if err := DetectConflicts([]Binding{a, b}); err != nil {
		t.Fatalf("DetectConflicts() error = %v", err)
	}
}
