package main

import (
	"context"
	"log"
	"strings"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	app.Bind("Ctrl+Shift+M",
		goahk.ClipboardRead("clipboard"),
		goahk.Func(func(ctx *goahk.Context) error {
			ctx.Vars["clipboard"] = strings.ToUpper(ctx.Vars["clipboard"])
			return nil
		}),
		goahk.ClipboardWrite("{{clipboard}}"),
	)
	app.Bind("Escape", goahk.ControlStop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
