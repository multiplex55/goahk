package main

import (
	"context"
	"log"
	"strings"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	// snippet:start:window-aware-script-main
	app.Bind("Ctrl+Shift+W", goahk.Func(func(ctx *goahk.Context) error {
		if err := ctx.Window.ActivateMatch(goahk.MatchTitleContains("code")); err != nil {
			return err
		}
		active, err := ctx.Window.Active()
		if err != nil {
			return err
		}

		title := strings.ToLower(active.Title)
		exe := strings.ToLower(active.Exe)

		switch {
		case strings.Contains(exe, "code") || strings.Contains(title, "visual studio"):
			return ctx.Input.SendText("// editor mode")
		case strings.Contains(exe, "chrome") || strings.Contains(exe, "msedge") || strings.Contains(title, "firefox"):
			return ctx.Input.SendText("https://github.com/")
		default:
			return ctx.Input.SendText("Window: " + active.Title)
		}
	}))
	// snippet:end:window-aware-script-main
	app.Bind("Escape", goahk.ControlStop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
