package actions

import (
	"context"
	"errors"
	"testing"
)

func TestWindowCopyActiveTitleToClipboard(t *testing.T) {
	r := NewRegistry()
	h, ok := r.Lookup("window.copy_active_title_to_clipboard")
	if !ok {
		t.Fatal("expected action registration")
	}

	var copied string
	err := h(ActionContext{
		Context: context.Background(),
		Services: Services{
			ActiveWindowTitle: func(context.Context) (string, error) { return "My Active Window", nil },
			ClipboardWrite:    func(_ context.Context, text string) error { copied = text; return nil },
		},
	}, Step{Name: "window.copy_active_title_to_clipboard"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if copied != "My Active Window" {
		t.Fatalf("copied=%q", copied)
	}
}

func TestWindowCopyActiveTitleToClipboard_PropagatesError(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("window.copy_active_title_to_clipboard")
	want := errors.New("active lookup failed")
	err := h(ActionContext{
		Context: context.Background(),
		Services: Services{
			ActiveWindowTitle: func(context.Context) (string, error) { return "", want },
			ClipboardWrite:    func(context.Context, string) error { t.Fatal("should not be called"); return nil },
		},
	}, Step{Name: "window.copy_active_title_to_clipboard"})
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
