package main

import (
	"context"
	"log"
	"strings"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	app.Bind("Ctrl+Shift+V", goahk.Func(func(ctx *goahk.Context) error {
		text, err := ctx.Clipboard.ReadText()
		if err != nil {
			return err
		}
		replaced := strings.ReplaceAll(text, "foo", "bar")
		return ctx.Input.Paste(replaced)
	}))
	app.Bind("Escape", goahk.Stop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
