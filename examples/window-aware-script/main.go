package main

import (
	"context"
	"log"
	"strings"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	app.Bind("Ctrl+Shift+W", goahk.Func(func(ctx *goahk.Context) error {
		active, err := ctx.Window.Active()
		if err != nil {
			return err
		}

		title := strings.ToLower(active.Title)
		exe := strings.ToLower(active.Exe)

		switch {
		case strings.Contains(exe, "code") || strings.Contains(title, "visual studio"):
			return ctx.Input.SendText("// goahk: editor-focused action")
		case strings.Contains(exe, "chrome") || strings.Contains(exe, "msedge") || strings.Contains(title, "firefox"):
			return ctx.Input.SendText("https://github.com/")
		default:
			return ctx.Input.SendText("Active window: " + active.Title)
		}
	}))
	app.Bind("Escape", goahk.Stop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
