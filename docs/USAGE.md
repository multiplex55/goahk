# goahk practical usage guide

## 1. Purpose and audience

This guide is for developers integrating `goahk` into their own Windows automation workflows (either by running `goahk` as a background process or invoking it from another application).

### Prerequisites

- **OS:** Windows (the runtime is Windows-focused in v1).
- **Toolchain:** Go installed (use the version declared in `go.mod`).
- **Session requirements:** an interactive user desktop session (global hotkeys are registered per-user in the current session).
- **Permissions:** ability to register/unregister global hotkeys in the current user context.

## 2. Quick start

### Minimal setup

```powershell
# from your clone of this repository
cd C:\path\to\goahk

go mod download
```

Create a minimal config file (`config.hello-hotkey.json`):

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
            "body": "Hello from Ctrl+Alt+H",
            "icon": "info",
            "options": "ok"
          }
        }
      ]
    }
  ]
}
```

### Canonical “hello hotkey” flow

```powershell
# 1) Run goahk with your config (registers hotkeys on startup)
go run ./cmd/goahk -config .\config.hello-hotkey.json

# 2) Press Ctrl+Alt+H in any app to trigger the action.
#    You should see a Windows message box.

# 3) Stop goahk with Ctrl+C in the terminal.
#    goahk performs unregister/cleanup during shutdown.
```

## 3. Single cohesive code example covering all core actions

Use this single PowerShell script to demonstrate register, trigger-test, unregister, and clean shutdown behavior end-to-end.

```powershell
# demo-hotkey.ps1
# This script:
#   - writes a real goahk config
#   - starts the runtime (which registers the hotkey)
#   - performs a deterministic trigger test via cmd/goahk-test
#   - lets you manually verify the global hotkey
#   - shuts down cleanly so registrations are removed

$ConfigPath = Join-Path $PWD "config.hello-hotkey.json"

@'
{
  "hotkeys": [
    {
      "id": "hello-hotkey",
      "hotkey": "Ctrl+Alt+H",
      "steps": [
        {
          "action": "system.message_box",
          "params": {
            "title": "goahk Demo",
            "body": "Hello from goahk",
            "icon": "info",
            "options": "ok"
          }
        }
      ]
    }
  ]
}
'@ | Set-Content -Encoding UTF8 $ConfigPath

# (A) Trigger-test pathway: execute the configured binding directly.
#     This validates action wiring and should show the same dialog-box style feedback.
go run ./cmd/goahk-test -config $ConfigPath -binding hello-hotkey

# (B) Start runtime in background. Runtime startup registers Ctrl+Alt+H.
$proc = Start-Process -FilePath "go" `
  -ArgumentList @("run", "./cmd/goahk", "-config", $ConfigPath) `
  -PassThru

Start-Sleep -Seconds 2

Write-Host "Press Ctrl+Alt+H now to verify global hotkey dispatch."
Read-Host "Press Enter after you close the message box"

# (C) Clean shutdown. goahk shutdown path unregisters hotkeys and closes listener resources.
Stop-Process -Id $proc.Id
Wait-Process -Id $proc.Id

Write-Host "goahk stopped. Hotkey registration should now be released."
```

## 4. Embedding into another project

`goahk` does **not** currently expose a stable external library package (runtime packages are under `internal/`), so the practical integration model is:

1. Treat `goahk` as an executable dependency in your app.
2. Generate/manage a `config.json` from your host app.
3. Launch `goahk -config <path>` after your app has prepared config and user session state.
4. On host-app shutdown, terminate `goahk` gracefully so it can unregister hotkeys.

### Typical initialization order in a host app

1. Ensure config file exists and validates.
2. Start `goahk` process.
3. Wait for process readiness (small startup delay or health/log check).
4. Let user trigger hotkeys during normal app lifetime.
5. Stop `goahk` during host shutdown.

### Import/setup patterns

- **Go host app:** use `os/exec` to run `goahk` as a child process.
- **Non-Go app:** spawn the binary with equivalent process APIs and pass `-config`.

### Platform caveats

- Works in Windows user desktop sessions (not intended for Linux/macOS in v1).
- Conflicting global shortcuts from other tools can block registration.
- If your environment enforces security controls around synthetic input/UI actions, configure exceptions as required by your org policy.

## 5. Troubleshooting

### Common errors and exact remediation

- **Error:** `register binding "...": hotkey ... already registered`
  - Another app (or another `goahk` instance) owns the same chord.
  - **Fix:** choose a different `hotkey` value, or stop the conflicting process and restart `goahk`.

- **Error:** `unsupported key "..."` or parse/validation failures for the `hotkey` string.
  - **Fix:** use supported format `<Modifier>+<Modifier>+<Key>` (for example `Ctrl+Alt+H`, `Win+Shift+F12`).

- **Error:** action-specific failures like `system.message_box requires body`.
  - **Fix:** provide required action params in `steps[].params` (for message boxes include `body`, optionally `title`, `icon`, `options`).

- **Symptom:** hotkey never fires.
  - **Fixes:**
    1. Confirm `goahk` is running in the same user desktop session.
    2. Run `go run ./cmd/goahk-test -config <file> -binding <id>` to validate flow/action wiring independent of global hotkey delivery.
    3. Check for conflicting shortcuts and update the configured chord.

## 6. Cross-linking

- Main project overview: [`README.md`](../README.md)
- Build and packaging notes: [`build/README.md`](../build/README.md)
- Testing strategy: [`docs/testing.md`](./testing.md)
