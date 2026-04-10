# mouse-move-and-click

## Program snippet

```go
app := goahk.NewApp()
app.Bind("Ctrl+Alt+M",
    goahk.MouseMoveAbsolute(1200, 480),
    goahk.Sleep(40),
    goahk.MouseClick("left"),
)
app.Bind("Escape", goahk.ControlStop())
```

## Expected runtime log sequence

```text
dispatch_startup
policy_serial_admit
job_started (input.mouse_move_absolute -> flow.sleep -> input.mouse_click)
dispatch_trigger_result (success)
```

## Cancellation behavior notes

- Cancellation between steps should prevent later input actions from running.
- Add a short `Sleep` only when required for UI stability; keep it minimal.
- For safety-critical automations, prefer foreground-window checks before clicking.
