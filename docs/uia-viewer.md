# UIA Viewer

This document describes the `goahk-uia-viewer` desktop app (`cmd/goahk-uia-viewer`) used to inspect UI Automation data and export selector-friendly metadata.

## Architecture

`goahk-uia-viewer` is a Wails application with a Go backend and a React frontend.

- **Backend (Go):** `cmd/goahk-uia-viewer/app.go` exposes methods bound into Wails and emits events for UI updates.
- **Frontend (React/TypeScript):** `cmd/goahk-uia-viewer/frontend/src/App.tsx` renders the viewer panes, invokes backend methods, and reacts to app events.
- **Transport:** Wails runtime bindings (`frontend/src/bindings.ts`) provide typed calls from the frontend to backend methods.

At startup, `main.go` wires the app state and launches Wails. The frontend then requests snapshots and detail payloads via bound methods.

## API contract

The viewer API surface is intentionally small and request/response based.

### Core backend methods

The backend exposes Wails-bound methods on `ViewerApp` in `cmd/goahk-uia-viewer/app.go`.

- `ListWindows`, `RefreshWindows`, `ActivateWindow`: enumerate and focus candidate windows.
- `InspectWindow`: starts inspection context for a chosen top-level window.
- `GetTreeRoot`, `GetNodeChildren`: tree navigation and lazy expansion.
- `SelectNode`, `GetNodeDetails`: selected node state and detailed property payloads.
- `GetFocusedElement`, `GetElementUnderCursor`: focus/cursor-driven discovery.
- `HighlightNode`, `ClearHighlight`: visual feedback overlays for selected elements.
- `CopyBestSelector`: selector/export text generation.
- `GetPatternActions`, `InvokePattern`: discover and execute supported UIA pattern actions.
- `ToggleFollowCursor`: starts/stops cursor-follow polling and emits selection-style events.

### Event contract

Backend may emit Wails events for non-request updates.

- `inspect:follow-cursor`: emitted when cursor-follow selects a new node.
- `inspect:follow-cursor-error`: emitted when cursor-follow polling fails.

Frontend should treat these as advisory updates and keep request methods as source-of-truth for deterministic state.

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

### `wails: command not found`

Install Wails CLI and verify it is on PATH:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails doctor
```

### Viewer builds, but dist folder is empty

Run the build script from repo root and verify `cmd/goahk-uia-viewer/build/bin` exists after `wails build`.

- Windows: `build\build-uia-viewer.bat`
- POSIX: `./build/build-uia-viewer.sh`

Both scripts copy `build/bin` outputs to `dist/goahk-uia-viewer`.

### Frontend dependency issues

From `cmd/goahk-uia-viewer/frontend`:

```bash
npm ci
npm run test
```

Then rerun viewer dev/build scripts.

### UI tree looks stale

Use refresh actions in the viewer and ensure target application window remains available and interactive. If stale state persists, restart the viewer dev session (`build/dev-uia-viewer.*`).
