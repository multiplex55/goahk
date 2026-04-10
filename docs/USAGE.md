# goahk practical usage guide

> Platform support: runtime execution is Windows-only for v1.
> On Linux/macOS, `App.Run(...)` returns `goahk.ErrUnsupportedPlatform`.

## 1) Code-first quick start (primary)

Use `goahk.NewApp()` and declare each hotkey in Go.

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
		Bind("Escape", goahk.ControlStop()).
		Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
```

Run from this repository:

```powershell
go run ./examples/basic-script
```

## 2) Custom callback example

Use `goahk.Func` for imperative logic when needed.

```powershell
go run ./examples/custom-callback
```

The callback receives `*goahk.Context`, exposing typed runtime APIs:

- `ctx.Clipboard` (`ReadText`, `WriteText`, `AppendText`, `PrependText`)
- `ctx.Input` (`SendText`, `SendKeys`, `SendChord`, `Paste`)
- `ctx.Window` (`Active`, `List`, `Activate`, `Title`)
- `ctx.Process` (`Launch`, `Open`)
- `ctx.Runtime` (`Stop`, `Sleep`)

## 3) Declarative and callback can be mixed

Built-in action helpers and callbacks can appear in the same `Bind(...)` call.

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

Complete runnable example:

```powershell
go run ./examples/mixed-actions
```

## 4) State model: `Vars` vs shared `AppState`

### `ctx.Vars` (per-trigger)

- Local map for one hotkey execution.
- Good for passing data between steps inside one binding.
- Re-created for each trigger invocation.

### `ctx.AppState` (shared process state)

- Shared across bindings and callback invocations.
- Use for data that must persist across triggers.
- Methods: `Get`, `Set`, `LoadOrStore`.

### Thread-safety guidance

- `ctx.AppState` operations are synchronized and safe across concurrent callback runs.
- `ctx.Vars` should be treated as callback-local scratch state.
- For multi-step shared updates, prefer a single callback section to keep read/modify/write logic coherent.

## 5) Additional code-first examples

- Clipboard transform + paste: `go run ./examples/clipboard-transform-paste`
- Window-aware logic by active title/process: `go run ./examples/window-aware-script`
- Built-in + callback + built-in pipeline: `go run ./examples/mixed-actions`

## 6) Per-binding execution policy

Each binding defaults to `serial` concurrency. Use fluent policy helpers to change it:

```go
app := goahk.NewApp().
	On("Ctrl+Shift+R").Replace().Do(
		goahk.Func(func(ctx *goahk.Context) error {
			for i := 0; i < 30; i++ {
				if !ctx.Runtime.Sleep(100 * time.Millisecond) {
					return ctx.Err()
				}
			}
			return nil
		}),
	).
	On("Ctrl+Shift+Q").QueueOne().Do(goahk.Log("queued"))
```

Helpers:

- `.Serial()` (default)
- `.Replace()`
- `.QueueOne()`
- `.Parallel()`
- `.Drop()`

You can also call `.WithPolicy("replace")` directly; unknown policy names fail fast during `Run()`.

## 7) Logging with `WithLogger`

`WithLogger` attaches a structured runtime logger to all runtime paths. The logger receives:

- bootstrap lifecycle events (`runtime_startup`, `binding_registration_summary`, `runtime_shutdown`)
- dispatch/supervisor events emitted through runtime dispatch (`goahk.dispatch` records)
- compile/validation/load failures (`runtime_validate_failed`, `runtime_compile_failed`, etc.)
- callback and action-context logs (callbacks get the same configured logger through `actions.ActionContext.Logger`)

If you do not call `WithLogger`, goahk uses a default no-op logger.

```go
type stdLogger struct{}

func (stdLogger) Info(msg string, fields map[string]any) {
	log.Printf("%s %#v", msg, fields)
}

func main() {
	app := goahk.NewApp(goahk.WithLogger(stdLogger{}))
	_ = app.Bind("Ctrl+H", goahk.Log("hello")).Run(context.Background())
}
```

## 8) Runtime reliability scenario docs

Operational scenario references (config snippets + expected logs + cancellation notes):

- [`docs/examples/custom-callback-stop-safe.md`](./examples/custom-callback-stop-safe.md)
- [`docs/examples/window-inspect.md`](./examples/window-inspect.md)
- [`docs/examples/move-window.md`](./examples/move-window.md)
- [`docs/examples/mouse-move-and-click.md`](./examples/mouse-move-and-click.md)
- [`docs/examples/long-running-task-with-replace-policy.md`](./examples/long-running-task-with-replace-policy.md)
- [`docs/examples/emergency-stop.md`](./examples/emergency-stop.md)

## 9) Config mode compatibility adapter (`cmd/goahk`)

JSON config mode remains supported for existing deployments/tooling.

```powershell
go run ./cmd/goahk -config .\config.json
```

Use config mode when you need to preserve existing JSON assets; prefer code-first for all new authoring.

Internally, both code-first and config compatibility paths compile through the same canonical entrypoint:
`internal/runtime.CompileRuntimeBindings(program.Program, *actions.Registry)`.

## 10) Compatibility matrix

Guaranteed equivalent behavior between:

- builder path (`goahk.NewApp().Bind(...).Run(...)`)
- JSON adapter path (`internal/config/adapter.go`)

| Capability | Equivalent guarantee |
| --- | --- |
| Hotkey normalization/chord parsing | Yes |
| Linear action `steps` pipelines | Yes |
| Flow references (`flow`) | Yes |
| UIA selector mapping | Yes |
| Runtime compile-time action validation/defaulting | Yes |

Not guaranteed equivalent: arbitrary Go callbacks (`goahk.Func`) and other purely code-level logic.

## 11) Migration boundaries

- New projects should prefer code-first.
- Existing JSON config flows continue to work through the compatibility runner.
- Migration can be incremental per hotkey/flow; parity expectations are limited to features represented in both surfaces.

## 12) See also

- Project overview: [`README.md`](../README.md)
- Architecture: [`docs/architecture.md`](./architecture.md)
- ADR 0001: [`docs/adr/0001-code-first-primary.md`](./adr/0001-code-first-primary.md)
- Build/test workflows: [`docs/BUILD.md`](./BUILD.md)
