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
