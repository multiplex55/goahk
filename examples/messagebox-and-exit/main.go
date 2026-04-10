package main

import (
	"context"
	"log"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	// snippet:start:messagebox-and-exit-main
	app.Bind("F1", goahk.MessageBox("goahk", "Hello from messagebox-and-exit"))
	app.Bind("Escape", goahk.ControlStop())
	// snippet:end:messagebox-and-exit-main

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
