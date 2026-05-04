package main

import (
	"context"
	"errors"
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
	return inspect.HighlightNodeResponse{}, nil
}
func (f *fakeInspectService) ClearHighlight(context.Context, inspect.ClearHighlightRequest) (inspect.ClearHighlightResponse, error) {
	f.mu.Lock()
	f.clearCalls++
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
	if err := c.SelectWindow("0x1"); err != nil {
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
	c.activateOnSelect = true
	if err := c.SelectWindow("0x1"); err != nil {
		t.Fatal(err)
	}
	if svc.clearCalls == 0 {
		t.Fatal("expected highlight clear on switch")
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
}
func TestController_SetValuePromptAndInvoke(t *testing.T) {
	svc := &fakeInspectService{}
	c := NewController(context.Background(), svc).WithDialogs(&fakeDialogs{value: "abc", ok: true})
	c.selectedNodeID = "node-1"
	_, err := c.InvokeSetValue()
	if err != nil {
		t.Fatal(err)
	}
	if len(svc.invokeReqs) != 1 || svc.invokeReqs[0].Action != "setValue" {
		t.Fatal("missing invoke")
	}
}
func TestController_StatusInteraction_CopyPath(t *testing.T) {
	svc := &fakeInspectService{}
	cb := &fakeClipboard{}
	c := NewController(context.Background(), svc).WithClipboard(cb)
	c.lastACCPath = "a/b"
	got := c.OnStatusInteraction()
	if got != "ACC path copied" || len(cb.copied) != 1 {
		t.Fatalf("got=%q copied=%v", got, cb.copied)
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
}

func (f *fakeControllerService) ActivateWindow(context.Context, inspect.ActivateWindowRequest) (inspect.ActivateWindowResponse, error) {
	f.activateCalls++
	return inspect.ActivateWindowResponse{}, nil
}
func (f *fakeControllerService) InspectWindow(context.Context, inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	f.inspectCalls++
	return inspect.InspectWindowResponse{}, nil
}
func (f *fakeControllerService) GetTreeRoot(context.Context, inspect.GetTreeRootRequest) (inspect.GetTreeRootResponse, error) {
	f.treeRootCalls++
	return inspect.GetTreeRootResponse{Root: f.root}, nil
}
func (f *fakeControllerService) GetNodeDetails(context.Context, inspect.GetNodeDetailsRequest) (inspect.GetNodeDetailsResponse, error) {
	f.nodeDetailsCalls++
	return f.nodeDetailsResp, nil
}
func (f *fakeControllerService) SelectNode(context.Context, inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	f.selectCalls++
	return inspect.SelectNodeResponse{}, nil
}
func (f *fakeControllerService) HighlightNode(context.Context, inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	f.highlightCalls++
	return inspect.HighlightNodeResponse{}, nil
}
