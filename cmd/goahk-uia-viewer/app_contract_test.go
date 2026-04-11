package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"goahk/internal/inspect"
)

type contractCall struct {
	method string
	ctx    context.Context
	req    any
}

type contractService struct {
	err   error
	calls []contractCall
}

func (s *contractService) record(method string, ctx context.Context, req any) {
	s.calls = append(s.calls, contractCall{method: method, ctx: ctx, req: req})
}

func (s *contractService) ListWindows(ctx context.Context, req inspect.ListWindowsRequest) (inspect.ListWindowsResponse, error) {
	s.record("ListWindows", ctx, req)
	return inspect.ListWindowsResponse{}, s.err
}
func (s *contractService) InspectWindow(ctx context.Context, req inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	s.record("InspectWindow", ctx, req)
	return inspect.InspectWindowResponse{RootNodeID: "root-1"}, s.err
}
func (s *contractService) GetTreeRoot(ctx context.Context, req inspect.GetTreeRootRequest) (inspect.GetTreeRootResponse, error) {
	s.record("GetTreeRoot", ctx, req)
	return inspect.GetTreeRootResponse{Root: inspect.TreeNodeDTO{NodeID: "root-1", HasChildren: true}}, s.err
}
func (s *contractService) GetNodeChildren(ctx context.Context, req inspect.GetNodeChildrenRequest) (inspect.GetNodeChildrenResponse, error) {
	s.record("GetNodeChildren", ctx, req)
	return inspect.GetNodeChildrenResponse{ParentNodeID: req.NodeID}, s.err
}
func (s *contractService) SelectNode(ctx context.Context, req inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	s.record("SelectNode", ctx, req)
	return inspect.SelectNodeResponse{Selected: inspect.TreeNodeDTO{NodeID: req.NodeID}}, s.err
}
func (s *contractService) GetNodeDetails(ctx context.Context, req inspect.GetNodeDetailsRequest) (inspect.GetNodeDetailsResponse, error) {
	s.record("GetNodeDetails", ctx, req)
	return inspect.GetNodeDetailsResponse{StatusText: "ok"}, s.err
}
func (s *contractService) GetFocusedElement(ctx context.Context, req inspect.GetFocusedElementRequest) (inspect.GetFocusedElementResponse, error) {
	s.record("GetFocusedElement", ctx, req)
	return inspect.GetFocusedElementResponse{}, s.err
}
func (s *contractService) GetElementUnderCursor(ctx context.Context, req inspect.GetElementUnderCursorRequest) (inspect.GetElementUnderCursorResponse, error) {
	s.record("GetElementUnderCursor", ctx, req)
	return inspect.GetElementUnderCursorResponse{}, s.err
}
func (s *contractService) HighlightNode(ctx context.Context, req inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	s.record("HighlightNode", ctx, req)
	return inspect.HighlightNodeResponse{Highlighted: true}, s.err
}
func (s *contractService) ClearHighlight(ctx context.Context, req inspect.ClearHighlightRequest) (inspect.ClearHighlightResponse, error) {
	s.record("ClearHighlight", ctx, req)
	return inspect.ClearHighlightResponse{Cleared: true}, s.err
}
func (s *contractService) CopyBestSelector(ctx context.Context, req inspect.CopyBestSelectorRequest) (inspect.CopyBestSelectorResponse, error) {
	s.record("CopyBestSelector", ctx, req)
	return inspect.CopyBestSelectorResponse{Selector: "#id"}, s.err
}
func (s *contractService) GetPatternActions(ctx context.Context, req inspect.GetPatternActionsRequest) (inspect.GetPatternActionsResponse, error) {
	s.record("GetPatternActions", ctx, req)
	return inspect.GetPatternActionsResponse{}, s.err
}
func (s *contractService) InvokePattern(ctx context.Context, req inspect.InvokePatternRequest) (inspect.InvokePatternResponse, error) {
	s.record("InvokePattern", ctx, req)
	return inspect.InvokePatternResponse{Invoked: true}, s.err
}
func (s *contractService) ActivateWindow(ctx context.Context, req inspect.ActivateWindowRequest) (inspect.ActivateWindowResponse, error) {
	s.record("ActivateWindow", ctx, req)
	return inspect.ActivateWindowResponse{Activated: true}, s.err
}
func (s *contractService) ToggleFollowCursor(ctx context.Context, req inspect.ToggleFollowCursorRequest) (inspect.ToggleFollowCursorResponse, error) {
	s.record("ToggleFollowCursor", ctx, req)
	return inspect.ToggleFollowCursorResponse{Enabled: req.Enabled}, s.err
}
func (s *contractService) RefreshWindows(ctx context.Context, req inspect.RefreshWindowsRequest) (inspect.RefreshWindowsResponse, error) {
	s.record("RefreshWindows", ctx, req)
	return inspect.RefreshWindowsResponse{Windows: []inspect.WindowSummary{{HWND: "0x1"}}}, s.err
}

func TestViewerApp_WailsBoundMethods_HaveSingleRequestParameter(t *testing.T) {
	t.Parallel()

	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
	responseType := reflect.TypeOf((*error)(nil)).Elem()
	appType := reflect.TypeOf(&ViewerApp{})
	boundMethods := []string{
		"ListWindows",
		"InspectWindow",
		"GetTreeRoot",
		"GetNodeChildren",
		"SelectNode",
		"GetNodeDetails",
		"GetFocusedElement",
		"GetElementUnderCursor",
		"HighlightNode",
		"ClearHighlight",
		"CopyBestSelector",
		"GetPatternActions",
		"InvokePattern",
		"ActivateWindow",
		"ToggleFollowCursor",
		"RefreshWindows",
	}

	for _, methodName := range boundMethods {
		method, ok := appType.MethodByName(methodName)
		if !ok {
			t.Fatalf("missing ViewerApp method %q", methodName)
		}

		if method.Type.NumIn() != 2 {
			t.Fatalf("%s must accept exactly one request argument; got %d parameters", methodName, method.Type.NumIn()-1)
		}
		if method.Type.In(1).Implements(contextType) {
			t.Fatalf("%s must not accept context.Context", methodName)
		}
		if method.Type.NumOut() != 2 || !method.Type.Out(1).Implements(responseType) {
			t.Fatalf("%s must return (response, error)", methodName)
		}
	}
}

func TestViewerApp_MethodPassthroughAndErrorMapping(t *testing.T) {
	t.Parallel()

	t.Run("passthrough", func(t *testing.T) {
		svc := &contractService{}
		app := NewViewerApp(svc)
		runtimeCtx := context.WithValue(context.Background(), struct{}{}, "runtime")
		app.runtimeCtx = runtimeCtx

		tests := []struct {
			name       string
			call       func() error
			method     string
			req        any
			expectsSvc bool
		}{
			{name: "RefreshWindows", call: func() error { _, err := app.RefreshWindows(inspect.RefreshWindowsRequest{Filter: "note"}); return err }, method: "RefreshWindows", req: inspect.RefreshWindowsRequest{Filter: "note"}, expectsSvc: true},
			{name: "ListWindows", call: func() error { _, err := app.ListWindows(inspect.ListWindowsRequest{}); return err }, method: "ListWindows", expectsSvc: true, req: inspect.ListWindowsRequest{}},
			{name: "InspectWindow", call: func() error { _, err := app.InspectWindow(inspect.InspectWindowRequest{HWND: "0x1"}); return err }, method: "InspectWindow", expectsSvc: true, req: inspect.InspectWindowRequest{HWND: "0x1"}},
			{name: "GetTreeRoot", call: func() error { _, err := app.GetTreeRoot(inspect.GetTreeRootRequest{HWND: "0x1"}); return err }, method: "GetTreeRoot", expectsSvc: true, req: inspect.GetTreeRootRequest{HWND: "0x1"}},
			{name: "GetNodeChildren", call: func() error { _, err := app.GetNodeChildren(inspect.GetNodeChildrenRequest{NodeID: "n1"}); return err }, method: "GetNodeChildren", expectsSvc: true, req: inspect.GetNodeChildrenRequest{NodeID: "n1"}},
			{name: "SelectNode", call: func() error { _, err := app.SelectNode(inspect.SelectNodeRequest{NodeID: "n1"}); return err }, method: "SelectNode", expectsSvc: true, req: inspect.SelectNodeRequest{NodeID: "n1"}},
			{name: "GetNodeDetails", call: func() error { _, err := app.GetNodeDetails(inspect.GetNodeDetailsRequest{NodeID: "n1"}); return err }, method: "GetNodeDetails", expectsSvc: true, req: inspect.GetNodeDetailsRequest{NodeID: "n1"}},
			{name: "GetFocusedElement", call: func() error { _, err := app.GetFocusedElement(inspect.GetFocusedElementRequest{}); return err }, method: "GetFocusedElement", expectsSvc: true, req: inspect.GetFocusedElementRequest{}},
			{name: "GetElementUnderCursor", call: func() error { _, err := app.GetElementUnderCursor(inspect.GetElementUnderCursorRequest{}); return err }, method: "GetElementUnderCursor", expectsSvc: true, req: inspect.GetElementUnderCursorRequest{}},
			{name: "HighlightNode", call: func() error { _, err := app.HighlightNode(inspect.HighlightNodeRequest{NodeID: "n1"}); return err }, method: "HighlightNode", expectsSvc: true, req: inspect.HighlightNodeRequest{NodeID: "n1"}},
			{name: "ClearHighlight", call: func() error { _, err := app.ClearHighlight(inspect.ClearHighlightRequest{}); return err }, method: "ClearHighlight", expectsSvc: true, req: inspect.ClearHighlightRequest{}},
			{name: "GetPatternActions", call: func() error {
				_, err := app.GetPatternActions(inspect.GetPatternActionsRequest{NodeID: "n1"})
				return err
			}, method: "GetPatternActions", expectsSvc: true, req: inspect.GetPatternActionsRequest{NodeID: "n1"}},
			{name: "InvokePattern", call: func() error {
				_, err := app.InvokePattern(inspect.InvokePatternRequest{NodeID: "n1", Action: "invoke"})
				return err
			}, method: "InvokePattern", expectsSvc: true, req: inspect.InvokePatternRequest{NodeID: "n1", Action: "invoke"}},
			{name: "ActivateWindow", call: func() error { _, err := app.ActivateWindow(inspect.ActivateWindowRequest{HWND: "0x1"}); return err }, method: "ActivateWindow", expectsSvc: true, req: inspect.ActivateWindowRequest{HWND: "0x1"}},
			{name: "ToggleFollowCursor", call: func() error {
				_, err := app.ToggleFollowCursor(inspect.ToggleFollowCursorRequest{Enabled: true})
				return err
			}, method: "ToggleFollowCursor", req: inspect.ToggleFollowCursorRequest{Enabled: true}, expectsSvc: false},
			{name: "CopyBestSelector", call: func() error {
				_, err := app.CopyBestSelector(inspect.CopyBestSelectorRequest{NodeID: "n1"})
				return err
			}, method: "CopyBestSelector", expectsSvc: true, req: inspect.CopyBestSelectorRequest{NodeID: "n1"}},
		}

		expectedCalls := 0
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				before := len(svc.calls)
				if err := tc.call(); err != nil {
					t.Fatalf("%s returned error: %v", tc.name, err)
				}
				if tc.expectsSvc {
					expectedCalls++
					if got := len(svc.calls); got != expectedCalls {
						t.Fatalf("expected %d service calls, got %d", expectedCalls, got)
					}
					last := svc.calls[len(svc.calls)-1]
					if last.method != tc.method {
						t.Fatalf("expected method %s got %s", tc.method, last.method)
					}
					if !reflect.DeepEqual(last.req, tc.req) {
						t.Fatalf("expected req %#v got %#v", tc.req, last.req)
					}
					if last.ctx != runtimeCtx {
						t.Fatalf("expected runtime context to be forwarded")
					}
					return
				}
				if got := len(svc.calls); got != before {
					t.Fatalf("expected no immediate service call for %s", tc.name)
				}
			})
		}
	})

	t.Run("error_passthrough", func(t *testing.T) {
		want := errors.New("service boom")
		app := NewViewerApp(&contractService{err: want})
		_, err := app.RefreshWindows(inspect.RefreshWindowsRequest{})
		if !errors.Is(err, want) {
			t.Fatalf("expected forwarded error, got %v", err)
		}
	})
}

func TestViewerApp_WindowSelectionLoggingAndErrorPassthrough(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		call             func(app *ViewerApp) error
		wantRequestLog   string
		wantResponseLog  string
		wantErrorLog     string
		serviceErr       error
		wantForwardedErr error
	}{
		{
			name: "InspectWindow success",
			call: func(app *ViewerApp) error {
				_, err := app.InspectWindow(inspect.InspectWindowRequest{HWND: "0x44"})
				return err
			},
			wantRequestLog:  "method=InspectWindow phase=request hwnd=0x44",
			wantResponseLog: "method=InspectWindow phase=response hwnd=0x44 rootNodeID=root-1",
		},
		{
			name: "InspectWindow error",
			call: func(app *ViewerApp) error {
				_, err := app.InspectWindow(inspect.InspectWindowRequest{HWND: "0x44"})
				return err
			},
			wantRequestLog:   "method=InspectWindow phase=request hwnd=0x44",
			wantErrorLog:     "method=InspectWindow phase=error hwnd=0x44 errorType=ErrProviderActionUnsupported",
			serviceErr:       inspect.ErrProviderActionUnsupported,
			wantForwardedErr: inspect.ErrProviderActionUnsupported,
		},
		{
			name: "GetTreeRoot success",
			call: func(app *ViewerApp) error {
				_, err := app.GetTreeRoot(inspect.GetTreeRootRequest{HWND: "0x21", Refresh: true})
				return err
			},
			wantRequestLog:  "method=GetTreeRoot phase=request hwnd=0x21 refresh=true",
			wantResponseLog: "method=GetTreeRoot phase=response hwnd=0x21 nodeID=root-1",
		},
		{
			name: "GetTreeRoot error",
			call: func(app *ViewerApp) error {
				_, err := app.GetTreeRoot(inspect.GetTreeRootRequest{HWND: "0x21"})
				return err
			},
			wantRequestLog:   "method=GetTreeRoot phase=request hwnd=0x21 refresh=false",
			wantErrorLog:     "method=GetTreeRoot phase=error hwnd=0x21 errorType=ErrStaleCache",
			serviceErr:       inspect.ErrStaleCache,
			wantForwardedErr: inspect.ErrStaleCache,
		},
		{
			name: "GetNodeDetails success",
			call: func(app *ViewerApp) error {
				_, err := app.GetNodeDetails(inspect.GetNodeDetailsRequest{NodeID: "node-1"})
				return err
			},
			wantRequestLog:  "method=GetNodeDetails phase=request nodeID=node-1",
			wantResponseLog: "method=GetNodeDetails phase=response nodeID=node-1",
		},
		{
			name: "GetNodeDetails error",
			call: func(app *ViewerApp) error {
				_, err := app.GetNodeDetails(inspect.GetNodeDetailsRequest{NodeID: "node-1"})
				return err
			},
			wantRequestLog:   "method=GetNodeDetails phase=request nodeID=node-1",
			wantErrorLog:     "method=GetNodeDetails phase=error nodeID=node-1 errorType=ErrProviderActionUnsupported",
			serviceErr:       inspect.ErrProviderActionUnsupported,
			wantForwardedErr: inspect.ErrProviderActionUnsupported,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var debugLogs []string
			var errorLogs []string

			app := NewViewerApp(&contractService{err: tc.serviceErr})
			app.logDebugf = func(_ context.Context, format string, args ...any) {
				debugLogs = append(debugLogs, renderLog(format, args...))
			}
			app.logErrorf = func(_ context.Context, format string, args ...any) {
				errorLogs = append(errorLogs, renderLog(format, args...))
			}

			err := tc.call(app)
			if tc.wantForwardedErr != nil {
				if !errors.Is(err, tc.wantForwardedErr) {
					t.Fatalf("expected forwarded error %v, got %v", tc.wantForwardedErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assertContainsLog(t, debugLogs, tc.wantRequestLog)
			if tc.wantResponseLog != "" {
				assertContainsLog(t, debugLogs, tc.wantResponseLog)
			}
			if tc.wantErrorLog != "" {
				assertContainsLog(t, errorLogs, tc.wantErrorLog)
			}
		})
	}
}

func renderLog(format string, args ...any) string {
	return strings.TrimSpace(fmt.Sprintf(format, args...))
}

func assertContainsLog(t *testing.T, logs []string, needle string) {
	t.Helper()
	for _, line := range logs {
		if strings.Contains(line, needle) {
			return
		}
	}
	t.Fatalf("did not find log containing %q in %v", needle, logs)
}
