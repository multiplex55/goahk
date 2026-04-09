# emergency-stop

## Program snippet

```json
{
  "hotkeys": [
    { "id": "graceful", "hotkey": "esc", "steps": [{ "action": "system.log", "params": { "message": "graceful" } }] },
    { "id": "hard", "hotkey": "shift+esc", "steps": [{ "action": "system.log", "params": { "message": "hard" } }] }
  ]
}
```

> Runtime compilation maps `esc` to control command `stop` and `shift+esc` to `hard_stop`.

## Expected runtime log sequence

```text
dispatch_startup
control_command_received (Escape => stop)
control_command_received (Shift+Escape => hard_stop)
```

## Cancellation behavior notes

- `Escape` is a graceful stop: allow in-flight jobs to observe cancellation and exit.
- `Shift+Escape` is a hard-stop intent and should force termination when grace timeout is exceeded.
- Keep both bindings available in long-running automation profiles to preserve operator control.
