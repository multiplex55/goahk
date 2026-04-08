# goahk practical usage guide

## 1) Script mode quickstart (primary)

Use the public `goahk` package to declare bindings in code.

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

Run it from this repo:

```powershell
go run ./examples/basic-script
```

## 2) Config mode compatibility (`cmd/goahk`)

JSON config mode remains supported for existing deployments and tooling.

```powershell
go run ./cmd/goahk -config .\config.json
```

Minimal JSON example:

```json
{
  "hotkeys": [
    {
      "id": "hello-hotkey",
      "hotkey": "Ctrl+Alt+H",
      "steps": [
        {
          "action": "system.message_box",
          "params": {
            "title": "goahk",
            "body": "Hello from Ctrl+Alt+H"
          }
        }
      ]
    }
  ]
}
```

## 3) Migration note: JSON users -> code API

Map existing JSON fields to code calls as follows:

- `hotkeys[].hotkey` -> first argument to `app.Bind(...)`.
- `hotkeys[].steps[]` -> subsequent `goahk` actions in the same `Bind` call.
- `system.message_box` -> `goahk.MessageBox(title, body)`.
- runtime stop action -> `goahk.Stop()`.

Example migration:

- JSON:
  - `"hotkey": "Ctrl+Alt+H"`
  - `"action": "system.message_box"` with title/body params
- Code:

```go
app.Bind("Ctrl+Alt+H", goahk.MessageBox("goahk", "Hello from Ctrl+Alt+H"))
```

## 4) Troubleshooting

- **Conflict on registration:** choose another chord or stop the conflicting app.
- **Invalid hotkey format:** use `<Modifier>+<Key>` (for example `Ctrl+Alt+H`).
- **Action validation errors:** ensure required params are present (for message box, include `body`).

## 5) See also

- Project overview: [`README.md`](../README.md)
- Architecture: [`docs/architecture.md`](./architecture.md)
- Build/test workflows: [`docs/BUILD.md`](./BUILD.md)
