# Architecture

## Scope

`goahk` v1 is a lightweight Windows runtime that:

1. Loads strict JSON configuration.
2. Applies defaults and validates hotkey definitions.
3. Initializes runtime services in a deterministic order.
4. Registers hotkeys and runs a message loop.
5. Shuts down gracefully with reverse-order cleanup.

## Explicit v1 non-goals

To keep v1 narrow and reliable, these are explicit non-goals:

- **Windows only**: no Linux/macOS behavior is defined for v1.
- **Single compiled EXE**: no installer-managed multi-process architecture.
- **Per-user background runtime**: no machine-wide Windows service.
- **Config-driven bindings only**: no custom DSL or script parser in v1.
- **UIA deferred**: advanced UI Automation capabilities are postponed to a future milestone.

## Startup lifecycle

Startup order is fixed:

1. Load config.
2. Initialize logging.
3. Initialize services.
4. Register hotkeys.
5. Run message loop.

If any stage fails, the runtime cleans up already-initialized resources in reverse order.

## Shutdown lifecycle

Shutdown is context-driven. Callers cancel the run context (for example, from OS signal handling). The runtime exits the message loop and closes registered runtime resources in reverse initialization order.
