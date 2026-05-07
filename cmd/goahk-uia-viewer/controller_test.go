package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"goahk/internal/inspect"
)

type fakeInspectService struct {
	mu                sync.Mutex
	underCursorCalls  int
	underCursorValues []inspect.TreeNodeDTO
	inspectWindowReqs []inspect.InspectWindowRequest
	inspectWindowResp inspect.InspectWindowResponse
	refreshReqs       []inspect.RefreshWindowsRequest
	nodeChildrenReqs  []inspect.GetNodeChildrenRequest
	childrenByNode    map[string][]inspect.TreeNodeDTO
	clearCalls        int
	callOrder         []string
	nodeDetailsResp   inspect.GetNodeDetailsResponse
	invokeReqs        []inspect.InvokePatternRequest
}

func (f *fakeInspectService) ListWindows(context.Context, inspect.ListWindowsRequest) (inspect.ListWindowsResponse, error) {
	return inspect.ListWindowsResponse{}, nil
}
func (f *fakeInspectService) InspectWindow(_ context.Context, req inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.inspectWindowReqs = append(f.inspectWindowReqs, req)
	return f.inspectWindowResp, nil
}
func (f *fakeInspectService) GetTreeRoot(context.Context, inspect.GetTreeRootRequest) (inspect.GetTreeRootResponse, error) {
	return inspect.GetTreeRootResponse{}, nil
}
func (f *fakeInspectService) GetNodeChildren(_ context.Context, req inspect.GetNodeChildrenRequest) (inspect.GetNodeChildrenResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nodeChildrenReqs = append(f.nodeChildrenReqs, req)
	children := f.childrenByNode[req.NodeID]
	return inspect.GetNodeChildrenResponse{Children: children}, nil
}
func (f *fakeInspectService) SelectNode(context.Context, inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	return inspect.SelectNodeResponse{}, nil
}
func (f *fakeInspectService) GetNodeDetails(context.Context, inspect.GetNodeDetailsRequest) (inspect.GetNodeDetailsResponse, error) {
	return f.nodeDetailsResp, nil
}
func (f *fakeInspectService) GetFocusedElement(context.Context, inspect.GetFocusedElementRequest) (inspect.GetFocusedElementResponse, error) {
	return inspect.GetFocusedElementResponse{}, nil
}
func (f *fakeInspectService) GetElementUnderCursor(context.Context, inspect.GetElementUnderCursorRequest) (inspect.GetElementUnderCursorResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.underCursorValues) == 0 {
		return inspect.GetElementUnderCursorResponse{}, errors.New("no data")
	}
	idx := f.underCursorCalls
	if idx >= len(f.underCursorValues) {
		idx = len(f.underCursorValues) - 1
	}
	f.underCursorCalls++
	return inspect.GetElementUnderCursorResponse{Element: f.underCursorValues[idx]}, nil
}
func (f *fakeInspectService) HighlightNode(context.Context, inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	f.mu.Lock()
	f.callOrder = append(f.callOrder, "highlight")
	f.mu.Unlock()
	return inspect.HighlightNodeResponse{}, nil
}
func (f *fakeInspectService) ClearHighlight(context.Context, inspect.ClearHighlightRequest) (inspect.ClearHighlightResponse, error) {
	f.mu.Lock()
	f.clearCalls++
	f.callOrder = append(f.callOrder, "clear")
	f.mu.Unlock()
	return inspect.ClearHighlightResponse{}, nil
}
func (f *fakeInspectService) CopyBestSelector(context.Context, inspect.CopyBestSelectorRequest) (inspect.CopyBestSelectorResponse, error) {
	return inspect.CopyBestSelectorResponse{}, nil
}
func (f *fakeInspectService) GetPatternActions(context.Context, inspect.GetPatternActionsRequest) (inspect.GetPatternActionsResponse, error) {
	return inspect.GetPatternActionsResponse{}, nil
}
func (f *fakeInspectService) InvokePattern(_ context.Context, req inspect.InvokePatternRequest) (inspect.InvokePatternResponse, error) {
	f.invokeReqs = append(f.invokeReqs, req)
	return inspect.InvokePatternResponse{}, nil
}
func (f *fakeInspectService) ActivateWindow(context.Context, inspect.ActivateWindowRequest) (inspect.ActivateWindowResponse, error) {
	return inspect.ActivateWindowResponse{}, nil
}
func (f *fakeInspectService) ToggleFollowCursor(context.Context, inspect.ToggleFollowCursorRequest) (inspect.ToggleFollowCursorResponse, error) {
	return inspect.ToggleFollowCursorResponse{}, nil
}
func (f *fakeInspectService) PauseFollowCursor(context.Context, inspect.PauseFollowCursorRequest) (inspect.PauseFollowCursorResponse, error) {
	return inspect.PauseFollowCursorResponse{Paused: true}, nil
}
func (f *fakeInspectService) ResumeFollowCursor(context.Context, inspect.ResumeFollowCursorRequest) (inspect.ResumeFollowCursorResponse, error) {
	return inspect.ResumeFollowCursorResponse{Paused: false}, nil
}
func (f *fakeInspectService) LockFollowCursor(_ context.Context, req inspect.LockFollowCursorRequest) (inspect.LockFollowCursorResponse, error) {
	return inspect.LockFollowCursorResponse{Locked: true, NodeID: req.NodeID}, nil
}
func (f *fakeInspectService) UnlockFollowCursor(context.Context, inspect.UnlockFollowCursorRequest) (inspect.UnlockFollowCursorResponse, error) {
	return inspect.UnlockFollowCursorResponse{Locked: false}, nil
}
func (f *fakeInspectService) RefreshWindows(_ context.Context, req inspect.RefreshWindowsRequest) (inspect.RefreshWindowsResponse, error) {
	f.mu.Lock()
	f.refreshReqs = append(f.refreshReqs, req)
	f.mu.Unlock()
	return inspect.RefreshWindowsResponse{}, nil
}
func (f *fakeInspectService) RefreshTreeRoot(context.Context, inspect.RefreshTreeRootRequest) (inspect.RefreshTreeRootResponse, error) {
	return inspect.RefreshTreeRootResponse{}, nil
}
func (f *fakeInspectService) RefreshNodeChildren(context.Context, inspect.RefreshNodeChildrenRequest) (inspect.RefreshNodeChildrenResponse, error) {
	return inspect.RefreshNodeChildrenResponse{}, nil
}
func (f *fakeInspectService) RefreshNodeDetails(context.Context, inspect.RefreshNodeDetailsRequest) (inspect.RefreshNodeDetailsResponse, error) {
	return inspect.RefreshNodeDetailsResponse{}, nil
}
func (f *fakeInspectService) GetDiagnostics(context.Context, inspect.GetDiagnosticsRequest) (inspect.GetDiagnosticsResponse, error) {
	return inspect.GetDiagnosticsResponse{}, nil
}

type fakeClipboard struct{ copied []string }

func (f *fakeClipboard) CopyText(v string) error { f.copied = append(f.copied, v); return nil }

type fakeDialogs struct {
	value string
	ok    bool
	err   error
}

func (f *fakeDialogs) PromptSetValue(string) (string, bool, error) { return f.value, f.ok, f.err }

func TestController_Refresh_ForwardsFilters(t *testing.T) {
	svc := &fakeInspectService{}
	c := NewController(context.Background(), svc)
	_, err := c.RefreshWindows("abc", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.refreshReqs) != 1 {
		t.Fatalf("expected 1 call")
	}
}

func TestController_LoadTreeRoot_UsesSelectedWindow(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	if _, err := c.SelectWindow("0x1", false); err != nil {
		t.Fatal(err)
	}
	resp, err := c.LoadTreeRoot()
	if err != nil {
		t.Fatal(err)
	}
	if resp.Root.NodeID != "root" || svc.treeRootCalls < 2 {
		t.Fatalf("unexpected root response: %+v calls=%d", resp.Root, svc.treeRootCalls)
	}
}

func TestController_ExpandNode_CachesChildren(t *testing.T) {
	svc := &fakeInspectService{childrenByNode: map[string][]inspect.TreeNodeDTO{"n1": {{NodeID: "c1"}, {NodeID: "c2"}}}}
	c := NewController(context.Background(), svc)
	resp, err := c.ExpandNode("n1")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Children) != 2 {
		t.Fatalf("children=%d", len(resp.Children))
	}
}

func TestController_ExpandTreeDepth_ExpandsBreadthFirstToDepth(t *testing.T) {
	svc := &fakeInspectService{childrenByNode: map[string][]inspect.TreeNodeDTO{
		"root": {{NodeID: "a"}, {NodeID: "b"}},
		"a":    {{NodeID: "a1"}},
		"b":    {{NodeID: "b1"}},
	}}
	c := NewController(context.Background(), svc)
	results := c.ExpandTreeDepth("root", 2)
	if len(results) == 0 {
		t.Fatal("expected expansion results")
	}
	if len(svc.nodeChildrenReqs) < 3 {
		t.Fatalf("expected depth expansion calls, got %d", len(svc.nodeChildrenReqs))
	}
}

func TestController_InvokePattern_ForSelection(t *testing.T) {
	svc := &fakeInspectService{}
	c := NewController(context.Background(), svc)
	c.selectedNodeID = "node-1"
	_, err := c.InvokePatternForSelection("invoke")
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.invokeReqs) != 1 || svc.invokeReqs[0].NodeID != "node-1" || svc.invokeReqs[0].Action != "invoke" {
		t.Fatalf("unexpected invoke request: %+v", svc.invokeReqs)
	}
}

func TestController_SelectWindow_Pipeline(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	if _, err := c.SelectWindow("0x1", true); err != nil {
		t.Fatal(err)
	}
	if svc.clearCalls == 0 {
		t.Fatal("expected highlight clear on switch")
	}
	if svc.activateCalls != 1 || svc.treeRootCalls != 1 || svc.nodeDetailsCalls != 1 {
		t.Fatalf("unexpected pipeline calls: %+v", svc)
	}
}

func TestSelectWindowReturnsRootAndDetails(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root"}}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	result, err := c.SelectWindow("0x1", false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Root.Root.NodeID != "root" || result.Details.ACCPath != "root" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSelectWindowDoesNotRequireInspectWindow(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root"}},
		root:               inspect.TreeNodeDTO{NodeID: "root"},
		inspectErr:         errors.New("inspect failed"),
	}
	c := NewController(context.Background(), svc)
	result, err := c.SelectWindow("0x1", false)
	if err != nil {
		t.Fatalf("expected select to succeed without InspectWindow, got: %v", err)
	}
	if result.Root.Root.NodeID != "root" || result.Details.ACCPath != "root" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
func TestSelectWindowSetsSelectedNodeIDToRoot(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root-id"}}
	c := NewController(context.Background(), svc)
	if _, err := c.SelectWindow("0x1", false); err != nil {
		t.Fatal(err)
	}
	if c.selectedNodeID != "root-id" {
		t.Fatalf("selectedNodeID=%q", c.selectedNodeID)
	}
}
func TestSelectWindowStoresRootNodeInCache(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root-id"}}
	c := NewController(context.Background(), svc)
	if _, err := c.SelectWindow("0x1", false); err != nil {
		t.Fatal(err)
	}
	if _, ok := c.nodesByID["root-id"]; !ok {
		t.Fatal("expected root in cache")
	}
}
func TestSelectWindowHighlightsRootNode(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root-id"}}
	c := NewController(context.Background(), svc)
	if _, err := c.SelectWindow("0x1", false); err != nil {
		t.Fatal(err)
	}
	if svc.highlightCalls != 1 || svc.selectCalls != 1 {
		t.Fatalf("expected root select/highlight once: %+v", svc)
	}
}
func TestSelectWindowActivateTrueCallsActivateWindow(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root-id"}}
	c := NewController(context.Background(), svc)
	if _, err := c.SelectWindow("0x1", true); err != nil {
		t.Fatal(err)
	}
	if svc.activateCalls != 1 {
		t.Fatalf("activate calls=%d", svc.activateCalls)
	}
}
func TestSelectWindowActivateFalseDoesNotCallActivateWindow(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root-id"}}
	c := NewController(context.Background(), svc)
	if _, err := c.SelectWindow("0x1", false); err != nil {
		t.Fatal(err)
	}
	if svc.activateCalls != 0 {
		t.Fatalf("activate calls=%d", svc.activateCalls)
	}
}
func TestRefreshSelectedNodeDetailsAfterSelectWindowUsesRoot(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root-id"}}, root: inspect.TreeNodeDTO{NodeID: "root-id"}}
	c := NewController(context.Background(), svc)
	if _, err := c.SelectWindow("0x1", false); err != nil {
		t.Fatal(err)
	}
	details, err := c.RefreshSelectedNodeDetails()
	if err != nil {
		t.Fatal(err)
	}
	if details.ACCPath != "root-id" {
		t.Fatalf("node=%q", details.ACCPath)
	}
}
func TestController_SelectNode_Pipeline(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root/child"}}}
	c := NewController(context.Background(), svc)
	c.ToggleAccPathCapture()
	if err := c.SelectNode("n1"); err != nil {
		t.Fatal(err)
	}
	if svc.highlightCalls != 1 {
		t.Fatal("expected highlight")
	}
	if c.statusText != "Path: root/child" {
		t.Fatalf("status=%q", c.statusText)
	}
}
func TestController_Shutdown_CleansUp(t *testing.T) {
	svc := &fakeInspectService{underCursorValues: []inspect.TreeNodeDTO{{NodeID: "a"}}}
	c := NewController(context.Background(), svc)
	tick := make(chan time.Time, 1)
	c.followTicker = func() <-chan time.Time { return tick }
	_ = c.ToggleFollowCursor(true)
	c.Shutdown()
	if svc.clearCalls == 0 {
		t.Fatal("expected clear")
	}
	if c.followEnabled || c.followPaused || c.followLocked || c.followCancel != nil {
		t.Fatal("expected safe terminal follow state")
	}
	if c.selectedNodeID != "" || c.lastFollowNode != "" {
		t.Fatal("expected selection and follow cache reset on shutdown")
	}
}
func TestController_SetValuePromptAndInvoke(t *testing.T) {
	svc := &fakeInspectService{}
	c := NewController(context.Background(), svc).WithDialogs(&fakeDialogs{value: "abc", ok: true})
	c.selectedNodeID = "node-1"
	_, accepted, err := c.InvokeSetValue()
	if err != nil {
		t.Fatal(err)
	}
	if !accepted {
		t.Fatal("expected accepted set value")
	}
	if len(svc.invokeReqs) != 1 || svc.invokeReqs[0].Action != "setValue" {
		t.Fatal("missing invoke")
	}
}
func TestController_StatusInteraction_CopyPath(t *testing.T) {
	svc := &fakeInspectService{}
	cb := &fakeClipboard{}
	c := NewController(context.Background(), svc).WithClipboard(cb)
	c.ToggleAccPathCapture()
	c.lastACCPath = "a/b"
	got := c.OnStatusInteraction()
	if got != "Path: a/b" || len(cb.copied) != 1 {
		t.Fatalf("got=%q copied=%v", got, cb.copied)
	}
}
func TestController_StatusInteraction_TogglesCaptureWithoutPath(t *testing.T) {
	c := NewController(context.Background(), &fakeInspectService{})
	upd := c.OnStatusInteractionUpdate()
	if !upd.CaptureEnabled || upd.Text != "Click on path to copy to Clipboard" {
		t.Fatalf("unexpected update: %+v", upd)
	}
}

func TestController_StatusInteraction_DoesNotCopyWhenCaptureDisabled(t *testing.T) {
	cb := &fakeClipboard{}
	c := NewController(context.Background(), &fakeInspectService{}).WithClipboard(cb)
	c.lastACCPath = "a/b"
	upd := c.OnStatusInteractionUpdate()
	if len(cb.copied) != 0 || !upd.CaptureEnabled {
		t.Fatalf("expected toggle, no copy: copied=%v update=%+v", cb.copied, upd)
	}
}

func TestController_Refresh_ClearsHighlightFirst(t *testing.T) {
	svc := &fakeInspectService{}
	c := NewController(context.Background(), svc)
	_, _ = c.RefreshWindows("", true, true)
	if len(svc.callOrder) == 0 || svc.callOrder[0] != "clear" {
		t.Fatalf("expected first call clear, got %v", svc.callOrder)
	}
}
func TestController_FollowCursor_EmitsOnNodeChange_RespectsPauseLock(t *testing.T) {
	svc := &fakeInspectService{underCursorValues: []inspect.TreeNodeDTO{{NodeID: "a"}, {NodeID: "a"}, {NodeID: "b"}, {NodeID: "c"}}}
	c := NewController(context.Background(), svc)
	tick := make(chan time.Time, 8)
	c.followTicker = func() <-chan time.Time { return tick }
	var got []string
	var mu sync.Mutex
	c.OnFollowCursorElement(func(n inspect.TreeNodeDTO) { mu.Lock(); got = append(got, n.NodeID); mu.Unlock() })
	_ = c.ToggleFollowCursor(true)
	tick <- time.Now()
	tick <- time.Now()
	tick <- time.Now()
	c.PauseFollowCursor()
	tick <- time.Now()
	c.ResumeFollowCursor()
	c.LockFollowCursor()
	tick <- time.Now()
	c.UnlockFollowCursor()
	tick <- time.Now()
	time.Sleep(20 * time.Millisecond)
	_ = c.ToggleFollowCursor(false)
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
}
func TestController_FollowCursor_DisableClearsPauseAndLock(t *testing.T) {
	svc := &fakeInspectService{}
	c := NewController(context.Background(), svc)
	tick := make(chan time.Time, 1)
	c.followTicker = func() <-chan time.Time { return tick }
	_ = c.ToggleFollowCursor(true)
	c.PauseFollowCursor()
	c.LockFollowCursor()
	if err := c.ToggleFollowCursor(false); err != nil {
		t.Fatal(err)
	}
	if c.followEnabled || c.followPaused || c.followLocked {
		t.Fatal("expected follow flags reset when disabled")
	}
}
func TestController_OnStatusInteractionUpdate_CopiesPathWhenEnabled(t *testing.T) {
	cb := &fakeClipboard{}
	c := NewController(context.Background(), &fakeInspectService{}).WithClipboard(cb)
	c.ToggleAccPathCapture()
	c.lastACCPath = "desktop/window/button"
	upd := c.OnStatusInteractionUpdate()
	if !upd.LastACCPathCopied || upd.Text != "Path: desktop/window/button" || len(cb.copied) != 1 {
		t.Fatalf("unexpected update: %+v copied=%v", upd, cb.copied)
	}
}
func TestController_InvokeSetValue_RequiresDialog(t *testing.T) {
	c := NewController(context.Background(), &fakeInspectService{})
	if _, _, err := c.InvokeSetValue(); err == nil {
		t.Fatal("expected dialog unavailable error")
	}
}
func TestNormalizeInspectError_StableMapping(t *testing.T) {
	if normalizeInspectError(inspect.ErrInvalidNodeID) != "ErrInvalidNodeID" {
		t.Fatal("mapping changed")
	}
	if normalizeInspectError(errors.New("x")) == "none" {
		t.Fatal("unexpected")
	}
}

type fakeControllerService struct {
	fakeInspectService
	root                                                                                      inspect.TreeNodeDTO
	activateCalls, inspectCalls, treeRootCalls, nodeDetailsCalls, selectCalls, highlightCalls int
	inspectErr                                                                                error
	treeRootErr                                                                               error
	treeRootErrs                                                                              []error
	nodeDetailsErr                                                                            error
	nodeChildrenErr                                                                           error
	selectErr                                                                                 error
	highlightErr                                                                              error
	rootState                                                                                 inspect.InspectModeState
}

func (f *fakeControllerService) ActivateWindow(context.Context, inspect.ActivateWindowRequest) (inspect.ActivateWindowResponse, error) {
	f.activateCalls++
	return inspect.ActivateWindowResponse{}, nil
}
func (f *fakeControllerService) InspectWindow(context.Context, inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	f.inspectCalls++
	if f.inspectErr != nil {
		return inspect.InspectWindowResponse{}, f.inspectErr
	}
	return inspect.InspectWindowResponse{}, nil
}
func (f *fakeControllerService) GetTreeRoot(context.Context, inspect.GetTreeRootRequest) (inspect.GetTreeRootResponse, error) {
	f.treeRootCalls++
	if len(f.treeRootErrs) > 0 {
		err := f.treeRootErrs[0]
		f.treeRootErrs = f.treeRootErrs[1:]
		if err != nil {
			return inspect.GetTreeRootResponse{}, err
		}
	}
	if f.treeRootErr != nil {
		return inspect.GetTreeRootResponse{}, f.treeRootErr
	}
	return inspect.GetTreeRootResponse{Root: f.root, State: f.rootState}, nil
}

func TestGetTreeRootWithRetry_RetriesTransient(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{},
		root:               inspect.TreeNodeDTO{NodeID: "root-id"},
		treeRootErrs:       []error{inspect.ErrTransientFailure, inspect.ErrStaleCache, nil},
	}
	c := NewController(context.Background(), svc)
	root, warnings, err := c.getTreeRootWithRetry("0x1", inspect.InspectModeUIATree)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if root.Root.NodeID != "root-id" || svc.treeRootCalls != 3 {
		t.Fatalf("unexpected root/calls: root=%+v calls=%d", root.Root, svc.treeRootCalls)
	}
	if len(warnings) != 2 {
		t.Fatalf("expected warnings for retries, got %d", len(warnings))
	}
}

func TestGetTreeRootWithRetry_DoesNotRetryNonTransient(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{},
		root:               inspect.TreeNodeDTO{NodeID: "root-id"},
		treeRootErrs:       []error{errors.New("boom")},
	}
	c := NewController(context.Background(), svc)
	_, warnings, err := c.getTreeRootWithRetry("0x1", inspect.InspectModeUIATree)
	if err == nil {
		t.Fatal("expected error")
	}
	if svc.treeRootCalls != 1 {
		t.Fatalf("expected single attempt, got %d", svc.treeRootCalls)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %d", len(warnings))
	}
}
func (f *fakeControllerService) GetNodeDetails(context.Context, inspect.GetNodeDetailsRequest) (inspect.GetNodeDetailsResponse, error) {
	f.nodeDetailsCalls++
	if f.nodeDetailsErr != nil {
		return inspect.GetNodeDetailsResponse{}, f.nodeDetailsErr
	}
	return f.nodeDetailsResp, nil
}
func (f *fakeControllerService) GetNodeChildren(context.Context, inspect.GetNodeChildrenRequest) (inspect.GetNodeChildrenResponse, error) {
	if f.nodeChildrenErr != nil {
		return inspect.GetNodeChildrenResponse{}, f.nodeChildrenErr
	}
	return inspect.GetNodeChildrenResponse{Children: []inspect.TreeNodeDTO{{NodeID: "child-1"}}}, nil
}
func (f *fakeControllerService) SelectNode(context.Context, inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	f.selectCalls++
	if f.selectErr != nil {
		return inspect.SelectNodeResponse{}, f.selectErr
	}
	return inspect.SelectNodeResponse{}, nil
}
func (f *fakeControllerService) HighlightNode(context.Context, inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	f.highlightCalls++
	if f.highlightErr != nil {
		return inspect.HighlightNodeResponse{}, f.highlightErr
	}
	return inspect.HighlightNodeResponse{}, nil
}

func TestSelectWindowSelectsAndHighlightsRoot(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root-id"}}
	c := NewController(context.Background(), svc)
	if _, err := c.SelectWindow("0x1", false); err != nil {
		t.Fatal(err)
	}
	if svc.selectCalls != 1 || svc.highlightCalls != 1 {
		t.Fatalf("expected select/highlight once for root, got select=%d highlight=%d", svc.selectCalls, svc.highlightCalls)
	}
}

func TestSelectWindowReturnsDetailsWhenRootChildrenFail(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root"}},
		root:               inspect.TreeNodeDTO{NodeID: "root-id"},
		nodeChildrenErr:    errors.New("children failed"),
	}
	c := NewController(context.Background(), svc)
	result, err := c.SelectWindow("0x1", false)
	if err != nil {
		t.Fatalf("expected nil fatal error, got: %v", err)
	}
	if result.Root.Root.NodeID == "" || result.Details.ACCPath == "" {
		t.Fatalf("expected root/details to be populated: %+v", result)
	}
	if result.ChildLoadErr == nil {
		t.Fatal("expected ChildLoadErr warning")
	}
}

func TestSelectWindowReturnsDetailsWhenHighlightFails(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root"}},
		root:               inspect.TreeNodeDTO{NodeID: "root-id"},
		highlightErr:       errors.New("highlight failed"),
	}
	c := NewController(context.Background(), svc)
	result, err := c.SelectWindow("0x1", false)
	if err != nil {
		t.Fatalf("expected nil fatal error, got: %v", err)
	}
	if result.Root.Root.NodeID == "" || result.Details.ACCPath == "" {
		t.Fatalf("expected root/details to be populated: %+v", result)
	}
	if result.HighlightErr == nil {
		t.Fatal("expected HighlightErr warning")
	}
}

func TestSelectWindowReturnsDetailsWhenSelectNodeFails(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root"}},
		root:               inspect.TreeNodeDTO{NodeID: "root-id"},
		selectErr:          errors.New("select failed"),
	}
	c := NewController(context.Background(), svc)
	result, err := c.SelectWindow("0x1", false)
	if err != nil {
		t.Fatalf("expected nil fatal error, got: %v", err)
	}
	if result.Root.Root.NodeID == "" || result.Details.ACCPath == "" {
		t.Fatalf("expected root/details to be populated: %+v", result)
	}
	if result.SelectErr == nil {
		t.Fatal("expected SelectErr warning")
	}
}

func TestSelectWindowCollectsRootRetryWarnings(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root"}},
		root:               inspect.TreeNodeDTO{NodeID: "root-id"},
		treeRootErrs:       []error{inspect.ErrTransientFailure, nil},
	}
	c := NewController(context.Background(), svc)
	result, err := c.SelectWindow("0x1", false)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(result.RootRetryWarnings) != 1 {
		t.Fatalf("expected one retry warning, got %d", len(result.RootRetryWarnings))
	}
}

func TestSelectWindowReturnsRootWhenDetailsFail(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{},
		root:               inspect.TreeNodeDTO{NodeID: "root-id"},
		nodeDetailsErr:     errors.New("details failed"),
	}
	c := NewController(context.Background(), svc)
	result, err := c.SelectWindow("0x1", false)
	if err != nil {
		t.Fatalf("expected non-fatal details fallback, got %v", err)
	}
	if result.Root.Root.NodeID != "root-id" {
		t.Fatalf("expected root retained, got %+v", result.Root.Root)
	}
	if result.DetailsErr == nil {
		t.Fatal("expected details error recorded")
	}
	if result.Details.Element.NodeID != "root-id" || result.Details.StatusText == "" {
		t.Fatalf("expected synthesized details, got %+v", result.Details)
	}
}

func TestControllerModeAccessors(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}}
	c := NewController(context.Background(), svc)
	if got := c.Mode(); got != inspect.InspectModeAuto {
		t.Fatalf("default mode=%q", got)
	}
	c.SetMode(inspect.InspectModeWindowTree)
	if got := c.Mode(); got != inspect.InspectModeWindowTree {
		t.Fatalf("mode=%q", got)
	}
	c.SetMode(inspect.InspectMode("bogus"))
	if got := c.Mode(); got != inspect.InspectModeAuto {
		t.Fatalf("invalid mode should normalize to AUTO, got=%q", got)
	}
}

func TestSelectWindowCarriesFallbackModeState(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{nodeDetailsResp: inspect.GetNodeDetailsResponse{ACCPath: "root"}},
		root:               inspect.TreeNodeDTO{NodeID: "root-id"},
		rootState: inspect.InspectModeState{
			RequestedMode: inspect.InspectModeUIATree,
			ActiveMode:    inspect.InspectModeHWNDTree,
			FallbackUsed:  true,
			GuidanceText:  "UIA tree is unavailable. Switch to ACC/MSAA mode to continue inspecting this window.",
		},
	}
	c := NewController(context.Background(), svc)
	c.SetMode(inspect.InspectModeUIATree)
	result, err := c.SelectWindow("0x1", false)
	if err != nil {
		t.Fatalf("expected fallback success, got %v", err)
	}
	if !result.Root.State.FallbackUsed || result.Root.State.ActiveMode != inspect.InspectModeHWNDTree {
		t.Fatalf("expected degraded fallback mode state, got %+v", result.Root.State)
	}
	if result.Root.State.GuidanceText == "" {
		t.Fatalf("expected fallback guidance text, got %+v", result.Root.State)
	}
}

func TestSelectWindowLogsDetailsError(t *testing.T) {
	svc := &fakeControllerService{
		fakeInspectService: fakeInspectService{},
		root:               inspect.TreeNodeDTO{NodeID: "root-id"},
		nodeDetailsErr:     errors.New("details failed"),
	}
	c := NewController(context.Background(), svc)
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(orig) })
	_, err := c.SelectWindow("0x1", false)
	if err != nil {
		t.Fatalf("expected non-fatal details fallback, got %v", err)
	}
	logs := buf.String()
	if !strings.Contains(logs, "inspect_details_start") || !strings.Contains(logs, "inspect_details_err") || !strings.Contains(logs, "inspect_details_fallback_ok") {
		t.Fatalf("expected details logs, got %s", logs)
	}
}
