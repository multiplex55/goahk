package actions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"goahk/internal/window"
)

func TestWindowListOpenApplications_HappyPathWritesMetadata(t *testing.T) {
	r := NewRegistry()
	h, ok := r.Lookup("window.list_open_applications")
	if !ok {
		t.Fatal("handler not found")
	}

	ctx := ActionContext{
		Context: context.Background(),
		Services: Services{
			WindowList: func(context.Context) ([]window.Info, error) {
				return []window.Info{
					{Title: "Editor", Exe: "Code.exe", PID: 101, HWND: window.HWND(0x1), Class: "Chrome_WidgetWin_1", Active: true},
					{Title: "Editor", Exe: "Code.exe", PID: 101, HWND: window.HWND(0x2), Class: "Chrome_WidgetWin_1", Active: false},
				}, nil
			},
		},
		Metadata: map[string]string{},
	}
	step := Step{Name: "window.list_open_applications", Params: map[string]string{"save_as": "apps"}}

	if err := h(ctx, step); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw := ctx.Metadata["apps"]
	if raw == "" {
		t.Fatal("metadata key apps missing")
	}
	var got []map[string]any
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("entry count = %d, want 1", len(got))
	}
	if got[0]["title"] != "Editor" || got[0]["exe"] != "Code.exe" {
		t.Fatalf("entry = %#v", got[0])
	}
	if got[0]["hwnd"] != "0x1" {
		t.Fatalf("hwnd = %#v", got[0]["hwnd"])
	}
}

func TestWindowListOpenApplications_RequiresSaveAs(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("window.list_open_applications")

	err := h(ActionContext{
		Context:  context.Background(),
		Services: Services{WindowList: func(context.Context) ([]window.Info, error) { return nil, nil }},
	}, Step{Name: "window.list_open_applications", Params: map[string]string{}})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "window.list_open_applications requires save_as" {
		t.Fatalf("err = %q", err.Error())
	}
}

func TestWindowListOpenApplications_ServiceErrorsPropagate(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("window.list_open_applications")
	boom := errors.New("enumerate failed")

	err := h(ActionContext{
		Context:  context.Background(),
		Services: Services{WindowList: func(context.Context) ([]window.Info, error) { return nil, boom }},
	}, Step{Name: "window.list_open_applications", Params: map[string]string{"save_as": "apps"}})
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want %v", err, boom)
	}
}

func TestWindowListOpenApplications_FiltersEmptyTitleAndSystemWindows(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("window.list_open_applications")
	ctx := ActionContext{
		Context: context.Background(),
		Services: Services{
			WindowList: func(context.Context) ([]window.Info, error) {
				return []window.Info{
					{Title: "", Exe: "Code.exe", PID: 1, HWND: window.HWND(0x1), Class: "Chrome_WidgetWin_1", Active: true},
					{Title: "Program Manager", Exe: "explorer.exe", PID: 2, HWND: window.HWND(0x2), Class: "Progman", Active: true},
					{Title: "Taskbar", Exe: "explorer.exe", PID: 3, HWND: window.HWND(0x3), Class: "Shell_TrayWnd", Active: true},
					{Title: "Browser", Exe: "msedge.exe", PID: 4, HWND: window.HWND(0x4), Class: "Chrome_WidgetWin_1", Active: true},
				}, nil
			},
		},
		Metadata: map[string]string{},
	}

	err := h(ctx, Step{Name: "window.list_open_applications", Params: map[string]string{"save_as": "apps", "include_background": "true", "dedupe_by": "window"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []map[string]any
	if err := json.Unmarshal([]byte(ctx.Metadata["apps"]), &got); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("entry count = %d, want 1", len(got))
	}
	if got[0]["title"] != "Browser" {
		t.Fatalf("entry = %#v", got[0])
	}
}
