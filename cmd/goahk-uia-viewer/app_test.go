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
	toggleResp        inspect.ToggleFollowCursorResponse
	inspectWindowReqs []inspect.InspectWindowRequest
	inspectWindowResp inspect.InspectWindowResponse
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
func (f *fakeInspectService) GetNodeChildren(context.Context, inspect.GetNodeChildrenRequest) (inspect.GetNodeChildrenResponse, error) {
	return inspect.GetNodeChildrenResponse{}, nil
}
func (f *fakeInspectService) SelectNode(context.Context, inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	return inspect.SelectNodeResponse{}, nil
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
	return inspect.ClearHighlightResponse{}, nil
}
func (f *fakeInspectService) CopyBestSelector(context.Context, inspect.CopyBestSelectorRequest) (inspect.CopyBestSelectorResponse, error) {
	return inspect.CopyBestSelectorResponse{}, nil
}
func (f *fakeInspectService) GetPatternActions(context.Context, inspect.GetPatternActionsRequest) (inspect.GetPatternActionsResponse, error) {
	return inspect.GetPatternActionsResponse{}, nil
}
func (f *fakeInspectService) InvokePattern(context.Context, inspect.InvokePatternRequest) (inspect.InvokePatternResponse, error) {
	return inspect.InvokePatternResponse{}, nil
}
func (f *fakeInspectService) ToggleFollowCursor(context.Context, inspect.ToggleFollowCursorRequest) (inspect.ToggleFollowCursorResponse, error) {
	return f.toggleResp, nil
}
func (f *fakeInspectService) RefreshWindows(context.Context, inspect.RefreshWindowsRequest) (inspect.RefreshWindowsResponse, error) {
	return inspect.RefreshWindowsResponse{}, nil
}

func TestToggleFollowCursorIdempotentEnableDisable(t *testing.T) {
	svc := &fakeInspectService{underCursorValues: []inspect.TreeNodeDTO{{NodeID: "n1"}}}
	app := NewViewerApp(svc)

	tick := make(chan time.Time, 1)
	app.followTicker = func() <-chan time.Time { return tick }

	resp, err := app.ToggleFollowCursor(context.Background(), inspect.ToggleFollowCursorRequest{Enabled: true})
	if err != nil || !resp.Enabled {
		t.Fatalf("enable failed: %v %#v", err, resp)
	}
	resp, err = app.ToggleFollowCursor(context.Background(), inspect.ToggleFollowCursorRequest{Enabled: true})
	if err != nil || !resp.Enabled {
		t.Fatalf("second enable failed: %v %#v", err, resp)
	}

	resp, err = app.ToggleFollowCursor(context.Background(), inspect.ToggleFollowCursorRequest{Enabled: false})
	if err != nil || resp.Enabled {
		t.Fatalf("disable failed: %v %#v", err, resp)
	}
	resp, err = app.ToggleFollowCursor(context.Background(), inspect.ToggleFollowCursorRequest{Enabled: false})
	if err != nil || resp.Enabled {
		t.Fatalf("second disable failed: %v %#v", err, resp)
	}
}

func TestToggleFollowCursorEmitsOnChangedNodeOnly(t *testing.T) {
	svc := &fakeInspectService{underCursorValues: []inspect.TreeNodeDTO{{NodeID: "a"}, {NodeID: "a"}, {NodeID: "b"}}}
	app := NewViewerApp(svc)

	tick := make(chan time.Time, 4)
	app.followTicker = func() <-chan time.Time { return tick }

	emitted := make([]followCursorEvent, 0)
	var mu sync.Mutex
	app.SetEventEmitter(func(name string, payload any) {
		if name != "inspect:follow-cursor" {
			return
		}
		e, ok := payload.(followCursorEvent)
		if !ok {
			return
		}
		mu.Lock()
		emitted = append(emitted, e)
		mu.Unlock()
	})

	_, _ = app.ToggleFollowCursor(context.Background(), inspect.ToggleFollowCursorRequest{Enabled: true})
	tick <- time.Now()
	tick <- time.Now()
	tick <- time.Now()
	time.Sleep(10 * time.Millisecond)
	_, _ = app.ToggleFollowCursor(context.Background(), inspect.ToggleFollowCursorRequest{Enabled: false})

	mu.Lock()
	defer mu.Unlock()
	if len(emitted) != 2 {
		t.Fatalf("expected 2 emitted events, got %d", len(emitted))
	}
	if emitted[0].Element.NodeID != "a" || emitted[1].Element.NodeID != "b" {
		t.Fatalf("unexpected sequence: %#v", emitted)
	}
}

func TestViewerApp_InspectWindow_ForwardsRequestUnchanged(t *testing.T) {
	svc := &fakeInspectService{
		inspectWindowResp: inspect.InspectWindowResponse{
			Window:     inspect.WindowSummary{HWND: "0x200", Title: "Demo"},
			RootNodeID: "node:root",
		},
	}
	app := NewViewerApp(svc)
	req := inspect.InspectWindowRequest{HWND: "0x200"}

	resp, err := app.InspectWindow(context.Background(), req)
	if err != nil {
		t.Fatalf("InspectWindow returned error: %v", err)
	}
	if got, want := len(svc.inspectWindowReqs), 1; got != want {
		t.Fatalf("expected %d service call, got %d", want, got)
	}
	if got := svc.inspectWindowReqs[0]; got != req {
		t.Fatalf("expected forwarded request %+v, got %+v", req, got)
	}
	if resp != svc.inspectWindowResp {
		t.Fatalf("expected response %+v, got %+v", svc.inspectWindowResp, resp)
	}
}
