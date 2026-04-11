package main

import (
	"context"
	"errors"
	"testing"

	"goahk/internal/inspect"
)

type passthroughService struct {
	err error
}

func (s *passthroughService) ListWindows(context.Context, inspect.ListWindowsRequest) (inspect.ListWindowsResponse, error) {
	return inspect.ListWindowsResponse{}, s.err
}
func (s *passthroughService) InspectWindow(context.Context, inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	return inspect.InspectWindowResponse{RootNodeID: "root-1"}, s.err
}
func (s *passthroughService) GetTreeRoot(context.Context, inspect.GetTreeRootRequest) (inspect.GetTreeRootResponse, error) {
	return inspect.GetTreeRootResponse{Root: inspect.TreeNodeDTO{NodeID: "root-1", HasChildren: true}}, s.err
}
func (s *passthroughService) GetNodeChildren(context.Context, inspect.GetNodeChildrenRequest) (inspect.GetNodeChildrenResponse, error) {
	return inspect.GetNodeChildrenResponse{ParentNodeID: "root-1"}, s.err
}
func (s *passthroughService) SelectNode(context.Context, inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	return inspect.SelectNodeResponse{Selected: inspect.TreeNodeDTO{NodeID: "n1"}}, s.err
}
func (s *passthroughService) GetNodeDetails(context.Context, inspect.GetNodeDetailsRequest) (inspect.GetNodeDetailsResponse, error) {
	return inspect.GetNodeDetailsResponse{StatusText: "ok"}, s.err
}
func (s *passthroughService) GetFocusedElement(context.Context, inspect.GetFocusedElementRequest) (inspect.GetFocusedElementResponse, error) {
	return inspect.GetFocusedElementResponse{}, s.err
}
func (s *passthroughService) GetElementUnderCursor(context.Context, inspect.GetElementUnderCursorRequest) (inspect.GetElementUnderCursorResponse, error) {
	return inspect.GetElementUnderCursorResponse{}, s.err
}
func (s *passthroughService) HighlightNode(context.Context, inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	return inspect.HighlightNodeResponse{Highlighted: true}, s.err
}
func (s *passthroughService) ClearHighlight(context.Context, inspect.ClearHighlightRequest) (inspect.ClearHighlightResponse, error) {
	return inspect.ClearHighlightResponse{Cleared: true}, s.err
}
func (s *passthroughService) CopyBestSelector(context.Context, inspect.CopyBestSelectorRequest) (inspect.CopyBestSelectorResponse, error) {
	return inspect.CopyBestSelectorResponse{Selector: "#id"}, s.err
}
func (s *passthroughService) GetPatternActions(context.Context, inspect.GetPatternActionsRequest) (inspect.GetPatternActionsResponse, error) {
	return inspect.GetPatternActionsResponse{}, s.err
}
func (s *passthroughService) InvokePattern(context.Context, inspect.InvokePatternRequest) (inspect.InvokePatternResponse, error) {
	return inspect.InvokePatternResponse{Invoked: true}, s.err
}
func (s *passthroughService) ActivateWindow(context.Context, inspect.ActivateWindowRequest) (inspect.ActivateWindowResponse, error) {
	return inspect.ActivateWindowResponse{Activated: true}, s.err
}
func (s *passthroughService) ToggleFollowCursor(context.Context, inspect.ToggleFollowCursorRequest) (inspect.ToggleFollowCursorResponse, error) {
	return inspect.ToggleFollowCursorResponse{Enabled: true}, s.err
}
func (s *passthroughService) RefreshWindows(context.Context, inspect.RefreshWindowsRequest) (inspect.RefreshWindowsResponse, error) {
	return inspect.RefreshWindowsResponse{Windows: []inspect.WindowSummary{{HWND: "0x1"}}}, s.err
}

func TestViewerApp_MethodPassthroughAndErrorMapping(t *testing.T) {
	t.Parallel()

	t.Run("passthrough", func(t *testing.T) {
		app := NewViewerApp(&passthroughService{})
		if _, err := app.RefreshWindows(inspect.RefreshWindowsRequest{}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.InspectWindow(inspect.InspectWindowRequest{HWND: "0x1"}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.GetTreeRoot(inspect.GetTreeRootRequest{HWND: "0x1"}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.GetNodeChildren(inspect.GetNodeChildrenRequest{NodeID: "n1"}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.SelectNode(inspect.SelectNodeRequest{NodeID: "n1"}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.GetNodeDetails(inspect.GetNodeDetailsRequest{NodeID: "n1"}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.HighlightNode(inspect.HighlightNodeRequest{NodeID: "n1"}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.ClearHighlight(inspect.ClearHighlightRequest{}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.GetPatternActions(inspect.GetPatternActionsRequest{NodeID: "n1"}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.InvokePattern(inspect.InvokePatternRequest{NodeID: "n1", Action: "invoke"}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.ActivateWindow(inspect.ActivateWindowRequest{HWND: "0x1"}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.ToggleFollowCursor(inspect.ToggleFollowCursorRequest{Enabled: true}); err != nil {
			t.Fatal(err)
		}
		if _, err := app.CopyBestSelector(inspect.CopyBestSelectorRequest{NodeID: "n1"}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("error_passthrough", func(t *testing.T) {
		want := errors.New("service boom")
		app := NewViewerApp(&passthroughService{err: want})
		_, err := app.RefreshWindows(inspect.RefreshWindowsRequest{})
		if !errors.Is(err, want) {
			t.Fatalf("expected forwarded error, got %v", err)
		}
	})
}
