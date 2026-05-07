# UIA Viewer

This document describes the `goahk-uia-viewer` desktop app (`cmd/goahk-uia-viewer`) used to inspect UI Automation data and export selector-friendly metadata.

## Architecture

`goahk-uia-viewer` is a native Windows desktop application built with Walk. The active architecture is controller/service driven and intentionally keeps runtime inspection logic outside the UI shell.

- **Command package (`cmd/goahk-uia-viewer`)**: owns Walk controls, layout composition, event wiring, and UI-facing models.
- **Controller (`cmd/goahk-uia-viewer/controller.go`)**: owns workflow state and orchestration across refresh/select/highlight/invoke flows.
- **Inspection backend (`internal/inspect` and adapters)**: owns backend inspection, node/pattern invocation, selector production, and highlight responsibilities.

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



### Status shows `backend=synthetic`

`backend=synthetic` means the viewer is presenting a synthesized compatibility tree instead of a native UIA provider tree. This is useful for diagnostics, but it is not a UIA parity pass.

### How UIA-only failures should appear

When UIA calls fail, status text should keep stage context (for example `Failed GetTreeRoot: ...` or `Failed GetNodeDetails: ...`). If fallback succeeds, the status should explicitly indicate fallback/degraded mode so failures are triaged as UIA-path issues rather than total viewer failure.

### What HWND fallback implies for parity expectations

`HWND_TREE` fallback means the viewer is using window-handle hierarchy data. This mode is intentionally degraded for parity comparisons: selector and structure differences from UIA are expected, so treat results as compatibility diagnostics only.

### Manual verification vs deterministic tests

UIA viewer parity checks on Notepad are manual Windows desktop verification. Keep deterministic fake-provider tests as the baseline regression suite, and use manual runs as a complementary signal for desktop integration behavior.

## UIA parity criteria

A run is parity-valid only when diagnostics/reporting indicates `Provider=uia`, `Mode=UIA_TREE`, and `Fallback=No`. Any `HWND_TREE` result is degraded-by-design and non-parity.
