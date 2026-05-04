package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"goahk/internal/inspect"
)

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
