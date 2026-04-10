# Contributing to goahk

Thanks for contributing.

## Local toolchain expectations

- Minimum supported Go version: **1.22.x**.
- Recommended local version for day-to-day development: **1.25.x** (the latest CI lane).
- Keep `.github/workflows/go.yml` explicit and version-pinned (no `stable`).
- Keep workflow/toolchain policy aligned with `go.mod`.

Quick local verification:

```powershell
go version
go mod download
go build -v ./...
go vet ./...
go test -v ./...
```

## CI policy

CI runs a Go matrix for minimum and latest supported versions. A repository hygiene test enforces version consistency between `go.mod` and workflow configuration.


## Release artifact pull requests

- Default policy: compiled binaries (`.exe`, `.dll`, `.so`, `.dylib`) are blocked from source PRs.
- Exception policy: binaries may be committed only for dedicated release artifact PRs that:
  - use a branch name prefixed with `release-artifacts/`, and
  - include the `release-artifacts` PR label.
- Approved location for release binaries is `dist/releases/` only.
- Automation toggle: CI sets `GOAHK_ALLOW_RELEASE_ARTIFACTS=1` only when the `release-artifacts` label is present, so checks remain deterministic.
