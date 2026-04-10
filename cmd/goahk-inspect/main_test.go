package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

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
