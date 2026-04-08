package main

import (
	"context"
	"log"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	app.Bind("Alt+Shift+Left", goahk.Func(func(ctx *goahk.Context) error {
		active, err := ctx.Window.Active()
		if err != nil {
			return err
		}
		if err := ctx.Window.Move(active.HWND, 0, 0); err != nil {
			return err
		}
		if err := ctx.Window.Resize(active.HWND, 960, 1080); err != nil {
			return err
		}
		return nil
	}))
	app.Bind("Alt+Shift+Right", goahk.Func(func(ctx *goahk.Context) error {
		active, err := ctx.Window.Active()
		if err != nil {
			return err
		}
		if err := ctx.Window.Move(active.HWND, 960, 0); err != nil {
			return err
		}
		if err := ctx.Window.Resize(active.HWND, 960, 1080); err != nil {
			return err
		}
		return nil
	}))
	app.Bind("Escape", goahk.Stop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
