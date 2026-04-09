package goahk

import (
	"context"
	"strings"
	"testing"

	"goahk/internal/program"
)

type nilActionStep struct{}

func (*nilActionStep) stepSpec() program.StepSpec {
	return program.StepSpec{Action: "system.log", Params: map[string]any{"message": "noop"}}
}

func TestAPIErrorMessages_BadChordIncludesHelpfulCompileContext(t *testing.T) {
	t.Parallel()

	app := NewApp().Bind("Ctrl+Nope", Log("x"))
	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("Run() error = nil, want failure")
	}
	msg := err.Error()
	for _, token := range []string{"compile app program", "binding_1", "unsupported key"} {
		if !strings.Contains(msg, token) {
			t.Fatalf("Run() error = %q, missing %q", msg, token)
		}
	}
}

func TestAPIErrorMessages_EmptyBindingAndNilActions(t *testing.T) {
	t.Parallel()

	var nilStep *nilActionStep
	app := NewApp()
	app.Bind("", Log("x"))
	app.Bind("Ctrl+E")
	app.Bind("Ctrl+N", nilStep)
	app.Bind("Ctrl+C", Func(nil))

	err := app.Run(context.Background())
	if err == nil {
		t.Fatal("Run() error = nil, want failure")
	}
	msg := err.Error()
	for _, token := range []string{
		"invalid binding wiring",
		"binding hotkey cannot be empty",
		`binding "Ctrl+E" must include at least one action`,
		`binding "Ctrl+N" step 1 is nil`,
		`binding "Ctrl+C" step 1 callback is nil`,
	} {
		if !strings.Contains(msg, token) {
			t.Fatalf("Run() error = %q, missing %q", msg, token)
		}
	}
}
