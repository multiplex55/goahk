package goahk

import "testing"

func TestOnSimpleKeysStored(t *testing.T) {
	a := NewApp()
	a.On("1").Do(Log("digit"))
	a.On("Escape").Do(Log("esc"))

	p := a.toProgram()
	if got, want := p.Bindings[0].Hotkey, "1"; got != want {
		t.Fatalf("hotkey[0] = %q, want %q", got, want)
	}
	if got, want := p.Bindings[1].Hotkey, "Escape"; got != want {
		t.Fatalf("hotkey[1] = %q, want %q", got, want)
	}
}

func TestOnModifierComboNormalizedAndStored(t *testing.T) {
	a := NewApp()
	a.On("shift+control+v").Do(Log("combo"))
	p := a.toProgram()
	if got, want := p.Bindings[0].Hotkey, "Ctrl+Shift+V"; got != want {
		t.Fatalf("hotkey = %q, want %q", got, want)
	}
}
