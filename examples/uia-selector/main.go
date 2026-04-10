package main

import (
	"context"
	"log"
	"time"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	app.Bind("Ctrl+Shift+F", goahk.Func(func(ctx *goahk.Context) error {
		sel := goahk.SelectByAutomationID("searchBox").WithAncestors(
			goahk.SelectByName("Settings").WithControlType("Window"),
		)
		if _, err := ctx.UIA.Invoke(sel, 2*time.Second, 100*time.Millisecond); err != nil {
			return err
		}
		return ctx.Input.SendText("goahk")
	}))
	app.Bind("Escape", goahk.Stop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
