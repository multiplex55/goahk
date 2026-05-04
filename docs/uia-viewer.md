# UIA Viewer

This document describes the `goahk-uia-viewer` desktop app (`cmd/goahk-uia-viewer`) used to inspect UI Automation data and export selector-friendly metadata.

## Architecture

`goahk-uia-viewer` is a native Windows desktop application built with Walk. The active architecture is controller/service driven and intentionally keeps runtime inspection logic outside the UI shell.

- **Walk shell (`cmd/goahk-uia-viewer`)**: owns native windows, pane widgets, table/tree models, and event wiring.
- **Controller (`cmd/goahk-uia-viewer/controller.go`)**: orchestrates user intent, service calls, view-state updates, refresh transitions, and error/status reporting.
- **Inspection services (`internal/inspect` and related adapters)**: provide window listing, UIA tree traversal, node detail hydration, selector generation, highlight operations, and pattern invocation.

Startup flow:

1. Walk shell builds pane models and binds UI events.
2. Controller receives intent from UI events and coordinates service operations.
3. Controller updates shared view-state and pushes results into pane models.

## API contract

The viewer interface between controller and backend services is request/response based with explicit operations:

- `RefreshWindows` / `ListWindows` and optional activation paths for top-level window selection.
- `InspectWindow`, `GetTreeRoot`, and `GetNodeChildren` for tree initialization and expansion.
- `SelectNode` and `GetNodeDetails` for property, pattern, and selector pane hydration.
- `HighlightNode` and `ClearHighlight` for visual focus feedback.
- `CopyBestSelector` for selector export workflows.
- `GetPatternActions` and `InvokePattern` for supported control-pattern actions.
- `GetFocusedElement` and `GetElementUnderCursor` for focus/cursor-driven discovery.

Controller/state behavior contracts:

- Selection updates tree + property + selector panes as one transaction.
- Window refresh/switch paths clear stale highlight and stale selection safely.
- Status messages should preserve actionable stage context (`RefreshWindows`, `GetTreeRoot`, `GetNodeDetails`, etc.).

## Pane responsibilities

Each pane has one primary responsibility:

- **Tree pane**: hierarchy presentation, expansion state, selected node ownership.
- **Property pane**: selected-node metadata (name, class, automation id, control type, bounds, pattern list).
- **Selector pane**: generated selector preview and copy/export interactions.
- **Pattern/actions pane**: enabled/disabled action state and invocation controls for supported patterns.
- **Status/toolbar pane**: refresh, filter controls, short diagnostics, timing/status text.

Recommended behavior contracts:

- Tree selection is the source of truth for node-focused panes.
- Property/selector/pattern panes render empty guidance when no node is selected.
- Refresh attempts to preserve logical user position; if unavailable, fallback to root safely.

## Troubleshooting

### `go test` on viewer package fails on non-Windows

Some viewer tests are Windows-specific (`*_windows_test.go`). Run general tests on all platforms, and run Windows-specific checks in a Windows environment.

### Viewer binary not found in `dist/goahk-uia-viewer`

Build from repository root using the Go build command documented in `docs/BUILD.md`, then verify the output path exists.

### UI tree appears stale after switching target windows

Use the refresh action, then re-select the target window to force a new root/details fetch path. If stale state persists, restart the viewer process and retry.
