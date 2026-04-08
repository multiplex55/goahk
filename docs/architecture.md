# Architecture

## Scope

`goahk` v1 is a lightweight Windows runtime centered on a **program model core** with optional adapters:

1. Build a hotkey program (code-first API or config adapter).
2. Apply defaults and validate compiled bindings.
3. Initialize runtime services in deterministic order.
4. Register hotkeys and run a message loop.
5. Shut down gracefully with reverse-order cleanup.

## Explicit v1 non-goals

To keep v1 narrow and reliable, these are explicit non-goals:

- **Windows only**: no Linux/macOS behavior is defined for v1.
- **Single compiled EXE**: no installer-managed multi-process architecture.
- **Per-user background runtime**: no machine-wide Windows service.
- **No custom DSL parser**: avoid adding another language surface in v1.
- **UIA deferred**: advanced UI Automation capabilities are postponed to a future milestone.

## Program model core + adapters

- **Core model:** hotkey bindings and action steps compiled into runtime bindings.
- **Primary adapter:** code-first API (`goahk.NewApp`, `Bind`, `Run`).
- **Compatibility adapter:** JSON config loader used by `cmd/goahk`.

This keeps runtime dispatch independent from how programs are authored, while preserving existing JSON workflows.

## Startup lifecycle

Startup order is fixed:

1. Load/compile program input.
2. Initialize logging.
3. Initialize services.
4. Register hotkeys.
5. Run message loop.

If any stage fails, the runtime cleans up already-initialized resources in reverse order.

## Shutdown lifecycle

Shutdown is context-driven. Callers cancel the run context (for example, from OS signal handling). The runtime exits the message loop and closes registered runtime resources in reverse initialization order.
