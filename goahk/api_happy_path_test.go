package goahk

import (
	"context"
	"testing"
)

func TestAPIHappyPath_TinyScriptFlow(t *testing.T) {
	t.Parallel()

	app := NewApp().
		Bind("1", MessageBox("Quick note", "Pressed 1"), Stop()).
		Bind("Escape", Stop())

	prog := app.toProgram()
	if got, want := len(prog.Bindings), 2; got != want {
		t.Fatalf("bindings len = %d, want %d", got, want)
	}
	if got, want := prog.Bindings[0].Hotkey, "1"; got != want {
		t.Fatalf("binding[0] hotkey = %q, want %q", got, want)
	}
	if got, want := prog.Bindings[1].Hotkey, "Escape"; got != want {
		t.Fatalf("binding[1] hotkey = %q, want %q", got, want)
	}
	if got, want := prog.Bindings[0].Steps[0].Action, "system.message_box"; got != want {
		t.Fatalf("binding[0] step[0] action = %q, want %q", got, want)
	}
	if got, want := prog.Bindings[0].Steps[1].Action, "runtime.stop"; got != want {
		t.Fatalf("binding[0] step[1] action = %q, want %q", got, want)
	}
	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
}
