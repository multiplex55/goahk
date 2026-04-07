# Testing strategy

## CI matrix

### 1) Pure unit tests (always)

Run on every push/PR for fast feedback and deterministic behavior.

- `go test ./...`
- `go test -race ./...` (where supported by CI runtime)

These tests must not rely on real desktop/UI state. OS interactions are abstracted behind boundaries and faked in unit tests.

### 2) Windows integration tests (gated/nightly)

Run on Windows workers only, and preferably on nightly or manually-dispatched workflows.

- Exercise real clipboard backend
- Exercise real window enumeration/activation
- Exercise real hotkey listener registration
- Exercise UIA traversal against known fixture apps

Use explicit tags (for example, `-tags=integration`) so normal CI remains fast.

### 3) Manual acceptance checklist

Before release:

1. Launch app with representative config and confirm startup.
2. Verify hotkeys register and trigger expected actions.
3. Validate clipboard read/write/append/prepend behavior.
4. Validate window activation matcher behavior against multiple apps.
5. Validate `goahk-inspect` output in both `text` and `json` formats.
6. Confirm graceful shutdown and resource cleanup.
