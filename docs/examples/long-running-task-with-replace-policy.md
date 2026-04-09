# long-running-task-with-replace-policy

## Program snippet

```json
{
  "hotkeys": [
    {
      "id": "refresh-index",
      "hotkey": "ctrl+shift+r",
      "policy": { "concurrency": "replace" },
      "steps": [
        { "action": "goahk.callback", "params": { "callback_ref": "refresh" } }
      ]
    }
  ]
}
```

```go
registry.MustRegisterCallback("refresh", func(ctx actions.CallbackContext) error {
    for i := 0; i < 100; i++ {
        if !ctx.Sleep(100 * time.Millisecond) {
            return ctx.Err()
        }
    }
    return nil
})
```

## Expected runtime log sequence

```text
dispatch_startup
policy_replace_admit_latest (run #1)
job_started
policy_replace_cancel_running (run #2 arrives)
policy_replace_admit_latest (run #2)
job_canceled (run #1)
dispatch_trigger_result (run #2 success)
```

## Cancellation behavior notes

- `replace` guarantees newest intent wins for a binding.
- Incoming triggers cancel currently running work for that binding before admitting the latest run.
- Callback code must treat cancellation as expected control flow and return quickly.
