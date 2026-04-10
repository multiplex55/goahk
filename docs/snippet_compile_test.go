package docs_test

import (
	"testing"
	"time"

	"goahk/goahk"
)

// These snippets mirror callback examples in docs/examples/*.md and compile as
// part of CI so API drift is caught quickly.
func TestDocsCallbackSnippetsCompile(t *testing.T) {
	t.Parallel()

	app := goahk.NewApp()
	app.Bind("Ctrl+Shift+R", goahk.Func(func(ctx *goahk.Context) error {
		for {
			if err := ctx.Err(); err != nil {
				ctx.Log("callback canceled cleanly", "err", err)
				return err
			}
			if !ctx.Sleep(150 * time.Millisecond) {
				return ctx.Err()
			}
			ctx.Log("heartbeat", "binding", ctx.Binding(), "trigger", ctx.Trigger())
		}
	}))

	app.Bind("Ctrl+Shift+I", goahk.Func(func(ctx *goahk.Context) error {
		active, err := ctx.Window.Active()
		if err != nil {
			return err
		}
		ctx.Log("active window", "title", active.Title, "exe", active.Exe)
		return nil
	}))

	app.Bind("Ctrl+Shift+L", goahk.Func(func(ctx *goahk.Context) error {
		for i := 0; i < 5; i++ {
			if !ctx.Sleep(100 * time.Millisecond) {
				return ctx.Err()
			}
		}
		return nil
	}))
}
