# UIA viewer testing ladder

This document defines a **tooling validation contract** for `goahk-uia-viewer`. The runtime and script-first Go API remain the product center; this viewer guidance helps diagnose UIA inspection behavior without changing core automation direction.

## Parity contract: `UIA_TREE` vs fallback trees

### Required parity mode

- `UIA_TREE` is the required mode for `UIATreeInspector.ahk`-style parity.
- A parity pass requires:
  - `GetTreeRoot(..., Mode: UIA_TREE)` completes successfully.
  - `resp.State.ActiveMode == UIA_TREE`.
  - `resp.State.FallbackUsed == false`.

### Degraded fallback modes

- `WINDOW_TREE` (ACC/MSAA) and `HWND_TREE` are compatibility/degraded views.
- `HWND_TREE` is explicitly **not equivalent** to UIA parity and must never be treated as a parity pass.
- ACC/MSAA parity is intentionally tracked separately from modern UIA parity; ACC output is expected to differ for UIA-rich applications.
- When fallback is active (`FallbackUsed == true`), expected operator-facing messaging is:
  - provider guidance text from inspect service: `UIA tree is unavailable. Switch to ACC/MSAA mode to continue inspecting this window.`
  - viewer status warning includes: `fallback mode active: degraded HWND/compatibility tree, selector parity may differ`.

### Failure interpretation

- **Parity failure, not product regression by itself:** UIA unavailable, but fallback tree loads.
- **Likely regression:** UIA mode requested and active mode remains `UIA_TREE`, but expected root/details/pattern ladders fail.
- **Hard failure:** UIA unavailable and fallback chain also fails (`Failed GetTreeRoot: <reason>` / `Failed to load window`).

## Deterministic validation ladder (Notepad flow)

Use a deterministic target window (Notepad) so teams can reproduce behavior quickly.

### Setup

1. Start `notepad.exe`.
2. Type sample text: `goahk parity check`.
3. Keep the Notepad window visible in foreground.

### Ladder checks

1) **Open + window enumerate**

- Click Refresh in viewer.
- Select Notepad row.
- **Pass:** status includes loaded counts; window row appears.
- **Fail:** `Failed to refresh windows`.

2) **Inspect root in UIA mode**

- Ensure mode is Auto/UIA + fallback (or explicitly UIA request path).
- Trigger inspect/select window.
- **Pass (parity-eligible):** active mode reports UIA and fallback=false.
- **Degraded pass (non-parity):** fallback message appears and active mode becomes `WINDOW_TREE` or `HWND_TREE`.
- **Fail:** `Failed GetTreeRoot: <reason>` or `Failed to load window`.

3) **Tree expansion + selection**

- Expand root and select first editable/content child.
- **Pass:** children load, selected node updates, highlight follows selection.
- **Fail:** `Failed to expand node` or `Failed to select node`.

4) **Property panel checks**

- Verify core property presence for selected element.
- Expected examples (UIA parity target):
  - `ControlType` resembles editable/text host for Notepad surface.
  - `Name` populated for top-level window/root.
  - `BoundingRectangle` present with non-empty geometry.
- **Pass:** expected fields resolve without global "unsupported" collapse.
- **Fail:** details load fails (`Failed GetNodeDetails: <reason>`) or property set is structurally empty for valid node.

5) **Pattern/actionability checks**

- Verify pattern actions pane for selected actionable node.
- **Pass:** supported actions are enabled; payload-required actions remain disabled until payload is entered.
- **Fail:** action invocation error such as `<ActionLabel> failed` for a known-supported action.

6) **Mode-status diagnosis check**

- Record mode summary in status text (requested/active/provider/fallback).
- **Pass:** status clearly distinguishes UIA parity from degraded fallback.
- **Fail:** UIA-unavailable scenarios surface without fallback labeling, causing ambiguous triage.

## 6-rung ladder (service/controller expectations)

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
  - Fallback cases are explicitly labeled as degraded and non-parity.
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

Coverage for this ladder is split across controller, service, and model tests in the viewer/inspect packages:

- `internal/inspect/validation_ladder_test.go`
- `internal/inspect/service_windows_test.go`
- `cmd/goahk-uia-viewer/controller_test.go`
- `cmd/goahk-uia-viewer/ui_events_test.go`
- `cmd/goahk-uia-viewer/model_window_table_windows_test.go`
- `cmd/goahk-uia-viewer/model_property_table_windows_test.go`
- `cmd/goahk-uia-viewer/model_pattern_tree_windows_test.go`
- `cmd/goahk-uia-viewer/model_uia_tree_windows_test.go`

## Gating rule

No viewer-polish task should be marked complete until:

- rungs **1–4** (window list, root resolution, root details, child expansion) pass in automated coverage, and
- fallback/degraded status labeling remains explicit so triage can separate UIA-unavailable environments from true regressions.


## UIA parity criteria

A run is parity-valid only when diagnostics/reporting indicates `Provider=uia`, `Mode=UIA_TREE`, and `Fallback=No`. Any `HWND_TREE` result is degraded-by-design and non-parity.


## Manual Windows desktop verification ladder (Notepad)

This ladder is **manual verification on an interactive Windows desktop**. It complements deterministic unit tests that use fake provider trees; it does not replace those tests.

### Preconditions

1. Use Windows desktop session (not headless CI).
2. Build and launch `goahk-uia-viewer`.
3. Launch `notepad.exe` and type `goahk parity check`.

### Step-by-step checks

1. Refresh windows and select Notepad.
2. Inspect in auto/UIA mode and record the status fields:
   - `Provider`
   - `Backend`
   - `Mode`
   - `Fallback`
3. Expand root, then expand the first `pane ""` node (empty-name pane must be retained).
4. Confirm expected descendants are visible, such as document/edit content and status elements.
5. Select an editable node and open details; verify UIA-backed properties populate.
6. Open pattern actions and verify common mappings (for example `Invoke()` and `SetValue()`) are shown when supported.

### Expected outcomes

- **Parity pass:** `Provider=uia`, `Mode=UIA_TREE`, `Fallback=No`.
- **UIA degraded fallback:** `Fallback=Yes`, with active mode moving to `WINDOW_TREE` or `HWND_TREE`; results are diagnostic, not parity-equivalent.
- **Failure:** root/details/actions fail without clear fallback messaging.

### Deterministic test complement

The fake-provider tests should remain authoritative for repeatable correctness checks (including empty-name panes, AHK-style labels, and pattern-action mappings). Manual ladder results are an additional confidence signal for real desktop behavior.
