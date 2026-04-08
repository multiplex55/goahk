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

## 5) `system.open` examples

You can open websites, folders, or an executable via the `system.open` action.

Website:

```json
{
  "action": "system.open",
  "params": {
    "target": "www.chatgpt.com",
    "kind": "url"
  }
}
```

`www.chatgpt.com` is normalized to `https://www.chatgpt.com`. Invalid URLs return a validation error.

Folder:

```json
{
  "action": "system.open",
  "params": {
    "target": "C:\\\\testing",
    "kind": "folder"
  }
}
```

Folder paths must be absolute and exist on disk. Unknown/missing paths return an error.

Application:

```json
{
  "action": "system.open",
  "params": {
    "target": "C:\\\\Windows\\\\System32\\\\notepad.exe",
    "kind": "application",
    "args": "notes.txt",
    "working_dir": "C:\\\\Temp",
    "env": "A=1;B=2"
  }
}
```

`kind=application` requires an absolute `.exe` path. Relative or non-`.exe` targets are rejected.

## 6) `window.list_open_applications` examples

Collect open apps into metadata:

```json
{
  "action": "window.list_open_applications",
  "params": {
    "save_as": "open_apps"
  }
}
```

This saves JSON to `metadata.open_apps` using stable fields:
`title`, `exe`, `pid`, `hwnd`, `class`, `active`.

Include inactive/background windows and disable process-level dedupe:

```json
{
  "action": "window.list_open_applications",
  "params": {
    "save_as": "open_windows",
    "include_background": "true",
    "dedupe_by": "window"
  }
}
```

Follow-up action example (for logging/inspection):

```json
{
  "action": "system.log",
  "params": {
    "message": "Collected open app inventory in metadata.open_apps"
  }
}
```

## 7) `window.list_open_folders` examples

Collect open File Explorer folders into metadata:

```json
{
  "action": "window.list_open_folders",
  "params": {
    "save_as": "open_folders"
  }
}
```

This saves JSON to `metadata.open_folders` with objects shaped like:

- `path` (string, resolved filesystem path for the folder tab/window)
- `title` (string, Explorer title/location label)
- `pid` (number, Explorer window process id)
- `hwnd` (string, hex window handle)
- `diagnostic` (optional string; emitted only for entries where path resolution failed)

Optional dedupe by folder path:

```json
{
  "action": "window.list_open_folders",
  "params": {
    "save_as": "open_folders",
    "dedupe": "true"
  }
}
```

If no open folders are found, the action stores `[]` (JSON array) in the metadata key.

Platform behavior:

- Windows: supported (uses Explorer/Shell enumeration and resolves real folder paths).
- Non-Windows: currently stubbed/unsupported and returns an action error.

## 8) See also

- Project overview: [`README.md`](../README.md)
- Architecture: [`docs/architecture.md`](./architecture.md)
- Build/test workflows: [`docs/BUILD.md`](./BUILD.md)
