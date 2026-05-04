# UIA Viewer

This document describes the `goahk-uia-viewer` desktop app (`cmd/goahk-uia-viewer`) used to inspect UI Automation data and export selector-friendly metadata.

## Architecture

`goahk-uia-viewer` is a native Windows desktop application built with Walk. The app is split into a UI shell, a controller layer, and the `internal/inspect` backend services.

- **Walk native app (`cmd/goahk-uia-viewer`):** Owns windows, panes, table/tree models, and user interactions.
- **Controller (`cmd/goahk-uia-viewer/controller.go`):** Coordinates UI events with backend calls, controls refresh/selection flows, and keeps pane state synchronized.
- **Inspection backend (`internal/inspect`):** Provides window listing, UIA tree traversal, element details, selector generation, and highlight/pattern operations via service interfaces.

At startup, the Walk app wires models and event handlers, then delegates inspection actions through the controller into `internal/inspect`-backed services.

## API contract

The viewer API surface is intentionally small and request/response based.

### Core controller/backend calls

The controller drives these core operations through backend service interfaces:

- `RefreshWindows` / `ListWindows` and optional activate paths for top-level window selection.
- `InspectWindow`, `GetTreeRoot`, and `GetNodeChildren` for tree initialization and expansion.
- `SelectNode` and `GetNodeDetails` for property/pattern/selector pane hydration.
- `HighlightNode` and `ClearHighlight` for visual focus feedback.
- `CopyBestSelector` for selector export workflows.
- `GetPatternActions` and `InvokePattern` for supported control-pattern actions.
- `GetFocusedElement` and `GetElementUnderCursor` for focus/cursor-driven discovery.

### State/event behavior

- UI state should be driven by controller-managed model updates.
- Selection changes should update tree, property, pattern, and selector panes as one flow.
- Refresh and window switching should clear stale highlight and stale selection safely.

## Pane responsibilities

Use strict pane boundaries so each panel has one job.

- **Tree pane:** displays hierarchy, expansion state, and node selection.
- **Property pane:** displays selected node attributes (name, class, automation id, control type, patterns, bounds).
- **Selector pane:** previews generated selector snippets and copy actions.
- **Status pane/toolbar:** refresh, timing/status text, and transient diagnostics.

Recommended behavior contracts:

- Selection state is owned by the tree pane and shared to property/selector panes.
- Property and selector panes render empty-state guidance when no node is selected.
- Refresh should preserve user context when possible (same logical node) and gracefully fallback if node no longer exists.

## Troubleshooting

### `go test` on viewer package fails on non-Windows

Some viewer tests are Windows-specific (`*_windows_test.go`). Run general tests on all platforms, and run Windows-specific checks in a Windows environment.

### Viewer binary not found in `dist/goahk-uia-viewer`

Build from repository root using the Go build command documented in `docs/BUILD.md`, then verify the output path exists.

### UI tree appears stale after switching target windows

Use the refresh action, then re-select the target window to force a new root/details fetch path. If stale state persists, restart the viewer process and retry.
