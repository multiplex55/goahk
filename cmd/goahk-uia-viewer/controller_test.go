package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"goahk/internal/inspect"
)

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
	if svc.refreshReqs[0] != (inspect.RefreshWindowsRequest{Filter: "abc", VisibleOnly: true, TitleOnly: false}) {
		t.Fatalf("bad req: %+v", svc.refreshReqs[0])
	}
}

func TestController_SelectWindow_Pipeline(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}, root: inspect.TreeNodeDTO{NodeID: "root"}}
	c := NewController(context.Background(), svc)
	c.activateOnSelect = true
	if err := c.SelectWindow("0x1"); err != nil {
		t.Fatal(err)
	}
	if svc.activateCalls != 1 || svc.inspectCalls != 1 || svc.treeRootCalls != 1 || svc.nodeDetailsCalls != 1 {
		t.Fatalf("unexpected call counts: %+v", svc)
	}
}

func TestController_SelectNode_Pipeline(t *testing.T) {
	svc := &fakeControllerService{fakeInspectService: fakeInspectService{}}
	c := NewController(context.Background(), svc)
	if err := c.SelectNode("n1"); err != nil {
		t.Fatal(err)
	}
	if svc.selectCalls != 1 || svc.nodeDetailsCalls != 1 || svc.highlightCalls != 1 {
		t.Fatalf("unexpected call counts")
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
	if c.followEnabled {
		t.Fatal("follow should be disabled")
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
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
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
	root inspect.TreeNodeDTO

	activateCalls    int
	inspectCalls     int
	treeRootCalls    int
	nodeDetailsCalls int
	selectCalls      int
	highlightCalls   int
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
	return inspect.GetNodeDetailsResponse{}, nil
}
func (f *fakeControllerService) SelectNode(context.Context, inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	f.selectCalls++
	return inspect.SelectNodeResponse{}, nil
}
func (f *fakeControllerService) HighlightNode(context.Context, inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	f.highlightCalls++
	return inspect.HighlightNodeResponse{}, nil
}
