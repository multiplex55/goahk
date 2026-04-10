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
}

func (f *fakeInspectService) ListWindows(context.Context, inspect.ListWindowsRequest) (inspect.ListWindowsResponse, error) {
	return inspect.ListWindowsResponse{}, nil
}
func (f *fakeInspectService) InspectWindow(context.Context, inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	return inspect.InspectWindowResponse{}, nil
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
