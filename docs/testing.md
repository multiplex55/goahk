# Testing strategy

This document defines a staged test plan so we can land deterministic unit coverage first, then add Windows-specific integration coverage as listener/runtime wiring hardens.

## Prerequisites

### Local development (all platforms)

- Go toolchain installed (match `go.mod` version).
- Repository cloned with testdata fixtures available.
- No desktop/UI session is required for pure unit suites.

### Windows integration runner requirements

Use a Windows runner (local machine or CI) with:

- Interactive desktop session available to the test process.
- Permissions to register/unregister global hotkeys.
- Stable keyboard layout for chord parsing/trigger checks.
- A clean environment with no conflicting global hotkey registrations.
- Set `GOAHK_ENABLE_WINDOWS_INPUT_ITEST=1` to enable real keyboard/mouse injection checks in the integration suite.

> Integration tests are intentionally gated with build tags so Linux/macOS CI and quick local loops stay fast.

## Tag and execution conventions

- **Default (no tags):** fast deterministic unit tests.
- **`integration` tag:** longer-running and environment-dependent integration coverage.
- **`windows` + `integration`:** real Windows runtime/listener path verification.

### Local command examples

- Unit-only quick loop:
  - `go test ./...`
- Unit with race detector:
  - `go test -race ./...`
- Targeted package unit runs:
  - `go test ./internal/hotkey ./internal/runtime ./internal/actions`
- Windows integration (from Windows host):
  - `GOAHK_ENABLE_WINDOWS_INPUT_ITEST=1 go test -tags=integration ./internal/runtime ./internal/hotkey`

### CI command examples

- Cross-platform unit job:
  - `go test ./...`
- Optional race job (where supported):
  - `go test -race ./...`
- Windows integration job (gated/manual/nightly):
  - `go test -tags=integration ./internal/runtime ./internal/hotkey`

## Staged suites

## 1) Hotkey manager unit suite (immediate)

Scope:

- lifecycle
- mapping
- cancellation
- listener close
- safe shutdown

Status: **immediate / required in normal CI**.

Current scaffold entrypoint: `internal/hotkey/manager_suite_test.go`.

## 2) Runtime wiring unit suite (immediate)

Scope:

- compile-to-plan mapping
- dispatch behavior
- early failure
- clean stop

Status: **immediate / required in normal CI**.

Current scaffold entrypoint: `internal/runtime/wiring_suite_test.go`.

## 3) Service adapter unit suites (as implemented)

Scope (expand incrementally as adapters are added):

- message box
- clipboard
- process
- window activation

Status: **incremental / required once adapter exists**.

Current scaffold entrypoint: `internal/actions/service_adapter_suite_test.go`.

## 4) Windows integration suite (post-listener)

Scope:

- runtime start
- single hotkey registration
- trigger path verification
- clean shutdown/unregister

Status: **post-listener hardening / gated on Windows + integration tags**.

Current scaffold entrypoint: `internal/runtime/windows_integration_suite_test.go`.

## 5) Manual acceptance checklist

Before release on Windows:

1. Build and launch the `goahk` executable.
2. Trigger configured hotkey once.
3. Verify action path fires exactly once.
4. Exit application cleanly.
5. Restart application and confirm no stale hotkey registration remains.
