package inspect

import (
	"context"
	"errors"
	"testing"
)

var _ Service = service{}
var _ WindowsProvider = (*contractProvider)(nil)

var errNotFound = errors.New("not found")

type contractProvider struct {
	refreshWindowsResp  RefreshWindowsResponse
	inspectWindowResp   InspectWindowResponse
	getTreeRootResp     GetTreeRootResponse
	getNodeChildrenResp GetNodeChildrenResponse
	selectNodeResp      SelectNodeResponse
	getNodeDetailsResp  GetNodeDetailsResponse
	highlightNodeResp   HighlightNodeResponse
	clearHighlightResp  ClearHighlightResponse
	getPatternResp      GetPatternActionsResponse
	invokePatternResp   InvokePatternResponse
	activateWindowResp  ActivateWindowResponse
	toggleFollowResp    ToggleFollowCursorResponse
	copySelectorResp    CopyBestSelectorResponse

	err error
}

func (p *contractProvider) ListWindows(context.Context, ListWindowsRequest) (ListWindowsResponse, error) {
	return ListWindowsResponse{}, p.err
}
func (p *contractProvider) InspectWindow(context.Context, InspectWindowRequest) (InspectWindowResponse, error) {
	return p.inspectWindowResp, p.err
}
func (p *contractProvider) GetTreeRoot(context.Context, GetTreeRootRequest) (GetTreeRootResponse, error) {
	return p.getTreeRootResp, p.err
}
func (p *contractProvider) GetNodeChildren(context.Context, GetNodeChildrenRequest) (GetNodeChildrenResponse, error) {
	return p.getNodeChildrenResp, p.err
}
func (p *contractProvider) SelectNode(context.Context, SelectNodeRequest) (SelectNodeResponse, error) {
	return p.selectNodeResp, p.err
}
func (p *contractProvider) GetNodeDetails(context.Context, GetNodeDetailsRequest) (GetNodeDetailsResponse, error) {
	return p.getNodeDetailsResp, p.err
}
func (p *contractProvider) GetFocusedElement(context.Context, GetFocusedElementRequest) (GetFocusedElementResponse, error) {
	return GetFocusedElementResponse{}, p.err
}
func (p *contractProvider) GetElementUnderCursor(context.Context, GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error) {
	return GetElementUnderCursorResponse{}, p.err
}
func (p *contractProvider) HighlightNode(context.Context, HighlightNodeRequest) (HighlightNodeResponse, error) {
	return p.highlightNodeResp, p.err
}
func (p *contractProvider) ClearHighlight(context.Context, ClearHighlightRequest) (ClearHighlightResponse, error) {
	return p.clearHighlightResp, p.err
}
func (p *contractProvider) CopyBestSelector(context.Context, CopyBestSelectorRequest) (CopyBestSelectorResponse, error) {
	return p.copySelectorResp, p.err
}
func (p *contractProvider) GetPatternActions(context.Context, GetPatternActionsRequest) (GetPatternActionsResponse, error) {
	return p.getPatternResp, p.err
}
func (p *contractProvider) InvokePattern(context.Context, InvokePatternRequest) (InvokePatternResponse, error) {
	return p.invokePatternResp, p.err
}
func (p *contractProvider) ActivateWindow(context.Context, ActivateWindowRequest) (ActivateWindowResponse, error) {
	return p.activateWindowResp, p.err
}
func (p *contractProvider) ToggleFollowCursor(context.Context, ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error) {
	return p.toggleFollowResp, p.err
}
func (p *contractProvider) RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error) {
	return p.refreshWindowsResp, p.err
}

func TestService_MethodContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		success  func(t *testing.T, svc Service)
		invalid  func(t *testing.T, svc Service)
		notFound func(t *testing.T, svc Service)
	}{
		{name: "RefreshWindows", success: func(t *testing.T, svc Service) {
			_, err := svc.RefreshWindows(context.Background(), RefreshWindowsRequest{Filter: "x", VisibleOnly: true})
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.RefreshWindows(context.Background(), RefreshWindowsRequest{})
			if err != nil {
				t.Fatalf("refresh should accept zero request: %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.RefreshWindows(context.Background(), RefreshWindowsRequest{})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "InspectWindow", success: func(t *testing.T, svc Service) {
			_, err := svc.InspectWindow(context.Background(), InspectWindowRequest{HWND: "0x1"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.InspectWindow(context.Background(), InspectWindowRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.InspectWindow(context.Background(), InspectWindowRequest{HWND: "0x1"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "GetTreeRoot", success: func(t *testing.T, svc Service) {
			_, err := svc.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.GetTreeRoot(context.Background(), GetTreeRootRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.GetTreeRoot(context.Background(), GetTreeRootRequest{HWND: "0x1"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "GetNodeChildren", success: func(t *testing.T, svc Service) {
			_, err := svc.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: "n"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.GetNodeChildren(context.Background(), GetNodeChildrenRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.GetNodeChildren(context.Background(), GetNodeChildrenRequest{NodeID: "n"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "SelectNode", success: func(t *testing.T, svc Service) {
			_, err := svc.SelectNode(context.Background(), SelectNodeRequest{NodeID: "n"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.SelectNode(context.Background(), SelectNodeRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.SelectNode(context.Background(), SelectNodeRequest{NodeID: "n"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "GetNodeDetails", success: func(t *testing.T, svc Service) {
			_, err := svc.GetNodeDetails(context.Background(), GetNodeDetailsRequest{NodeID: "n"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.GetNodeDetails(context.Background(), GetNodeDetailsRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.GetNodeDetails(context.Background(), GetNodeDetailsRequest{NodeID: "n"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "HighlightNode", success: func(t *testing.T, svc Service) {
			_, err := svc.HighlightNode(context.Background(), HighlightNodeRequest{NodeID: "n"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.HighlightNode(context.Background(), HighlightNodeRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.HighlightNode(context.Background(), HighlightNodeRequest{NodeID: "n"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "ClearHighlight", success: func(t *testing.T, svc Service) {
			_, err := svc.ClearHighlight(context.Background(), ClearHighlightRequest{})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.ClearHighlight(context.Background(), ClearHighlightRequest{})
			if err != nil {
				t.Fatal(err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.ClearHighlight(context.Background(), ClearHighlightRequest{})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "GetPatternActions", success: func(t *testing.T, svc Service) {
			_, err := svc.GetPatternActions(context.Background(), GetPatternActionsRequest{NodeID: "n"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.GetPatternActions(context.Background(), GetPatternActionsRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.GetPatternActions(context.Background(), GetPatternActionsRequest{NodeID: "n"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "InvokePattern", success: func(t *testing.T, svc Service) {
			_, err := svc.InvokePattern(context.Background(), InvokePatternRequest{NodeID: "n", Action: "invoke"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.InvokePattern(context.Background(), InvokePatternRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.InvokePattern(context.Background(), InvokePatternRequest{NodeID: "n", Action: "invoke"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "ActivateWindow", success: func(t *testing.T, svc Service) {
			_, err := svc.ActivateWindow(context.Background(), ActivateWindowRequest{HWND: "0x1"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.ActivateWindow(context.Background(), ActivateWindowRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.ActivateWindow(context.Background(), ActivateWindowRequest{HWND: "0x1"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "ToggleFollowCursor", success: func(t *testing.T, svc Service) {
			_, err := svc.ToggleFollowCursor(context.Background(), ToggleFollowCursorRequest{Enabled: true})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.ToggleFollowCursor(context.Background(), ToggleFollowCursorRequest{})
			if err != nil {
				t.Fatalf("toggle should accept zero request: %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.ToggleFollowCursor(context.Background(), ToggleFollowCursorRequest{})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
		{name: "CopyBestSelector", success: func(t *testing.T, svc Service) {
			_, err := svc.CopyBestSelector(context.Background(), CopyBestSelectorRequest{NodeID: "n"})
			if err != nil {
				t.Fatal(err)
			}
		}, invalid: func(t *testing.T, svc Service) {
			_, err := svc.CopyBestSelector(context.Background(), CopyBestSelectorRequest{})
			if !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("want invalid got %v", err)
			}
		}, notFound: func(t *testing.T, svc Service) {
			_, err := svc.CopyBestSelector(context.Background(), CopyBestSelectorRequest{NodeID: "n"})
			if !errors.Is(err, errNotFound) {
				t.Fatalf("want errNotFound got %v", err)
			}
		}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			t.Run("success", func(t *testing.T) {
				t.Parallel()
				svc := newServiceWithProvider(&contractProvider{})
				tc.success(t, svc)
			})
			t.Run("invalid_request", func(t *testing.T) {
				t.Parallel()
				svc := newServiceWithProvider(&contractProvider{})
				tc.invalid(t, svc)
			})
			t.Run("not_found", func(t *testing.T) {
				t.Parallel()
				svc := newServiceWithProvider(&contractProvider{err: errNotFound})
				tc.notFound(t, svc)
			})
		})
	}
}
