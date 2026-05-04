# Build and development workflow guide

This document is the **single source of truth** for local development, test, and build workflows for `goahk`.

Use these commands exactly (copy/paste ready) so local runs and CI behavior stay aligned.

## 1) Environment prerequisites

### Required toolchain/runtime

- Minimum supported Go is `1.22.x` (`go.mod` declares `go 1.22.0`).
- Recommended local development Go is `1.25.x` (latest CI lane).
- Git (required for clone/setup and optional version metadata in build script).
- Windows command shell for `build\build.bat` and `build\check-no-source-binaries.bat` (PowerShell is used internally for UTC timestamp normalization).

Verify toolchain:

```powershell
go version
git --version
```

### Supported OS and platform notes

- **Runtime hotkey behavior is Windows-first** (global hotkey registration/listener path).
- Linux/macOS are valid for unit tests and general development loops.
- Windows integration/hotkey validation requires an interactive desktop user session.

### Windows-specific dependencies/permissions for hotkey behavior

For real hotkey verification on Windows, the process must have:

- Permission to register/unregister global hotkeys in the active user session.
- Access to an interactive desktop (not a headless service session).
- No conflicting application currently owning the same key chord.

### Environment variables used by build tooling

`build\build.bat` supports optional overrides:

- `VERSION` (default: `v0.1.0`)
- `COMMIT` (default: short git SHA or `unknown`)
- `SOURCE_DATE_EPOCH` (default: current Unix epoch)

Example reproducible build metadata override:

```powershell
cmd /c "set VERSION=v0.1.0 && set COMMIT=abcdef0 && set SOURCE_DATE_EPOCH=1700000000 && build\build.bat"
```

## 2) Setup workflow

### Clean clone/setup sequence

```powershell
git clone <REPO_URL> goahk
cd goahk
go mod download
```

If you want an extra dependency integrity check:

```powershell
go mod verify
```

### Offline/locked environment note

In restricted environments, pre-populate the Go module cache from an allowed mirror/artifact source, then run:

```powershell
go mod download
```

If network access is unavailable and cache is incomplete, module resolution will fail with messages like `dial tcp`, `TLS handshake timeout`, or `module lookup disabled`.

## 3) Test commands

### Unit tests (default, all platforms)

```powershell
go test -v ./...
```

Focused deterministic hotkey unit suite (registration/dispatch/unregistration paths without OS hooks):

```powershell
go test -v ./internal/hotkey ./internal/runtime -run 'TestManager|TestDispatchHotkeyEvents|TestParse'
```

Expected outcome:

- `ok` lines for packages.
- Exit code `0`.

Common failure signatures:

- `FAIL` with package path and failing test name.
- `build failed` for compile/type errors.

### Optional race detector (slower)

```powershell
go test -race ./...
```

Expected outcome:

- No `WARNING: DATA RACE` output.
- Exit code `0`.

Common failure signature:

- `WARNING: DATA RACE` followed by stack traces.

### Integration/manual hotkey verification (Windows interactive session)

Automated integration-tagged suites:

```powershell
go test -tags=integration ./internal/runtime ./internal/hotkey
```

Manual runtime hotkey check:

1. Start app with a known config:
   ```powershell
   go run ./cmd/goahk -config <path-to-config.json>
   ```
2. Press configured hotkey once.
3. Confirm callback/effect occurs exactly once.
4. Stop app cleanly (`Ctrl+C`).
5. Restart and verify hotkey can be registered again (no stale registration).

Common hotkey-specific failure signatures:

- Registration conflict/duplicate hotkey errors.
- No callback when pressed (often non-interactive session or chord conflict).

## 4) Build commands

### Core compile/test commands

Run from repository root:

```powershell
go mod download
go build -v ./...
go vet ./...
go test -v ./...
```

### Packaging build script

```powershell
cmd /c build\build.bat
```

### UIA viewer build (Windows-only Walk GUI)

The viewer command (`cmd/goahk-uia-viewer`) is a Windows-only Walk desktop application.

Canonical script path:

```powershell
cmd /c build\build-uia-viewer.bat
```

Direct equivalent:

```powershell
go build -trimpath -v -o dist/goahk-uia-viewer/goahk-uia-viewer.exe ./cmd/goahk-uia-viewer
```

### Output artifact locations

- Main packaged binary output: `dist/goahk`
- UIA viewer binary output: `dist/goahk-uia-viewer/`
- Additional packaging metadata/assets are maintained in `build/`.

## 5) CI-aligned command sequence

Current CI/local build sequence should stay aligned with existing scripts/files:

```powershell
go mod download
cmd /c build\check-no-source-binaries.bat
cmd /c build\build.bat
cmd /c build\build-uia-viewer.bat
go build -v ./...
go vet ./...
go test -v ./...
```

## Related docs

- Testing strategy and staged suites: [`docs/testing.md`](./testing.md)
- Build metadata and packaging context: [`build/README.md`](../build/README.md)
- Runtime usage/configuration: [`docs/USAGE.md`](./USAGE.md)
- UIA viewer architecture and usage: [`docs/uia-viewer.md`](./uia-viewer.md)
