package inspect

import (
	"context"
	"errors"
	"testing"
)

var _ Service = service{}
var _ WindowsProvider = (*mockProvider)(nil)

type mockProvider struct {
	invokePatternErr error
	inspectWindowReq *InspectWindowRequest
}

func (m *mockProvider) ListWindows(context.Context, ListWindowsRequest) (ListWindowsResponse, error) {
	return ListWindowsResponse{Windows: []WindowSummary{{HWND: "0x1", Title: "Demo"}}}, nil
}

func (m *mockProvider) InspectWindow(_ context.Context, req InspectWindowRequest) (InspectWindowResponse, error) {
	copied := req
	m.inspectWindowReq = &copied
	return InspectWindowResponse{}, nil
}

func (m *mockProvider) GetTreeRoot(context.Context, GetTreeRootRequest) (GetTreeRootResponse, error) {
	return GetTreeRootResponse{}, nil
}

func (m *mockProvider) GetNodeChildren(context.Context, GetNodeChildrenRequest) (GetNodeChildrenResponse, error) {
	return GetNodeChildrenResponse{}, nil
}

func (m *mockProvider) SelectNode(context.Context, SelectNodeRequest) (SelectNodeResponse, error) {
	return SelectNodeResponse{}, nil
}

func (m *mockProvider) GetFocusedElement(context.Context, GetFocusedElementRequest) (GetFocusedElementResponse, error) {
	return GetFocusedElementResponse{}, nil
}

func (m *mockProvider) GetElementUnderCursor(context.Context, GetElementUnderCursorRequest) (GetElementUnderCursorResponse, error) {
	return GetElementUnderCursorResponse{}, nil
}

func (m *mockProvider) HighlightNode(context.Context, HighlightNodeRequest) (HighlightNodeResponse, error) {
	return HighlightNodeResponse{}, nil
}

func (m *mockProvider) ClearHighlight(context.Context, ClearHighlightRequest) (ClearHighlightResponse, error) {
	return ClearHighlightResponse{}, nil
}

func (m *mockProvider) CopyBestSelector(context.Context, CopyBestSelectorRequest) (CopyBestSelectorResponse, error) {
	return CopyBestSelectorResponse{}, nil
}

func (m *mockProvider) GetPatternActions(context.Context, GetPatternActionsRequest) (GetPatternActionsResponse, error) {
	return GetPatternActionsResponse{}, nil
}

func (m *mockProvider) InvokePattern(context.Context, InvokePatternRequest) (InvokePatternResponse, error) {
	if m.invokePatternErr != nil {
		return InvokePatternResponse{}, m.invokePatternErr
	}
	return InvokePatternResponse{Invoked: true}, nil
}

func (m *mockProvider) ToggleFollowCursor(context.Context, ToggleFollowCursorRequest) (ToggleFollowCursorResponse, error) {
	return ToggleFollowCursorResponse{}, nil
}

func (m *mockProvider) RefreshWindows(context.Context, RefreshWindowsRequest) (RefreshWindowsResponse, error) {
	return RefreshWindowsResponse{}, nil
}

func TestService_ArgumentValidation(t *testing.T) {
	t.Parallel()

	svc := newServiceWithProvider(&mockProvider{})
	tests := []struct {
		name string
		call func() error
	}{
		{"get node children requires node id", func() error {
			_, err := svc.GetNodeChildren(context.Background(), GetNodeChildrenRequest{})
			return err
		}},
		{"select node requires node id", func() error { _, err := svc.SelectNode(context.Background(), SelectNodeRequest{}); return err }},
		{"highlight node requires node id", func() error { _, err := svc.HighlightNode(context.Background(), HighlightNodeRequest{}); return err }},
		{"copy selector requires node id", func() error {
			_, err := svc.CopyBestSelector(context.Background(), CopyBestSelectorRequest{})
			return err
		}},
		{"invoke pattern requires action and node id", func() error { _, err := svc.InvokePattern(context.Background(), InvokePatternRequest{}); return err }},
		{"inspect window requires hwnd", func() error { _, err := svc.InspectWindow(context.Background(), InspectWindowRequest{}); return err }},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.call(); !errors.Is(err, ErrInvalidNodeID) {
				t.Fatalf("expected ErrInvalidNodeID, got %v", err)
			}
		})
	}
}

func TestService_InspectWindow_ForwardsRequestUnchanged(t *testing.T) {
	t.Parallel()

	provider := &mockProvider{}
	svc := newServiceWithProvider(provider)
	req := InspectWindowRequest{HWND: "0x42"}

	if _, err := svc.InspectWindow(context.Background(), req); err != nil {
		t.Fatalf("InspectWindow returned error: %v", err)
	}
	if provider.inspectWindowReq == nil {
		t.Fatalf("expected provider InspectWindow to be called")
	}
	if *provider.inspectWindowReq != req {
		t.Fatalf("expected request %+v, got %+v", req, *provider.inspectWindowReq)
	}
}

func TestService_ProviderErrorMapping(t *testing.T) {
	t.Parallel()

	t.Run("not-supported maps to unsupported pattern action", func(t *testing.T) {
		t.Parallel()
		svc := newServiceWithProvider(&mockProvider{invokePatternErr: ErrProviderActionUnsupported})
		_, err := svc.InvokePattern(context.Background(), InvokePatternRequest{NodeID: "n-1", Action: "invoke"})
		if !errors.Is(err, ErrUnsupportedPatternAction) {
			t.Fatalf("expected ErrUnsupportedPatternAction, got %v", err)
		}
	})

	t.Run("transient maps to standardized transient error", func(t *testing.T) {
		t.Parallel()
		svc := newServiceWithProvider(&mockProvider{invokePatternErr: ErrProviderTransientFailure})
		_, err := svc.InvokePattern(context.Background(), InvokePatternRequest{NodeID: "n-1", Action: "invoke"})
		if !errors.Is(err, ErrTransientFailure) {
			t.Fatalf("expected ErrTransientFailure, got %v", err)
		}
	})
}
