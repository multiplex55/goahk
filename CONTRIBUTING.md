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
