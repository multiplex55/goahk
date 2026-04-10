package main

import (
	"context"
	"log"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	// snippet:start:basic-script-main
	app.Bind("Ctrl+Alt+B", goahk.SendText("basic script trigger"))
	// snippet:end:basic-script-main
	app.Bind("Escape", goahk.ControlStop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
