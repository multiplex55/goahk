# UIA viewer migration: Walk shell adoption

Date: 2026-05-04

## Summary

The UIA viewer desktop shell has been migrated away from the Wails/React/TypeScript stack to a Windows-native Walk app in `cmd/goahk-uia-viewer`.

## What was removed

- Wails module/runtime dependencies from the UIA viewer shell.
- React/TypeScript frontend shell and associated web-view integration surface.

## What was preserved

- `internal/inspect` remains the backend owner for inspection behavior, including:
  - window enumeration
  - root/tree/node inspection and details resolution
  - highlight lifecycle
  - pattern discovery and invocation

## What was added

- Walk presentation/controller shell in `cmd/goahk-uia-viewer`.
- Controller/model regression tests covering the inspection ladder (window list, root resolution, details, expansion, highlight flow, and pattern dispatch).
- Repo hygiene guard to fail tests if Wails module references reappear in `go.mod` or `go.sum`.
