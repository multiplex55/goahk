# goahk

`goahk` is a Windows-focused hotkey automation runtime implemented in Go. It targets a small, dependable v1 that can run as a per-user background process and execute configurable hotkey workflows.

## v1 goals

- Compile to a single Windows executable.
- Run as a per-user background runtime.
- Configure behavior through file-based bindings.
- Provide deterministic startup and shutdown lifecycle handling.

## Explicit v1 non-goals

The following are intentionally **out of scope** for v1:

- Cross-platform support (v1 is **Windows only**).
- Multi-binary distribution; v1 ships as a **single compiled EXE**.
- System-wide/service runtime; v1 supports a **per-user background runtime** only.
- Authoring custom scripting languages; v1 uses **config-driven bindings only (no custom DSL)**.
- Deep UI Automation support; **UIA is deferred to a later milestone**.

## Current layout

- `cmd/goahk`: application entrypoint.
- `cmd/goahk-test`: optional fixture runner to execute a single configured binding.
- `internal/app`: runtime bootstrap and lifecycle orchestration.
- `internal/tray`: tray/status command routing.
- `internal/startup`: startup-on-login install/uninstall command helpers.
- `internal/config`: configuration schema, defaults, loading, and validation.
- `build`: packaging manifest, icon/version metadata, and reproducible build script (`build/README.md`).
- `testdata/config`: sample configuration files used by tests.
- `docs/architecture.md`: architectural decisions and constraints.
- `docs/startup-on-login.md`: startup install/uninstall behavior.
- `docs/USAGE.md`: practical setup, hello-hotkey flow, embedding, and troubleshooting.
