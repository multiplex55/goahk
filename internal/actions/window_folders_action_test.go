package actions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"goahk/internal/shell/folders"
)

func TestWindowListOpenFolders_RequiresSaveAs(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("window.list_open_folders")

	err := h(ActionContext{
		Context: context.Background(),
		Services: Services{FolderList: func(context.Context) ([]folders.FolderInfo, error) {
			return nil, nil
		}},
	}, Step{Name: "window.list_open_folders", Params: map[string]string{}})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "window.list_open_folders requires save_as" {
		t.Fatalf("err = %q", err.Error())
	}
}

func TestWindowListOpenFolders_EmptyResultWritesJSONArray(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("window.list_open_folders")
	ctx := ActionContext{
		Context: context.Background(),
		Services: Services{FolderList: func(context.Context) ([]folders.FolderInfo, error) {
			return []folders.FolderInfo{}, nil
		}},
		Metadata: map[string]string{},
	}

	if err := h(ctx, Step{Name: "window.list_open_folders", Params: map[string]string{"save_as": "folders"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ctx.Metadata["folders"]; got != "[]" {
		t.Fatalf("metadata = %q, want []", got)
	}
}

func TestWindowListOpenFolders_WritesMetadataJSON(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("window.list_open_folders")
	ctx := ActionContext{
		Context: context.Background(),
		Services: Services{FolderList: func(context.Context) ([]folders.FolderInfo, error) {
			return []folders.FolderInfo{
				{Path: "C:\\Users\\alice\\Projects", Title: "Projects", PID: 51, HWND: "0xA"},
				{Path: "C:\\Users\\alice\\Projects", Title: "Projects duplicate", PID: 52, HWND: "0xB"},
			}, nil
		}},
		Metadata: map[string]string{},
	}

	if err := h(ctx, Step{Name: "window.list_open_folders", Params: map[string]string{"save_as": "folders", "dedupe": "true"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got []map[string]any
	if err := json.Unmarshal([]byte(ctx.Metadata["folders"]), &got); err != nil {
		t.Fatalf("payload is not JSON: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("entry count = %d, want 1", len(got))
	}
	if got[0]["path"] != `C:\Users\alice\Projects` {
		t.Fatalf("entry = %#v", got[0])
	}
}

func TestWindowListOpenFolders_ServiceErrorsPropagate(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("window.list_open_folders")
	boom := errors.New("folder enumerate failed")

	err := h(ActionContext{
		Context:  context.Background(),
		Services: Services{FolderList: func(context.Context) ([]folders.FolderInfo, error) { return nil, boom }},
	}, Step{Name: "window.list_open_folders", Params: map[string]string{"save_as": "folders"}})
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want %v", err, boom)
	}
}
