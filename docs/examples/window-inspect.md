# window-inspect

## CLI inspect quick checks

```bash
goahk-inspect window active
goahk-inspect --format json window list
```

`window active` and `window list` now include parse-friendly geometry/state fields:

- `Rect` (bounds in `left,top,right,bottom` form for text output; object in JSON)
- `Visible`, `Minimized`, `Maximized`
- `ProcessPath`, `ProcessPathStatus`, and `ProcessPathError` for permission-safe process path diagnostics.

## Program snippet

```go
app := goahk.NewApp()
app.Bind("Ctrl+Shift+I", goahk.Func(func(ctx *goahk.Context) error {
    active, err := ctx.Window.Active()
    if err != nil {
        return err
    }
    ctx.Log("active window", "title", active.Title, "exe", active.Exe)
    return nil
}))
app.Bind("Escape", goahk.ControlStop())
```

## Expected runtime log sequence

```text
dispatch_startup
policy_serial_admit
job_started (window inspect)
dispatch_trigger_result (success)
```

## Cancellation behavior notes

- Inspection is expected to be short-lived; cancellation should normally not be required.
- If the runtime is stopping during inspection, the callback should return promptly when the context is canceled.
- Failures to query the active window should produce `dispatch_failure_detail` with the failing action name.
