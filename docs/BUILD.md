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

### Debug/development build (matches current CI intent)

```powershell
go build -v ./...
```

This validates that all packages compile.

### Production/release build (project packaging script)

```powershell
build\build.bat
```

### Output artifact locations

- Main packaged binary output: `dist/goahk`
- Additional packaging metadata/assets are maintained in `build/`.

## 5) Validation checklist

After building, run this smoke checklist:

- [ ] Binary/app starts successfully.
- [ ] Hotkey can be registered.
- [ ] Hotkey callback fires when chord is pressed.
- [ ] App exits cleanly.
- [ ] Hotkey can be registered again after restart (unregister worked).

Platform-specific notes:

- **Windows:** full checklist is applicable, including real hotkey registration/callback.
- **Linux/macOS:** compile/unit-test validation is expected; real Windows global hotkey behavior is not authoritative.

## 6) CI alignment

This guide is the **single source of truth** for the command set used locally and in CI.

Current CI workflow (`.github/workflows/go.yml`) runs on an explicit Go matrix (`1.22.x` and `1.25.x`) and executes:

```powershell
go mod download
build\check-no-source-binaries.bat
go build -v ./...
go vet ./...
go test -v ./...
```

Recommended local pre-PR command sequence (aligned to CI):

```powershell
go mod download
build\check-no-source-binaries.bat
go build -v ./...
go vet ./...
go test -v ./...
```

If CI evolves, update this document first (or in the same PR) so contributors and automation remain synchronized.

## Related docs

- Testing strategy and staged suites: [`docs/testing.md`](./testing.md)
- Build metadata and packaging context: [`build/README.md`](../build/README.md)
- Runtime usage/configuration: [`docs/USAGE.md`](./USAGE.md)
