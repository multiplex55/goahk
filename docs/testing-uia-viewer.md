# UIA viewer testing ladder

This ladder defines the minimum vertical slice that must stay healthy in `goahk-uia-viewer` from data acquisition through interaction.

## 6-rung ladder

### 1) Window list

- **Required backend method calls**
  - `RefreshWindows({ filter, visibleOnly, titleOnly })` for startup and filter changes.
  - `ClearHighlight({})` before loading a new window list.
- **Expected UI/store state**
  - `windows` is repopulated from backend response.
  - `statusText` reports loaded count (for example, `Loaded 1 windows`).
  - Existing highlight is cleared before new rows are shown.
- **Expected failure text when broken**
  - `Failed to refresh windows`.

### 2) Root resolution

- **Required backend method calls**
  - Optional `ActivateWindow({ hwnd })` when activate-on-select is enabled.
  - `InspectWindow({ hwnd })`.
  - `GetTreeRoot({ hwnd, refresh: true })`.
  - `GetNodeDetails({ nodeID: rootNodeID })` for initial pane hydration.
- **Expected UI/store state**
  - `selectedWindowID` and `selectedNodeID` move to the chosen window/root.
  - Root node is present in `nodesByID` and path is initialized.
  - Property/pattern/selector panes populate for the root.
- **Expected failure text when broken**
  - `Failed InspectWindow: <reason>`.
  - `Failed GetTreeRoot: <reason>`.
  - `Failed GetNodeDetails: <reason>`.
  - Fallback umbrella status: `Failed to load window`.

### 3) Root details

- **Required backend method calls**
  - `SelectNode({ nodeID })`.
  - `GetNodeDetails({ nodeID })`.
  - `HighlightNode({ nodeID })`.
- **Expected UI/store state**
  - `selectedNodeID` changes to the clicked node.
  - `properties`, `patterns`, `selectedPath`, and `selectorText` refresh from details.
  - Status reflects backend detail status text (for example `Loaded node details: Root` / `Details <nodeID>`).
- **Expected failure text when broken**
  - `Failed to select node`.
  - Stage-specific error text such as `Failed GetNodeDetails: boom` must surface when `preferStageFailure` is enabled in status rendering.

### 4) Child expansion

- **Required backend method calls**
  - `GetNodeChildren({ nodeID })` on first expand.
  - Optional `GetNodeChildren({ nodeID, refresh: true })` to invalidate and reload that branch.
- **Expected UI/store state**
  - `expandedByID[nodeID]` toggles true/false on row expansion.
  - `childrenByParentID[nodeID]` is filled after initial load.
  - `childrenLoadedByID[nodeID]` prevents duplicate fetches for plain expand/collapse cycles.
- **Expected failure text when broken**
  - `Failed to expand node`.

### 5) Highlight

- **Required backend method calls**
  - `HighlightNode({ nodeID })` when selection changes.
  - `ClearHighlight({})` before list/root refreshes and teardown paths.
- **Expected UI/store state**
  - Selected row remains synchronized with overlay/highlight intent.
  - No stale highlight remains after refreshing windows or switching windows.
- **Expected failure text when broken**
  - Selection pipeline failures typically surface as `Failed to select node` (highlight is part of that flow).

### 6) Pattern action

- **Required backend method calls**
  - `InvokePattern({ nodeID, action, payload? })` from `PatternPanel`.
- **Expected UI/store state**
  - Unsupported actions render disabled.
  - Payload-required actions stay disabled until payload is provided.
  - Success path emits action success feedback and store status updates for executed action.
- **Expected failure text when broken**
  - `<ActionLabel> failed` (for example `Invoke failed`).

## Automated suite cross-reference

Coverage for this ladder is split across Go service tests and frontend store/component tests:

- **Go backend**
  - `internal/inspect/service_windows_test.go`
- **TypeScript store/components**
  - `cmd/goahk-uia-viewer/frontend/src/store/inspectStore.test.ts`
  - `cmd/goahk-uia-viewer/frontend/src/components/TreePane.test.tsx`
  - `cmd/goahk-uia-viewer/frontend/src/components/PatternPanel.test.tsx`
  - `cmd/goahk-uia-viewer/frontend/src/components/StatusBar.test.tsx`

## Gating rule

No UI polish task should be marked complete until **rungs 1–4** (window list, root resolution, root details, child expansion) are passing in automated coverage.
