package goahk

import (
	"context"
	"errors"
	stdruntime "runtime"
	"strings"
	"testing"
)

func TestAPIHappyPath_TinyScriptFlow(t *testing.T) {

	app := NewApp().
		Bind("1", MessageBox("Quick note", "Pressed 1"), Stop()).
		Bind("Escape", ControlStop())

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
		if stdruntime.GOOS == "windows" {
			t.Fatalf("Run() error = %v, want nil", err)
		}
		if !errors.Is(err, ErrUnsupportedPlatform) {
			t.Fatalf("Run() error = %v, want ErrUnsupportedPlatform on non-windows", err)
		}
	}
}

func TestAPIHappyPath_UnknownPolicyRejected(t *testing.T) {
	app := NewApp()
	app.On("Ctrl+H").WithPolicy("fastest").Do(Log("x"))

	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("Run() error = nil, want invalid policy error")
	}
	if !strings.Contains(err.Error(), `unsupported concurrency policy "fastest"`) {
		t.Fatalf("Run() error = %q, want unsupported policy detail", err.Error())
	}
}

func TestAPIHappyPath_RuntimeArtifactsCarryPolicyAndDefault(t *testing.T) {
	app := NewApp().
		On("Ctrl+R").Replace().Do(Log("replace")).
		On("Ctrl+S").Do(Log("serial-default"))

	p, cfg, _ := app.runtimeArtifacts()
	if got := string(p.Bindings[0].ConcurrencyPolicy); got != "replace" {
		t.Fatalf("program policy[0] = %q, want replace", got)
	}
	if got := cfg.Hotkeys[0].ConcurrencyPolicy; got != "replace" {
		t.Fatalf("config policy[0] = %q, want replace", got)
	}
	if got := string(p.Bindings[1].ConcurrencyPolicy); got != "serial" {
		t.Fatalf("program policy[1] = %q, want serial", got)
	}
	if got := cfg.Hotkeys[1].ConcurrencyPolicy; got != "serial" {
		t.Fatalf("config policy[1] = %q, want serial", got)
	}
}
