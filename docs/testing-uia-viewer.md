# UIA viewer testing ladder

This ladder defines the minimum vertical slice that must stay healthy in `goahk-uia-viewer` from data acquisition through interaction.

## 6-rung ladder

### 1) Window list

- **Required service calls**
  - `RefreshWindows({ filter, visibleOnly, titleOnly })` for startup and filter changes.
  - `ClearHighlight({})` before loading a new window list.
- **Expected controller/view-state behavior**
  - `windows` view-state is repopulated from service response.
  - `statusText` reports loaded count (for example, `Loaded 1 windows`).
  - Existing highlight is cleared before new rows are shown.
- **Expected failure text when broken**
  - `Failed to refresh windows`.

### 2) Root resolution

- **Required service calls**
  - Optional `ActivateWindow({ hwnd })` when activate-on-select is enabled.
  - `InspectWindow({ hwnd })`.
  - `GetTreeRoot({ hwnd, refresh: true })`.
  - `GetNodeDetails({ nodeID: rootNodeID })` for initial pane hydration.
- **Expected controller/view-state behavior**
  - `selectedWindowID` and `selectedNodeID` move to the chosen window/root.
  - Root node is present in `nodesByID` and path is initialized.
  - Property/pattern/selector panes populate for the root.
- **Expected failure text when broken**
  - `Failed InspectWindow: <reason>`.
  - `Failed GetTreeRoot: <reason>`.
  - `Failed GetNodeDetails: <reason>`.
  - Fallback umbrella status: `Failed to load window`.

### 3) Root details

- **Required service calls**
  - `SelectNode({ nodeID })`.
  - `GetNodeDetails({ nodeID })`.
  - `HighlightNode({ nodeID })`.
- **Expected controller/view-state behavior**
  - `selectedNodeID` changes to the clicked node.
  - `properties`, `patterns`, `selectedPath`, and `selectorText` refresh from details.
  - Status reflects backend detail status text (for example `Loaded node details: Root` / `Details <nodeID>`).
- **Expected failure text when broken**
  - `Failed to select node`.
  - Stage-specific error text such as `Failed GetNodeDetails: boom` must surface when `preferStageFailure` is enabled in status rendering.

### 4) Child expansion

- **Required service calls**
  - `GetNodeChildren({ nodeID })` on first expand.
  - Optional `GetNodeChildren({ nodeID, refresh: true })` to invalidate and reload that branch.
- **Expected controller/view-state behavior**
  - `expandedByID[nodeID]` toggles true/false on row expansion.
  - `childrenByParentID[nodeID]` is filled after initial load.
  - `childrenLoadedByID[nodeID]` prevents duplicate fetches for plain expand/collapse cycles.
- **Expected failure text when broken**
  - `Failed to expand node`.

### 5) Highlight

- **Required service calls**
  - `HighlightNode({ nodeID })` when selection changes.
  - `ClearHighlight({})` before list/root refreshes and teardown paths.
- **Expected controller/view-state behavior**
  - Selected row remains synchronized with overlay/highlight intent.
  - No stale highlight remains after refreshing windows or switching windows.
- **Expected failure text when broken**
  - Selection pipeline failures typically surface as `Failed to select node` (highlight is part of that flow).

### 6) Pattern action

- **Required service calls**
  - `InvokePattern({ nodeID, action, payload? })` from pattern actions UI.
- **Expected controller/view-state behavior**
  - Unsupported actions render disabled.
  - Payload-required actions stay disabled until payload is provided.
  - Success path emits action success feedback and status updates for executed action.
- **Expected failure text when broken**
  - `<ActionLabel> failed` (for example `Invoke failed`).

## Automated suite cross-reference

Coverage for this ladder is split across controller and model tests in the native viewer package:

- `cmd/goahk-uia-viewer/controller_test.go`
- `cmd/goahk-uia-viewer/model_window_table_windows_test.go`
- `cmd/goahk-uia-viewer/model_property_table_windows_test.go`
- `cmd/goahk-uia-viewer/model_pattern_tree_windows_test.go`
- `cmd/goahk-uia-viewer/model_uia_tree_windows_test.go`

## Gating rule

No UI polish task should be marked complete until **rungs 1–4** (window list, root resolution, root details, child expansion) are passing in automated coverage.
