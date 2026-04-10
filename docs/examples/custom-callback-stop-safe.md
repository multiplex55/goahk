# custom-callback-stop-safe

## Program snippet

```go
app := goahk.NewApp()
app.Bind("Ctrl+Shift+R", goahk.Func(func(ctx *goahk.Context) error {
    for {
        select {
        case <-ctx.Context.Done():
            ctx.Logger.Info("callback canceled cleanly")
            return ctx.Context.Err()
        default:
            if !ctx.Sleep(150 * time.Millisecond) {
                return ctx.Context.Err()
            }
            ctx.Logger.Info("heartbeat")
        }
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

- The callback must observe `ctx.Context.Done()` quickly and return.
- A graceful stop (`Escape`) cancels in-flight callback work; no force-termination should be required for a responsive callback.
- Returning `ctx.Context.Err()` records a clear cancellation reason in diagnostics.
