# move-window

## Program snippet

```go
app := goahk.NewApp()
app.Bind("Alt+Shift+Left", goahk.Func(func(ctx *goahk.Context) error {
    active, err := ctx.Window.Active()
    if err != nil {
        return err
    }
    if err := ctx.Window.Move(active.HWND, 0, 0); err != nil {
        return err
    }
    return ctx.Window.Resize(active.HWND, 960, 1080)
}))
app.Bind("Escape", goahk.Stop())
```

## Expected runtime log sequence

```text
dispatch_startup
policy_serial_admit
job_started (move-window)
dispatch_trigger_result (success)
```

## Cancellation behavior notes

- Move/resize operations should be idempotent and quick.
- If canceled mid-run, stop additional window mutations and return immediately.
- Use a single callback step to keep read/move/resize coherent for one active window snapshot.
