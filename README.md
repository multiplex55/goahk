# goahk

`goahk` is a Windows-first hotkey automation runtime written in Go. It loads JSON configuration, registers global hotkeys for the current user session, and dispatches configured actions when those hotkeys are pressed.

## Current State

- **Maturity:** actively developed, with a stable core runtime shape (`config -> register -> listen -> dispatch -> shutdown`).
- **API stability:** no public library API is promised yet; most runtime code is in `internal/` packages, so integration should be process-based via `cmd/goahk`.
- **Known limitations:**
  - Windows desktop sessions only.
  - Runtime dispatch currently executes direct `steps` plans; flow-linked bindings are compiled but not dispatched by the main runtime path yet.
  - Global hotkeys can fail to register when another app already owns the same key chord.
- **Roadmap pointer:** see architecture and testing docs for ongoing direction and hardening plans (`docs/architecture.md`, `docs/testing.md`).

## What goahk does

At a high level, `goahk` runs as a per-user background process and turns key chords (for example, `Ctrl+Alt+H`) into deterministic action execution.

### Key features

- Global hotkey registration and lifecycle management (startup register, shutdown unregister).
- Config-driven bindings and validation with conflict detection.
- Action callback dispatch through a centralized executor.
- Windows message-loop listener integration for real hotkey events.
- Deterministic “test trigger” workflow via `cmd/goahk-test` for validating binding execution without relying on desktop hotkey delivery.

## How it works (high-level internals)

### 1) Input / listener layer

- `internal/hotkey` defines listener interfaces and a manager that translates platform listener events into binding trigger events.
- `internal/runtime/listener_windows.go` hosts the Windows listener + message-loop path used by `cmd/goahk`.

### 2) Hotkey registry / state management

- `runtime.Bootstrap.Run` loads config, compiles runtime bindings, creates the listener, and registers each binding with `hotkey.Manager`.
- `hotkey.Manager` owns registration IDs and binding mappings, and supports explicit register/unregister lifecycle methods.

### 3) Dispatch / callback execution path

- `runtime.DispatchHotkeyEvents` receives trigger events and looks up the compiled plan for each binding.
- `actions.Executor` executes configured steps and records per-step outcomes.
- `runtime.Bootstrap.Run` drains dispatch results and performs cleanup in shutdown, including unregister and listener close.

### Threading / event-loop assumptions

- The runtime assumes an **interactive Windows user session** and a running Windows message loop for global hotkey delivery.
- Event dispatch runs concurrently with listener/message-loop execution and shuts down through context cancellation.

## Getting started

Keep onboarding short and jump to focused docs:

- **Usage guide:** [`docs/USAGE.md`](docs/USAGE.md)
- **Build/test guide (single source of truth for local + CI workflows):** [`docs/BUILD.md`](docs/BUILD.md)
- **See also:** architecture and constraints in [`docs/architecture.md`](docs/architecture.md)

## API and examples index

| Goal | Entry point | Where to look |
| --- | --- | --- |
| Register global hotkeys | Runtime startup (`goahk -config ...`) -> `hotkey.Manager.Register` | `cmd/goahk/main.go`, `internal/runtime/bootstrap.go`, `internal/hotkey/manager.go` |
| Test/trigger feedback for a binding | Deterministic binding execution via `goahk-test` | `cmd/goahk-test/main.go` |
| Unregister hotkeys | Runtime shutdown -> `unregisterAll` + `hotkey.Manager.Unregister` | `internal/runtime/bootstrap.go`, `internal/hotkey/manager.go` |

## Maintenance notes

- **Contributing/testing pointers:** start with [`docs/testing.md`](docs/testing.md) for unit/integration commands and staged suite expectations.
- **Supported environments:** Windows user desktop sessions are supported for runtime behavior; non-Windows builds are primarily for development/unit testing paths.
- **Non-goals (current):** cross-platform runtime guarantees and a stable exported SDK/library surface.
