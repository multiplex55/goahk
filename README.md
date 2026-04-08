# goahk

`goahk` is a Windows-first hotkey automation runtime written in Go. The core model is **code-first** (`goahk.NewApp()` + `Bind(...)`), and the JSON config runner remains available as a compatibility adapter via `cmd/goahk`.

## Quick start (code-first, primary)

```go
package main

import (
	"context"
	"log"

	"goahk/goahk"
)

func main() {
	app := goahk.NewApp()
	app.Bind("1", goahk.MessageBox("goahk", "You pressed 1"))
	app.Bind("Escape", goahk.Stop())

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

Runnable examples: [`examples/basic-script/main.go`](examples/basic-script/main.go) and [`examples/standalone-script/main.go`](examples/standalone-script/main.go).

## JSON mode (optional compatibility path)

If you already have JSON configs, you can continue using:

```powershell
go run ./cmd/goahk -config .\config.json
```

`cmd/goahk/main.go` intentionally remains the config-runner entry point for compatibility while code-first usage becomes the primary documented API.

## Migration note: JSON bindings -> code API

Map each JSON hotkey object to one `Bind(...)` call:

- JSON `hotkeys[].hotkey` -> `app.Bind("<hotkey>", ...)`
- JSON `hotkeys[].steps[].action` -> corresponding `goahk` action helper (`MessageBox`, `SendText`, `Launch`, etc.)
- JSON stop-style behavior -> `goahk.Stop()`

Example mapping:

```json
{
  "id": "hello-hotkey",
  "hotkey": "Ctrl+Alt+H",
  "steps": [
    {
      "action": "system.message_box",
      "params": { "title": "goahk", "body": "Hello" }
    }
  ]
}
```

becomes:

```go
app.Bind("Ctrl+Alt+H", goahk.MessageBox("goahk", "Hello"))
```

## Documentation index

- Usage guide: [`docs/USAGE.md`](docs/USAGE.md)
- Architecture: [`docs/architecture.md`](docs/architecture.md)
- Build/test guide: [`docs/BUILD.md`](docs/BUILD.md)
- Testing strategy: [`docs/testing.md`](docs/testing.md)
