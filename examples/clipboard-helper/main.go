package main

import (
	"context"
	"log"
	"strings"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	// snippet:start:clipboard-helper-main
	app.Bind("Ctrl+Shift+V", goahk.Func(func(ctx *goahk.Context) error {
		text, err := ctx.Clipboard.ReadText()
		if err != nil {
			return err
		}
		transformed := strings.ReplaceAll(text, "foo", "bar")
		return ctx.Input.Paste(transformed)
	}))
	// snippet:end:clipboard-helper-main
	app.Bind("Escape", goahk.ControlStop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
