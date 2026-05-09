package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"goahk/internal/inspect"
	"goahk/internal/uia"
	"goahk/internal/window"
)

func TestRun_WindowActiveRoutes(t *testing.T) {
	called := false
	d := deps{
		window: windowProviderFunc{
			active: func(context.Context) (window.Info, error) {
				called = true
				return window.Info{HWND: 0x1, Title: "Editor", Active: true}, nil
			},
			list: func(context.Context) ([]window.Info, error) { return nil, nil },
		},
		uia: uiaProviderFunc{},
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"window", "active"}, &out, &bytes.Buffer{}, d); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !called {
		t.Fatal("expected window active route to be called")
	}
	if !strings.Contains(out.String(), "HWND: 0x1") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRun_UIATreeRequiresFlag(t *testing.T) {
	d := deps{}
	err := run(context.Background(), []string{"uia", "tree"}, &bytes.Buffer{}, &bytes.Buffer{}, d)
	if err == nil || !strings.Contains(err.Error(), "requires --active-window") {
		t.Fatalf("expected active-window validation error, got %v", err)
	}
}

func TestRun_ParsesJSONFormatWindowAndUIA(t *testing.T) {
	d := deps{
		window: windowProviderFunc{
			active: func(context.Context) (window.Info, error) {
				return window.Info{HWND: 0x2, Title: "Term", Class: "ConsoleWindowClass", PID: 42}, nil
			},
			list: func(context.Context) ([]window.Info, error) {
				return []window.Info{{HWND: 0x2, Title: "Term", Active: true}}, nil
			},
		},
		uia: uiaProviderFunc{
			under: func(context.Context) (uia.Element, error) {
				name := "Submit"
				return uia.Element{ID: "u-1", Name: &name}, nil
			},
		},
	}
	cases := []struct {
		args []string
		key  string
	}{
		{args: []string{"--format", "json", "window", "active"}, key: "Title"},
		{args: []string{"--format", "json", "window", "list"}, key: "Title"},
		{args: []string{"--format", "json", "uia", "under-cursor"}, key: "element"},
	}
	for _, tc := range cases {
		var out bytes.Buffer
		if err := run(context.Background(), tc.args, &out, &bytes.Buffer{}, d); err != nil {
			t.Fatalf("run(%v) error = %v", tc.args, err)
		}
		var decoded any
		if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
			t.Fatalf("json unmarshal(%v): %v", tc.args, err)
		}
		if !strings.Contains(out.String(), tc.key) {
			t.Fatalf("expected key %q in output %q", tc.key, out.String())
		}
	}
}

func TestRun_WindowActive_PartialProcessDataFallbackFields(t *testing.T) {
	d := deps{
		window: windowProviderFunc{
			active: func(context.Context) (window.Info, error) {
				return window.Info{
					HWND:              0x10,
					Title:             "Partial",
					Class:             "Widget",
					PID:               99,
					ProcessPathStatus: "open_failed",
					ProcessPathError:  "OpenProcess(99): access denied",
					Active:            true,
				}, nil
			},
		},
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"window", "active"}, &out, &bytes.Buffer{}, d); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	text := out.String()
	for _, want := range []string{
		"ProcessPath: (missing)",
		"ProcessPathStatus: open_failed",
		"ProcessPathError: OpenProcess(99): access denied",
		"Visible: (unknown)",
		"Rect: (unknown)",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in output %q", want, text)
		}
	}
}

func TestRun_UIACopyBestSelectorFlag(t *testing.T) {
	name := "Submit"
	controlType := "Button"
	automationID := "submitBtn"
	d := deps{
		uia: uiaProviderFunc{
			under: func(context.Context) (uia.Element, error) {
				return uia.Element{Name: &name, ControlType: &controlType, AutomationID: &automationID}, nil
			},
		},
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"uia", "under-cursor", "--copy-best-selector"}, &out, &bytes.Buffer{}, d); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), `BestSelectorJSON: {"automationId":"submitBtn"}`) {
		t.Fatalf("expected best selector in output, got %q", out.String())
	}
}

func TestRun_WindowJSON_IncludesGeometryAndStateFieldsWhenAvailable(t *testing.T) {
	visible := true
	minimized := false
	maximized := true
	d := deps{
		window: windowProviderFunc{
			active: func(context.Context) (window.Info, error) {
				return window.Info{
					HWND:              0x20,
					Title:             "Editor",
					Class:             "Notepad",
					PID:               777,
					Exe:               "notepad.exe",
					ProcessPath:       `C:\\Windows\\System32\\notepad.exe`,
					ProcessPathStatus: "ok",
					Visible:           &visible,
					Minimized:         &minimized,
					Maximized:         &maximized,
					Rect:              &window.Rect{Left: 10, Top: 20, Right: 210, Bottom: 320},
				}, nil
			},
		},
	}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"--format", "json", "window", "active"}, &out, &bytes.Buffer{}, d); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	body := out.String()
	for _, want := range []string{`"Visible": true`, `"Minimized": false`, `"Maximized": true`, `"ProcessPathStatus": "ok"`, `"Rect"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %q in JSON output %q", want, body)
		}
	}
}

func TestParseGlobal_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantFmt string
		want    []string
		wantErr string
	}{
		{name: "default", args: []string{"window", "active"}, wantFmt: "text", want: []string{"window", "active"}},
		{name: "format flag", args: []string{"--format", "json", "window", "active"}, wantFmt: "json", want: []string{"window", "active"}},
		{name: "format equals", args: []string{"--format=text", "uia", "focused"}, wantFmt: "text", want: []string{"uia", "focused"}},
		{name: "missing value", args: []string{"--format"}, wantErr: "requires a value"},
		{name: "unsupported", args: []string{"--format=yaml", "window", "list"}, wantErr: "unsupported format"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := "text"
			got, err := parseGlobal(tt.args, &format)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected err containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseGlobal err: %v", err)
			}
			if format != tt.wantFmt {
				t.Fatalf("format=%q want %q", format, tt.wantFmt)
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("args=%v want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultDeps_AreOperationalProviders(t *testing.T) {
	d := defaultDeps()
	if _, ok := d.window.(osWindowProvider); !ok {
		t.Fatalf("window provider type = %T, want osWindowProvider", d.window)
	}
	if _, ok := d.uia.(*uia.OSInspectProvider); !ok {
		t.Fatalf("uia provider type = %T, want *uia.OSInspectProvider", d.uia)
	}
	if d.inspect == nil {
		t.Fatal("inspect service is nil")
	}
}

func TestRun_UIADump_ParsesAndRoutes(t *testing.T) {
	called := false
	d := deps{inspect: inspectServiceFunc{dumpTree: func(_ context.Context, req inspect.DumpTreeRequest) (inspect.DumpTreeResponse, error) {
		called = true
		if req.HWND != "0x10" || req.Mode != inspect.InspectModeUIAOnly || req.Depth != 2 {
			t.Fatalf("unexpected req: %+v", req)
		}
		return inspect.DumpTreeResponse{Text: "Window \"Calculator\""}, nil
	}}}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"uia", "dump", "--hwnd", "0x10", "--mode", "UIA_ONLY", "--depth", "2"}, &out, &bytes.Buffer{}, d); err != nil {
		t.Fatalf("run error: %v", err)
	}
	if !called {
		t.Fatal("expected dump route call")
	}
	if !strings.Contains(out.String(), "MODE=UIA_ONLY") {
		t.Fatalf("missing mode header in %q", out.String())
	}
}

func TestParseDumpFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
		want    inspect.DumpTreeRequest
	}{
		{name: "default auto", args: []string{"--hwnd", "0x1"}, want: inspect.DumpTreeRequest{HWND: "0x1", Mode: inspect.InspectModeAuto, Depth: 3}},
		{name: "window tree alias", args: []string{"--hwnd", "0x1", "--mode", "acc_only", "--depth", "1"}, want: inspect.DumpTreeRequest{HWND: "0x1", Mode: inspect.InspectModeWindowTree, Depth: 1}},
		{name: "missing hwnd", args: []string{"--mode", "AUTO"}, wantErr: "requires --hwnd"},
		{name: "invalid mode", args: []string{"--hwnd", "0x1", "--mode", "BAD"}, wantErr: "unsupported mode"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDumpFlags(tt.args)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected err %q got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDumpFlags err: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v want %+v", got, tt.want)
			}
		})
	}
}

func TestRun_UIADump_ModeReportingAndDepthPassThrough(t *testing.T) {
	d := deps{inspect: inspectServiceFunc{dumpTree: func(_ context.Context, req inspect.DumpTreeRequest) (inspect.DumpTreeResponse, error) {
		if req.Depth != 4 {
			t.Fatalf("depth=%d", req.Depth)
		}
		return inspect.DumpTreeResponse{Text: "Root \"Main\"\n  Child \"Item\""}, nil
	}}}
	var out bytes.Buffer
	if err := run(context.Background(), []string{"uia", "dump", "--hwnd", "0x2", "--mode", "AUTO", "--depth", "4"}, &out, &bytes.Buffer{}, d); err != nil {
		t.Fatalf("run error: %v", err)
	}
	text := out.String()
	if !strings.Contains(text, "HWND=0x2 MODE=AUTO") || !strings.Contains(text, "Child \"Item\"") {
		t.Fatalf("unexpected output %q", text)
	}
}

func TestMapOpError_MapsPlatformAndBackendErrors(t *testing.T) {
	tests := []struct {
		op   string
		err  error
		want string
	}{
		{op: "window list", err: window.ErrUnsupportedPlatform, want: "window list: unsupported platform"},
		{op: "uia focused", err: uia.ErrUnsupportedPlatform, want: "uia focused: unsupported platform"},
		{op: "uia focused", err: uia.ErrInspectUnavailable, want: "uia focused: ui automation backend unavailable"},
		{op: "uia focused", err: errors.New("boom"), want: "uia focused: boom"},
	}
	for _, tt := range tests {
		got := mapOpError(tt.op, tt.err)
		if got == nil || got.Error() != tt.want {
			t.Fatalf("mapOpError(%q, %v) = %v, want %q", tt.op, tt.err, got, tt.want)
		}
	}
}

type windowProviderFunc struct {
	active func(context.Context) (window.Info, error)
	list   func(context.Context) ([]window.Info, error)
}

func (f windowProviderFunc) Active(ctx context.Context) (window.Info, error) {
	if f.active == nil {
		return window.Info{}, nil
	}
	return f.active(ctx)
}
func (f windowProviderFunc) List(ctx context.Context) ([]window.Info, error) {
	if f.list == nil {
		return nil, nil
	}
	return f.list(ctx)
}

type uiaProviderFunc struct {
	focused func(context.Context) (uia.Element, error)
	under   func(context.Context) (uia.Element, error)
	tree    func(context.Context, int) (*uia.Node, error)
}

type inspectServiceFunc struct {
	dumpTree func(context.Context, inspect.DumpTreeRequest) (inspect.DumpTreeResponse, error)
}

func (f inspectServiceFunc) DumpTree(ctx context.Context, req inspect.DumpTreeRequest) (inspect.DumpTreeResponse, error) {
	if f.dumpTree == nil {
		return inspect.DumpTreeResponse{}, nil
	}
	return f.dumpTree(ctx, req)
}

func (inspectServiceFunc) ListWindows(context.Context, inspect.ListWindowsRequest) (inspect.ListWindowsResponse, error) {
	return inspect.ListWindowsResponse{}, nil
}
func (inspectServiceFunc) InspectWindow(context.Context, inspect.InspectWindowRequest) (inspect.InspectWindowResponse, error) {
	return inspect.InspectWindowResponse{}, nil
}
func (inspectServiceFunc) GetTreeRoot(context.Context, inspect.GetTreeRootRequest) (inspect.GetTreeRootResponse, error) {
	return inspect.GetTreeRootResponse{}, nil
}
func (inspectServiceFunc) GetNodeChildren(context.Context, inspect.GetNodeChildrenRequest) (inspect.GetNodeChildrenResponse, error) {
	return inspect.GetNodeChildrenResponse{}, nil
}
func (inspectServiceFunc) SelectNode(context.Context, inspect.SelectNodeRequest) (inspect.SelectNodeResponse, error) {
	return inspect.SelectNodeResponse{}, nil
}
func (inspectServiceFunc) GetNodeDetails(context.Context, inspect.GetNodeDetailsRequest) (inspect.GetNodeDetailsResponse, error) {
	return inspect.GetNodeDetailsResponse{}, nil
}
func (inspectServiceFunc) GetFocusedElement(context.Context, inspect.GetFocusedElementRequest) (inspect.GetFocusedElementResponse, error) {
	return inspect.GetFocusedElementResponse{}, nil
}
func (inspectServiceFunc) GetElementUnderCursor(context.Context, inspect.GetElementUnderCursorRequest) (inspect.GetElementUnderCursorResponse, error) {
	return inspect.GetElementUnderCursorResponse{}, nil
}
func (inspectServiceFunc) HighlightNode(context.Context, inspect.HighlightNodeRequest) (inspect.HighlightNodeResponse, error) {
	return inspect.HighlightNodeResponse{}, nil
}
func (inspectServiceFunc) ClearHighlight(context.Context, inspect.ClearHighlightRequest) (inspect.ClearHighlightResponse, error) {
	return inspect.ClearHighlightResponse{}, nil
}
func (inspectServiceFunc) CopyBestSelector(context.Context, inspect.CopyBestSelectorRequest) (inspect.CopyBestSelectorResponse, error) {
	return inspect.CopyBestSelectorResponse{}, nil
}
func (inspectServiceFunc) GetPatternActions(context.Context, inspect.GetPatternActionsRequest) (inspect.GetPatternActionsResponse, error) {
	return inspect.GetPatternActionsResponse{}, nil
}
func (inspectServiceFunc) InvokePattern(context.Context, inspect.InvokePatternRequest) (inspect.InvokePatternResponse, error) {
	return inspect.InvokePatternResponse{}, nil
}
func (inspectServiceFunc) ActivateWindow(context.Context, inspect.ActivateWindowRequest) (inspect.ActivateWindowResponse, error) {
	return inspect.ActivateWindowResponse{}, nil
}
func (inspectServiceFunc) ToggleFollowCursor(context.Context, inspect.ToggleFollowCursorRequest) (inspect.ToggleFollowCursorResponse, error) {
	return inspect.ToggleFollowCursorResponse{}, nil
}
func (inspectServiceFunc) PauseFollowCursor(context.Context, inspect.PauseFollowCursorRequest) (inspect.PauseFollowCursorResponse, error) {
	return inspect.PauseFollowCursorResponse{}, nil
}
func (inspectServiceFunc) ResumeFollowCursor(context.Context, inspect.ResumeFollowCursorRequest) (inspect.ResumeFollowCursorResponse, error) {
	return inspect.ResumeFollowCursorResponse{}, nil
}
func (inspectServiceFunc) LockFollowCursor(context.Context, inspect.LockFollowCursorRequest) (inspect.LockFollowCursorResponse, error) {
	return inspect.LockFollowCursorResponse{}, nil
}
func (inspectServiceFunc) UnlockFollowCursor(context.Context, inspect.UnlockFollowCursorRequest) (inspect.UnlockFollowCursorResponse, error) {
	return inspect.UnlockFollowCursorResponse{}, nil
}
func (inspectServiceFunc) RefreshWindows(context.Context, inspect.RefreshWindowsRequest) (inspect.RefreshWindowsResponse, error) {
	return inspect.RefreshWindowsResponse{}, nil
}
func (inspectServiceFunc) RefreshTreeRoot(context.Context, inspect.RefreshTreeRootRequest) (inspect.RefreshTreeRootResponse, error) {
	return inspect.RefreshTreeRootResponse{}, nil
}
func (inspectServiceFunc) RefreshNodeChildren(context.Context, inspect.RefreshNodeChildrenRequest) (inspect.RefreshNodeChildrenResponse, error) {
	return inspect.RefreshNodeChildrenResponse{}, nil
}
func (inspectServiceFunc) RefreshNodeDetails(context.Context, inspect.RefreshNodeDetailsRequest) (inspect.RefreshNodeDetailsResponse, error) {
	return inspect.RefreshNodeDetailsResponse{}, nil
}
func (inspectServiceFunc) GetDiagnostics(context.Context, inspect.GetDiagnosticsRequest) (inspect.GetDiagnosticsResponse, error) {
	return inspect.GetDiagnosticsResponse{}, nil
}

func (f uiaProviderFunc) Focused(ctx context.Context) (uia.Element, error) {
	if f.focused == nil {
		return uia.Element{}, nil
	}
	return f.focused(ctx)
}
func (f uiaProviderFunc) UnderCursor(ctx context.Context) (uia.Element, error) {
	if f.under == nil {
		return uia.Element{}, nil
	}
	return f.under(ctx)
}
func (f uiaProviderFunc) ActiveWindowTree(ctx context.Context, depth int) (*uia.Node, error) {
	if f.tree == nil {
		return nil, nil
	}
	return f.tree(ctx, depth)
}
