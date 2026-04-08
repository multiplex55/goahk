package main

import (
	"context"
	"log"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	app.Bind("Ctrl+Shift+C", goahk.Func(func(ctx *goahk.Context) error {
		text, err := ctx.Clipboard.ReadText()
		if err != nil {
			return err
		}
		return ctx.Clipboard.WriteText("callback captured: " + text)
	}))
	app.Bind("Escape", goahk.Stop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
