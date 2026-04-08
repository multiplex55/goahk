# goahk

`goahk` is a Windows-first hotkey automation runtime written in Go.

## Quick start (code-first API, primary)

Use `goahk.NewApp()` and `Bind(...)` directly from Go.

```go
package main

import (
	"context"
	"log"

	"goahk/goahk"
)

func main() {
	if err := goahk.NewApp().
		Bind("1", goahk.MessageBox("goahk", "You pressed 1")).
		Bind("Escape", goahk.Stop()).
		Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

Runnable examples:

- [`examples/basic-script/main.go`](examples/basic-script/main.go)
- [`examples/custom-callback/main.go`](examples/custom-callback/main.go)
- [`examples/clipboard-transform-paste/main.go`](examples/clipboard-transform-paste/main.go)
- [`examples/window-aware-script/main.go`](examples/window-aware-script/main.go)
- [`examples/mixed-actions/main.go`](examples/mixed-actions/main.go)

## Declarative + callback can be mixed

You can chain built-in declarative actions and custom callback logic in one binding.

```go
app.Bind("Ctrl+Shift+M",
	goahk.ClipboardRead("clipboard"),
	goahk.Func(func(ctx *goahk.Context) error {
		ctx.Vars["clipboard"] = strings.ToUpper(ctx.Vars["clipboard"])
		return nil
	}),
	goahk.ClipboardWrite("{{clipboard}}"),
)
```

See: [`examples/mixed-actions/main.go`](examples/mixed-actions/main.go).

## Compatibility mode (JSON adapter)

- Code-first scripts are the primary and recommended API.
- JSON config mode remains available via `cmd/goahk` as a compatibility adapter.
- Existing declarative usage (action helpers like `MessageBox`, `ClipboardRead`, `SendKeys`, etc.) remains supported.

JSON compatibility runner:

```powershell
go run ./cmd/goahk -config .\config.json
```

## Compatibility matrix (guaranteed equivalent behavior)

The following are guaranteed equivalent between:

- goahk builder path (`goahk.NewApp().Bind(...).Run(...)`)
- JSON adapter path (`internal/config/adapter.go` via `cmd/goahk`)

| Capability | Builder path | JSON adapter path | Guarantee |
| --- | --- | --- | --- |
| Hotkey chord parsing/normalization | ✅ | ✅ | Equivalent chord registration semantics |
| Linear action pipelines (`steps`) | ✅ | ✅ | Same action order and params are compiled |
| Flow references (`flow`) | ✅ | ✅ | Same flow lookup and compiled plan behavior |
| UIA selector definitions | ✅ | ✅ | Same selector model reaches runtime |
| Runtime defaults/validation | ✅ | ✅ | Same runtime compiler + registry validation |

Out of scope for parity guarantees: arbitrary Go callback logic (`goahk.Func`) because JSON cannot encode executable Go code.

## State model (`Vars` vs `AppState`)

- `ctx.Vars` is per-trigger state (per execution of a hotkey callback). Treat it as local scratch data for that single run.
- `ctx.AppState` is process-wide shared state across all callbacks and hotkey executions.

Thread-safety guidance:

- `ctx.Vars` does not need cross-goroutine synchronization if you keep it inside one callback execution.
- `ctx.AppState` is safe for concurrent access (`Get` / `Set` / `LoadOrStore` are synchronized), and is the right place for shared counters/flags/cache values.
- If you store composite values that require read-modify-write semantics across multiple operations, keep updates in one callback section or encode as single string values with `LoadOrStore`/`Set` boundaries.

## Documentation index

- ADRs: [`docs/adr/`](docs/adr/)
- Usage guide: [`docs/USAGE.md`](docs/USAGE.md)
- Architecture: [`docs/architecture.md`](docs/architecture.md)
- Build/test guide: [`docs/BUILD.md`](docs/BUILD.md)
- Testing strategy: [`docs/testing.md`](docs/testing.md)
