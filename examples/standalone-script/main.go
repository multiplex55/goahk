package main

import (
	"context"
	"log"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()

	// Press "1" to show a MessageBox. The process stays running until Escape is pressed.
	app.Bind("1", goahk.MessageBox("goahk", "You pressed 1"))
	// Press "2" to snapshot open applications into metadata key "open_apps".
	app.Bind("2", goahk.ListOpenApplications("open_apps"), goahk.Log("Captured open app inventory to metadata.open_apps"))
	app.Bind("Escape", goahk.ControlStop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
