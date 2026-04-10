# custom-callback-stop-safe

## Program snippet

```go
app := goahk.NewApp()
app.Bind("Ctrl+Shift+R", goahk.Func(func(ctx *goahk.Context) error {
    for {
        if err := ctx.Err(); err != nil {
            ctx.Log("callback canceled cleanly", "err", err)
            return err
        }
        if !ctx.Sleep(150 * time.Millisecond) {
            return ctx.Err()
        }
        ctx.Log("heartbeat", "binding", ctx.Binding(), "trigger", ctx.Trigger())
    }
}))
app.Bind("Escape", goahk.ControlStop())
```

## Expected runtime log sequence

```text
dispatch_startup
policy_serial_admit
job_started (Ctrl+Shift+R callback)
control_command_received (Escape => stop)
job_canceled
dispatch_trigger_result (context canceled)
```

## Cancellation behavior notes

- The callback should prefer `ctx.Err()`/`ctx.Sleep(...)` over direct channel selects for cancellation checks.
- A graceful stop (`Escape`) cancels in-flight callback work; no force-termination should be required for a responsive callback.
- Returning `ctx.Err()` records a clear cancellation reason in diagnostics.
