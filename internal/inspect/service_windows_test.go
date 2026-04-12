//go:build windows
// +build windows

package inspect

import (
	"context"
	"errors"
	"testing"

	"goahk/internal/window"
)

type fakeWindowAdapter struct {
	windows      []window.Info
	enumerateErr error
	activateErr  error
	activated    []window.HWND
}

func (f *fakeWindowAdapter) EnumerateWindows(context.Context) ([]window.Info, error) {
	if f.enumerateErr != nil {
		return nil, f.enumerateErr
	}
	return append([]window.Info(nil), f.windows...), nil
}

func (f *fakeWindowAdapter) ActivateWindow(_ context.Context, hwnd window.HWND) error {
	if f.activateErr != nil {
		return f.activateErr
	}
	f.activated = append(f.activated, hwnd)
	return nil
}

func TestWindowsProvider_InspectAndTreeCacheBehavior(t *testing.T) {
	deps := &fakeAdapter{
		root: &uiaElement{Ref: "root", RuntimeID: "1", HWND: "0x1", Name: "Root"},
		kids: map[string][]*uiaElement{
			"root": {{Ref: "c1", RuntimeID: "2", ParentRef: "root", Name: "Child"}},
		},
	}
	provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)

	rootResp, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Refresh: false})
	if err != nil {
		t.Fatalf("GetTreeRoot setup failed: %v", err)
	}
	if _, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: rootResp.Root.NodeID}); err != nil {
		t.Fatalf("GetNodeChildren initial load failed: %v", err)
	}
	if _, err := provider.InspectWindow(context.Background(), InspectWindowRequest{HWND: "0x1"}); err != nil {
		t.Fatalf("InspectWindow failed: %v", err)
	}
	if _, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: rootResp.Root.NodeID}); err != nil {
		t.Fatalf("GetNodeChildren after InspectWindow failed: %v", err)
	}
	if got := deps.childrenCallCount["root"]; got != 1 {
		t.Fatalf("expected cached children after InspectWindow (refresh=false path), calls=%d", got)
	}

	if _, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Refresh: true}); err != nil {
		t.Fatalf("GetTreeRoot refresh failed: %v", err)
	}
	if _, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: rootResp.Root.NodeID}); err != nil {
		t.Fatalf("GetNodeChildren after refresh failed: %v", err)
	}
	if got := deps.childrenCallCount["root"]; got != 2 {
		t.Fatalf("expected children to reload after refresh invalidates cache, calls=%d", got)
	}
	if deps.resolveRootCalls < 2 {
		t.Fatalf("expected root resolution at tree roots only, calls=%d", deps.resolveRootCalls)
	}
}

func TestWindowsProvider_GetNodeChildrenPreservesProviderOrder(t *testing.T) {
	deps := &fakeAdapter{
		root: &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root"},
		kids: map[string][]*uiaElement{
			"root": {
				{Ref: "b", RuntimeID: "20", ParentRef: "root", Name: "Second"},
				{Ref: "a", RuntimeID: "10", ParentRef: "root", Name: "First"},
			},
		},
	}
	provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
	root, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
	if err != nil {
		t.Fatalf("GetTreeRoot failed: %v", err)
	}
	resp, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: root.Root.NodeID})
	if err != nil {
		t.Fatalf("GetNodeChildren failed: %v", err)
	}
	if len(resp.Children) != 2 || resp.Children[0].Name != "Second" || resp.Children[1].Name != "First" {
		t.Fatalf("expected UIA enumeration order to be preserved, got %+v", resp.Children)
	}
}

func TestWindowsProvider_UIAFailuresReturnStructuredErrors(t *testing.T) {
	deps := &fakeAdapter{
		root: &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root"},
		getChildrenErr: map[string]error{
			"root": errors.New("uia failure"),
		},
	}
	provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
	root, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
	if err != nil {
		t.Fatalf("GetTreeRoot failed: %v", err)
	}
	_, err = provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: root.Root.NodeID})
	var pErr *ProviderCallError
	if !errors.As(err, &pErr) {
		t.Fatalf("expected structured provider error, got %v", err)
	}
}

func TestWindowsProvider_GetTreeRoot_ModeRoutingAndFallbackState(t *testing.T) {
	windowTree := &fakeAdapter{
		root: &uiaElement{Ref: "root-window", RuntimeID: "101", HWND: "0x1", Name: "Window Root"},
	}
	provider := newWindowsProviderWithModeAdapters(newUIAAdapter(nil), newUIAAdapter(windowTree), &fakeWindowAdapter{}).(*windowsProvider)

	resp, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Mode: InspectModeUIATree})
	if err != nil {
		t.Fatalf("GetTreeRoot fallback failed: %v", err)
	}
	if resp.State.ActiveMode != InspectModeWindowTree || !resp.State.FallbackUsed {
		t.Fatalf("expected fallback state with WINDOW_TREE active mode, got %+v", resp.State)
	}
	if resp.State.FailureStage == "" || resp.State.GuidanceText == "" {
		t.Fatalf("expected explicit fallback metadata, got %+v", resp.State)
	}
}

func TestWindowsProvider_GetTreeRoot_ManualWindowTreeMode(t *testing.T) {
	uia := &fakeAdapter{root: &uiaElement{Ref: "root-uia", RuntimeID: "1", HWND: "0x1", Name: "UIA Root"}}
	windowTree := &fakeAdapter{root: &uiaElement{Ref: "root-window", RuntimeID: "2", HWND: "0x1", Name: "Window Root"}}
	provider := newWindowsProviderWithModeAdapters(newUIAAdapter(uia), newUIAAdapter(windowTree), &fakeWindowAdapter{}).(*windowsProvider)

	resp, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1", Mode: InspectModeWindowTree})
	if err != nil {
		t.Fatalf("GetTreeRoot manual mode failed: %v", err)
	}
	if resp.State.ActiveMode != InspectModeWindowTree || resp.State.FallbackUsed {
		t.Fatalf("expected WINDOW_TREE mode without fallback, got %+v", resp.State)
	}
	if windowTree.resolveRootCalls != 1 || uia.resolveRootCalls != 0 {
		t.Fatalf("expected only window tree adapter call, uia=%d window=%d", uia.resolveRootCalls, windowTree.resolveRootCalls)
	}
}

func TestWindowsProvider_WindowListingAndRefreshFilters(t *testing.T) {
	visible := true
	hidden := false
	windows := []window.Info{
		{HWND: 0x1, Title: "Notepad", Class: "Notepad", Exe: "notepad.exe", PID: 100, Visible: &visible},
		{HWND: 0x2, Title: "", Class: "Chrome_WidgetWin", Exe: "chrome.exe", PID: 200, Visible: &hidden},
		{HWND: 0x3, Title: "Terminal", Class: "CASCADIA_HOSTING_WINDOW_CLASS", Exe: "WindowsTerminal.exe", PID: 300, Visible: nil},
	}
	provider := newWindowsProviderWithDeps(&fakeAdapter{}, &fakeWindowAdapter{windows: windows}).(*windowsProvider)

	tests := []struct {
		name     string
		listReq  *ListWindowsRequest
		freshReq *RefreshWindowsRequest
		wantHWND []string
	}{
		{name: "list all", listReq: &ListWindowsRequest{}, wantHWND: []string{"0x2", "0x1", "0x3"}},
		{name: "list title contains", listReq: &ListWindowsRequest{TitleContains: "note"}, wantHWND: []string{"0x1"}},
		{name: "list class filter", listReq: &ListWindowsRequest{ClassName: "widget"}, wantHWND: []string{"0x2"}},
		{name: "refresh no filter", freshReq: &RefreshWindowsRequest{}, wantHWND: []string{"0x2", "0x1", "0x3"}},
		{name: "refresh visible only keeps nil visible", freshReq: &RefreshWindowsRequest{VisibleOnly: true}, wantHWND: []string{"0x1", "0x3"}},
		{name: "refresh title-only", freshReq: &RefreshWindowsRequest{Filter: "term", TitleOnly: true}, wantHWND: []string{"0x3"}},
		{name: "refresh broad filter checks class/exe/hwnd", freshReq: &RefreshWindowsRequest{Filter: "chrome"}, wantHWND: []string{"0x2"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got []WindowSummary
			var err error
			if tc.listReq != nil {
				var resp ListWindowsResponse
				resp, err = provider.ListWindows(context.Background(), *tc.listReq)
				got = resp.Windows
			} else {
				var resp RefreshWindowsResponse
				resp, err = provider.RefreshWindows(context.Background(), *tc.freshReq)
				got = resp.Windows
			}
			if err != nil {
				t.Fatalf("call failed: %v", err)
			}
			if len(got) != len(tc.wantHWND) {
				t.Fatalf("expected %d windows, got %d: %+v", len(tc.wantHWND), len(got), got)
			}
			for i := range tc.wantHWND {
				if got[i].HWND != tc.wantHWND[i] {
					t.Fatalf("index %d expected hwnd %s, got %s", i, tc.wantHWND[i], got[i].HWND)
				}
			}
		})
	}
}

func TestWindowsProvider_NodeAndPatternMethods(t *testing.T) {
	deps := &fakeAdapter{
		root: &uiaElement{Ref: "root", Name: "Root", SupportedPatterns: []string{"Invoke"}},
		byRef: map[string]*uiaElement{
			"root": {Ref: "root", Name: "Root", ControlType: "Button", ClassName: "Btn", AutomationID: "ok", SupportedPatterns: []string{"Invoke", "Value"}},
		},
	}
	provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{
		windows: []window.Info{{HWND: 0x1, Title: "Sample", Class: "SampleClass", Exe: "sample.exe", PID: 404}},
	}).(*windowsProvider)
	root, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
	if err != nil {
		t.Fatalf("GetTreeRoot: %v", err)
	}
	nodeID := root.Root.NodeID

	tests := []struct {
		name string
		fn   func(context.Context) error
	}{
		{name: "SelectNode", fn: func(ctx context.Context) error {
			resp, err := provider.SelectNode(ctx, SelectNodeRequest{NodeID: nodeID})
			if err == nil && resp.Selected.NodeID == "" {
				return errors.New("missing selected node")
			}
			return err
		}},
		{name: "GetNodeDetails", fn: func(ctx context.Context) error {
			resp, err := provider.GetNodeDetails(ctx, GetNodeDetailsRequest{NodeID: nodeID})
			if err == nil && (len(resp.Properties) == 0 || len(resp.Patterns) == 0) {
				return errors.New("missing node details")
			}
			if err == nil && (resp.WindowInfo.HWND == "" || resp.Element.ControlType == "" || len(resp.Path) == 0 || resp.SelectorPath.BestSelector == nil) {
				return errors.New("missing canonical detail fields")
			}
			if err == nil && resp.StatusText == "" {
				return errors.New("missing status text")
			}
			return err
		}},
		{name: "GetPatternActions", fn: func(ctx context.Context) error {
			resp, err := provider.GetPatternActions(ctx, GetPatternActionsRequest{NodeID: nodeID})
			if err == nil && len(resp.Actions) == 0 {
				return errors.New("missing actions")
			}
			return err
		}},
		{name: "InvokePattern", fn: func(ctx context.Context) error {
			_, err := provider.InvokePattern(ctx, InvokePatternRequest{NodeID: nodeID, Action: "invoke"})
			return err
		}},
		{name: "CopyBestSelector", fn: func(ctx context.Context) error {
			resp, err := provider.CopyBestSelector(ctx, CopyBestSelectorRequest{NodeID: nodeID})
			if err == nil && resp.Selector == "" {
				return errors.New("missing selector")
			}
			return err
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(context.Background()); err != nil {
				t.Fatalf("%s failed: %v", tc.name, err)
			}
		})
	}

	invalidNodeTests := []struct {
		name string
		fn   func(context.Context) error
	}{
		{name: "GetNodeChildren invalid", fn: func(ctx context.Context) error {
			_, err := provider.GetNodeChildren(ctx, GetNodeChildrenRequest{NodeID: "node:missing"})
			return err
		}},
		{name: "GetNodeDetails invalid", fn: func(ctx context.Context) error {
			_, err := provider.GetNodeDetails(ctx, GetNodeDetailsRequest{NodeID: "node:missing"})
			return err
		}},
		{name: "SelectNode invalid", fn: func(ctx context.Context) error {
			_, err := provider.SelectNode(ctx, SelectNodeRequest{NodeID: "node:missing"})
			return err
		}},
	}
	for _, tc := range invalidNodeTests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(context.Background()); !errors.Is(err, ErrStaleCache) {
				t.Fatalf("expected stale cache for invalid node id, got %v", err)
			}
		})
	}
}

func TestBuildPropertyList_IncludesDenseSetAndUnsupportedMarkers(t *testing.T) {
	selected := InspectElement{
		RuntimeID:            "1.2.3",
		LocalizedControlType: "button",
		ControlType:          "Button",
		Name:                 "Submit",
		AutomationID:         "submit-btn",
		BoundingRect:         &Rect{Left: 1, Top: 2, Width: 3, Height: 4},
		ClassName:            "Button",
		HelpText:             strPtr("help"),
		AccessKey:            strPtr("alt+s"),
		AcceleratorKey:       strPtr("ctrl+s"),
		HasKeyboardFocus:     true,
		IsKeyboardFocusable:  true,
		ItemType:             strPtr("action"),
		ItemStatus:           strPtr("ready"),
		ProcessID:            88,
		IsEnabled:            true,
		IsPassword:           false,
		IsOffscreen:          false,
		FrameworkID:          "WPF",
		IsRequiredForForm:    true,
		LabeledBy:            strPtr("Submit label"),
		UnsupportedProps: map[string]bool{
			"HelpText":          true,
			"BoundingRectangle": true,
			"ProcessId":         true,
		},
	}

	properties := buildPropertyList(selected)
	if len(properties) != 22 {
		t.Fatalf("expected dense property list of 22, got %d", len(properties))
	}
	assertProperty := func(name string, expectedGroup string, expectedStatus string, expectValue string) {
		t.Helper()
		for _, property := range properties {
			if property.Name != name {
				continue
			}
			if property.Group != expectedGroup || property.Status != expectedStatus {
				t.Fatalf("unexpected metadata for %s: %+v", name, property)
			}
			if expectValue == "<nil>" {
				if property.Value != nil {
					t.Fatalf("expected nil value for %s, got %q", name, *property.Value)
				}
				return
			}
			if property.Value == nil || *property.Value != expectValue {
				t.Fatalf("unexpected value for %s: %+v", name, property.Value)
			}
			return
		}
		t.Fatalf("missing property %s", name)
	}

	assertProperty("RuntimeID", "identity", "ok", "1.2.3")
	assertProperty("HelpText", "semantics", "unsupported", "<nil>")
	assertProperty("ProcessId", "identity", "unsupported", "<nil>")
	assertProperty("BoundingRectangle", "geometry", "unsupported", "<nil>")
	assertProperty("IsEnabled", "state", "ok", "true")
	assertProperty("LabeledBy", "relation", "ok", "Submit label")
}

func strPtr(v string) *string { return &v }

func TestWindowsProvider_GetNodeDetailsStatusAndPathFallbacks(t *testing.T) {
	deps := &fakeAdapter{
		root: &uiaElement{Ref: "root", Name: "Root"},
		byRef: map[string]*uiaElement{
			"root": {Ref: "root", Name: "Root", ControlType: "Window"},
		},
	}
	provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)

	root, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
	if err != nil {
		t.Fatalf("GetTreeRoot failed: %v", err)
	}
	resp, err := provider.GetNodeDetails(context.Background(), GetNodeDetailsRequest{NodeID: root.Root.NodeID})
	if err != nil {
		t.Fatalf("GetNodeDetails failed: %v", err)
	}
	if resp.StatusText != "Loaded node details: Root" {
		t.Fatalf("expected canonical status text, got %q", resp.StatusText)
	}
	if len(resp.Path) == 0 || resp.Path[len(resp.Path)-1].NodeID != root.Root.NodeID {
		t.Fatalf("expected populated path ending in selected node, got %+v", resp.Path)
	}
	if _, err := provider.GetNodeDetails(context.Background(), GetNodeDetailsRequest{NodeID: "node:missing"}); !errors.Is(err, ErrStaleCache) {
		t.Fatalf("expected stale cache on missing node details, got %v", err)
	}
}

func TestWindowsProvider_ActivateAndFollowCursor(t *testing.T) {
	deps := &fakeAdapter{
		root:    &uiaElement{Ref: "root", Name: "Root"},
		focused: &uiaElement{Ref: "focused", Name: "Focused"},
		under:   &uiaElement{Ref: "under", Name: "Under Cursor"},
		cursorX: 10,
		cursorY: 10,
	}
	windows := &fakeWindowAdapter{}
	provider := newWindowsProviderWithDeps(newUIAAdapter(deps), windows).(*windowsProvider)

	focused, err := provider.GetFocusedElement(context.Background(), GetFocusedElementRequest{})
	if err != nil {
		t.Fatalf("GetFocusedElement: %v", err)
	}
	if focused.Element.NodeID == "" {
		t.Fatalf("expected focused element node id")
	}

	if _, err := provider.ToggleFollowCursor(context.Background(), ToggleFollowCursorRequest{Enabled: true}); err != nil {
		t.Fatalf("enable follow cursor: %v", err)
	}
	resp, err := provider.GetElementUnderCursor(context.Background(), GetElementUnderCursorRequest{})
	if err != nil {
		t.Fatalf("GetElementUnderCursor: %v", err)
	}
	if resp.Element.NodeID == "" {
		t.Fatalf("expected cursor element node id")
	}
	provider.followMu.RLock()
	if provider.focusedUnderCursor.NodeID == "" {
		t.Fatalf("expected focused-under-cursor state to be updated")
	}
	provider.followMu.RUnlock()

	disableResp, err := provider.ToggleFollowCursor(context.Background(), ToggleFollowCursorRequest{Enabled: false})
	if err != nil {
		t.Fatalf("disable follow cursor: %v", err)
	}
	if disableResp.Enabled {
		t.Fatalf("expected disabled response")
	}

	if _, err := provider.ActivateWindow(context.Background(), ActivateWindowRequest{HWND: "0x2A"}); err != nil {
		t.Fatalf("ActivateWindow valid: %v", err)
	}
	if len(windows.activated) != 1 || windows.activated[0] != window.HWND(0x2A) {
		t.Fatalf("expected activated hwnd 0x2A, got %+v", windows.activated)
	}
	if _, err := provider.ActivateWindow(context.Background(), ActivateWindowRequest{HWND: "nope"}); !errors.Is(err, ErrInvalidNodeID) {
		t.Fatalf("expected invalid node id for invalid hwnd, got %v", err)
	}
}

func TestWindowsProvider_UIAErrorMappingAndNilFallbacks(t *testing.T) {
	provider := newWindowsProviderWithDeps(nil, &fakeWindowAdapter{}).(*windowsProvider)
	if _, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"}); !errors.Is(err, ErrProviderActionUnsupported) {
		t.Fatalf("expected unsupported adapter fallback when nil deps provided, got %v", err)
	}

	staleDeps := &fakeAdapter{focusErr: errUIAElementNotAvailable}
	provider = newWindowsProviderWithDeps(newUIAAdapter(staleDeps), &fakeWindowAdapter{}).(*windowsProvider)
	if _, err := provider.GetFocusedElement(context.Background(), GetFocusedElementRequest{}); !errors.Is(err, ErrStaleCache) {
		t.Fatalf("expected stale cache mapping for element-not-available, got %v", err)
	}

	nilDeps := &fakeAdapter{pointErr: errUIANilElement}
	provider = newWindowsProviderWithDeps(newUIAAdapter(nilDeps), &fakeWindowAdapter{}).(*windowsProvider)
	if _, err := provider.GetElementUnderCursor(context.Background(), GetElementUnderCursorRequest{}); !errors.Is(err, ErrStaleCache) {
		t.Fatalf("expected stale cache mapping for nil element error, got %v", err)
	}
}

func TestWindowsProvider_MethodErrorsPropagate(t *testing.T) {
	provider := newWindowsProviderWithDeps(&fakeAdapter{}, &fakeWindowAdapter{enumerateErr: errors.New("enumerate failed")}).(*windowsProvider)
	if _, err := provider.ListWindows(context.Background(), ListWindowsRequest{}); err == nil {
		t.Fatalf("expected list windows error")
	}
	if _, err := provider.RefreshWindows(context.Background(), RefreshWindowsRequest{}); err == nil {
		t.Fatalf("expected refresh windows error")
	}
}

func TestWindowsProvider_InvokePatternDispatchesToAdapterActions(t *testing.T) {
	deps := &fakeAdapter{
		root: &uiaElement{
			Ref:               "root",
			Name:              "Root",
			SupportedPatterns: []string{"Invoke", "SelectionItem", "Value", "Toggle", "ExpandCollapse"},
		},
		byRef: map[string]*uiaElement{
			"root": {
				Ref:               "root",
				Name:              "Root",
				SupportedPatterns: []string{"Invoke", "SelectionItem", "Value", "Toggle", "ExpandCollapse"},
			},
		},
	}
	provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
	root, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
	if err != nil {
		t.Fatalf("GetTreeRoot: %v", err)
	}
	nodeID := root.Root.NodeID

	actions := []InvokePatternRequest{
		{NodeID: nodeID, Action: "invoke"},
		{NodeID: nodeID, Action: "select"},
		{NodeID: nodeID, Action: "setValue", Payload: map[string]any{"value": "abc"}},
		{NodeID: nodeID, Action: "toggle"},
		{NodeID: nodeID, Action: "expand"},
		{NodeID: nodeID, Action: "collapse"},
	}
	for _, req := range actions {
		if _, err := provider.InvokePattern(context.Background(), req); err != nil {
			t.Fatalf("InvokePattern %s failed: %v", req.Action, err)
		}
	}

	if deps.invokeCount != 1 || deps.selectCount != 1 || deps.setValueCount != 1 {
		t.Fatalf("expected invoke/select/setValue dispatch count 1 each, got invoke=%d select=%d setValue=%d", deps.invokeCount, deps.selectCount, deps.setValueCount)
	}
	if deps.toggleCount != 1 || deps.expandCount != 1 || deps.collapseCount != 1 {
		t.Fatalf("expected toggle/expand/collapse dispatch count 1 each, got toggle=%d expand=%d collapse=%d", deps.toggleCount, deps.expandCount, deps.collapseCount)
	}
	if deps.lastSetValue != "abc" {
		t.Fatalf("expected setValue payload forwarded, got %q", deps.lastSetValue)
	}
}

func TestWindowsProvider_FollowCursorPauseAndLock(t *testing.T) {
	deps := &fakeAdapter{
		root:        &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root"},
		underCursor: &uiaElement{Ref: "cursor", RuntimeID: "2", Name: "Cursor"},
	}
	provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
	if _, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"}); err != nil {
		t.Fatalf("setup root: %v", err)
	}
	if _, err := provider.ToggleFollowCursor(context.Background(), ToggleFollowCursorRequest{Enabled: true}); err != nil {
		t.Fatalf("enable follow: %v", err)
	}
	if _, err := provider.PauseFollowCursor(context.Background(), PauseFollowCursorRequest{}); err != nil {
		t.Fatalf("pause follow: %v", err)
	}
	pausedResp, err := provider.GetElementUnderCursor(context.Background(), GetElementUnderCursorRequest{})
	if err != nil {
		t.Fatalf("under cursor when paused: %v", err)
	}
	if pausedResp.Element.NodeID != "" {
		t.Fatalf("expected empty element while paused, got %+v", pausedResp.Element)
	}
	if _, err := provider.ResumeFollowCursor(context.Background(), ResumeFollowCursorRequest{}); err != nil {
		t.Fatalf("resume follow: %v", err)
	}
	locked, err := provider.LockFollowCursor(context.Background(), LockFollowCursorRequest{})
	if err != nil {
		t.Fatalf("lock follow: %v", err)
	}
	if !locked.Locked || locked.NodeID == "" {
		t.Fatalf("expected lock response with node id, got %+v", locked)
	}
	first, _ := provider.GetElementUnderCursor(context.Background(), GetElementUnderCursorRequest{})
	second, _ := provider.GetElementUnderCursor(context.Background(), GetElementUnderCursorRequest{})
	if first.Element.NodeID == "" || first.Element.NodeID != second.Element.NodeID {
		t.Fatalf("expected locked node to stay stable: first=%+v second=%+v", first.Element, second.Element)
	}
	if _, err := provider.UnlockFollowCursor(context.Background(), UnlockFollowCursorRequest{}); err != nil {
		t.Fatalf("unlock follow: %v", err)
	}
}

func TestWindowsProvider_RefreshGranularityTargetsOnlyRequestedScope(t *testing.T) {
	deps := &fakeAdapter{
		root: &uiaElement{Ref: "root", RuntimeID: "1", Name: "Root"},
		kids: map[string][]*uiaElement{
			"root": {{Ref: "child", RuntimeID: "2", ParentRef: "root", Name: "Child"}},
		},
	}
	provider := newWindowsProviderWithDeps(newUIAAdapter(deps), &fakeWindowAdapter{}).(*windowsProvider)
	root, err := provider.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	if _, err := provider.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: root.Root.NodeID}); err != nil {
		t.Fatalf("children: %v", err)
	}
	childrenCallsBefore := deps.childrenCallCount["root"]
	rootCallsBefore := deps.resolveRootCalls

	if _, err := provider.RefreshNodeChildren(context.Background(), RefreshNodeChildrenRequest{NodeID: root.Root.NodeID}); err != nil {
		t.Fatalf("refresh children: %v", err)
	}
	if deps.childrenCallCount["root"] != childrenCallsBefore+1 {
		t.Fatalf("expected only child scope reloaded, calls=%d->%d", childrenCallsBefore, deps.childrenCallCount["root"])
	}
	if deps.resolveRootCalls != rootCallsBefore {
		t.Fatalf("did not expect root reload for child refresh, got %d->%d", rootCallsBefore, deps.resolveRootCalls)
	}
	if _, err := provider.RefreshNodeDetails(context.Background(), RefreshNodeDetailsRequest{NodeID: root.Root.NodeID}); err != nil {
		t.Fatalf("refresh details: %v", err)
	}
	if deps.resolveRootCalls != rootCallsBefore {
		t.Fatalf("did not expect root reload for detail refresh, got %d->%d", rootCallsBefore, deps.resolveRootCalls)
	}
}

func TestDiagnosticsFromError_MapsAccessDenied(t *testing.T) {
	diag := diagnosticsFromError("ResolveWindowRoot", errors.New("E_ACCESSDENIED"), "")
	if diag == nil {
		t.Fatalf("expected diagnostics payload")
	}
	if diag.ErrorCode != "access_denied" || diag.HResult != "0x80070005" {
		t.Fatalf("unexpected diagnostics code mapping: %+v", diag)
	}
	if diag.PrivilegeHint == "" {
		t.Fatalf("expected privilege hint")
	}
}
