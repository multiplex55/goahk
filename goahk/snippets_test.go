package goahk_test

import (
	"context"
	"testing"

	"goahk/goahk"
)

func TestScriptModeSnippetCompilesAndRuns(t *testing.T) {
	t.Parallel()

	app := goahk.NewApp()
	app.Bind("1", goahk.MessageBox("goahk", "Hello from hotkey 1"))
	app.Bind("Escape", goahk.Stop())

	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestConfigMappingSnippetCompilesAndRuns(t *testing.T) {
	t.Parallel()

	app := goahk.NewApp()
	app.Bind("Ctrl+Alt+H",
		goahk.MessageBox("goahk", "Hello from Ctrl+Alt+H"),
	)

	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestSnippetComposesActionsAndCallbackStep(t *testing.T) {
	t.Parallel()

	app := goahk.NewApp()
	app.Bind("Ctrl+Shift+C",
		goahk.ClipboardRead("clip"),
		goahk.Func(func(ctx *goahk.Context) error {
			clip := ctx.Metadata()["clip"]
			ctx.Metadata()["clip"] = "[" + clip + "]"
			return nil
		}),
		goahk.ClipboardWrite("done"),
	)

	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestSnippetComposesWindowAndInputStepsWithCallback(t *testing.T) {
	t.Parallel()

	app := goahk.NewApp()
	app.Bind("Ctrl+Shift+T",
		goahk.CopyActiveWindowTitle(),
		goahk.Func(func(ctx *goahk.Context) error {
			ctx.Metadata()["title_snapshot"] = ctx.Metadata()["clipboard"]
			return nil
		}),
		goahk.SendChord("ctrl", "l"),
		goahk.SendKeys("ctrl+v {enter}"),
	)

	if err := app.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}
