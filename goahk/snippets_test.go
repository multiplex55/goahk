package goahk_test

import (
	"context"
	"errors"
	stdruntime "runtime"
	"testing"

	"goahk/goahk"
)

func assertRunResultForPlatform(t *testing.T, err error) {
	t.Helper()
	if stdruntime.GOOS == "windows" {
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		return
	}
	if !errors.Is(err, goahk.ErrUnsupportedPlatform) {
		t.Fatalf("Run() error = %v, want unsupported platform on non-windows", err)
	}
}

func TestScriptModeSnippetCompilesAndRuns(t *testing.T) {

	app := goahk.NewApp()
	app.Bind("1", goahk.MessageBox("goahk", "Hello from hotkey 1"))
	app.Bind("Escape", goahk.ControlStop())

	assertRunResultForPlatform(t, app.Run(context.Background()))
}

func TestConfigMappingSnippetCompilesAndRuns(t *testing.T) {

	app := goahk.NewApp()
	app.Bind("Ctrl+Alt+H",
		goahk.MessageBox("goahk", "Hello from Ctrl+Alt+H"),
	)

	assertRunResultForPlatform(t, app.Run(context.Background()))
}

func TestSnippetComposesActionsAndCallbackStep(t *testing.T) {

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

	assertRunResultForPlatform(t, app.Run(context.Background()))
}

func TestSnippetComposesWindowAndInputStepsWithCallback(t *testing.T) {

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

	assertRunResultForPlatform(t, app.Run(context.Background()))
}
