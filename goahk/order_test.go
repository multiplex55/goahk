package goahk

import "testing"

func TestChainedActionsPreserveOrder(t *testing.T) {
	a := NewApp()
	a.On("Ctrl+Alt+H").Do(Log("1"), SendText("2"), ClipboardWrite("3"))

	p := a.toProgram()
	got := p.Bindings[0].Steps
	if len(got) != 3 {
		t.Fatalf("steps len = %d, want 3", len(got))
	}
	if got[0].Action != "system.log" || got[1].Action != "input.send_text" || got[2].Action != "clipboard.write" {
		t.Fatalf("steps order = [%s, %s, %s]", got[0].Action, got[1].Action, got[2].Action)
	}
}
