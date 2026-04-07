package actions

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeClipboard struct {
	text string
}

func (f *fakeClipboard) ReadText(context.Context) (string, error) { return f.text, nil }
func (f *fakeClipboard) WriteText(_ context.Context, text string) error {
	f.text = text
	return nil
}
func (f *fakeClipboard) AppendText(_ context.Context, suffix string) error {
	f.text += suffix
	return nil
}
func (f *fakeClipboard) PrependText(_ context.Context, prefix string) error {
	f.text = prefix + f.text
	return nil
}

func TestWindowCopyActiveTitleToClipboard(t *testing.T) {
	r := NewRegistry()
	h, ok := r.Lookup("window.copy_active_title_to_clipboard")
	if !ok {
		t.Fatal("handler not found")
	}

	clip := &fakeClipboard{}
	err := h(ActionContext{
		Context: context.Background(),
		Services: Services{
			ActiveWindowTitle: func(context.Context) (string, error) { return "Editor — README.md", nil },
			Clipboard:         clip,
		},
	}, Step{Name: "window.copy_active_title_to_clipboard"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clip.text != "Editor — README.md" {
		t.Fatalf("copied=%q", clip.text)
	}
}

func TestWindowCopyActiveTitleToClipboard_PropagatesError(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("window.copy_active_title_to_clipboard")
	boom := errors.New("title failed")

	err := h(ActionContext{
		Context: context.Background(),
		Services: Services{
			ActiveWindowTitle: func(context.Context) (string, error) { return "", boom },
			Clipboard:         &fakeClipboard{},
		},
	}, Step{Name: "window.copy_active_title_to_clipboard"})
	if !errors.Is(err, boom) {
		t.Fatalf("error=%v want %v", err, boom)
	}
}

func TestWindowActions_MissingServiceErrors(t *testing.T) {
	r := NewRegistry()
	cases := []struct {
		step Step
		ctx  ActionContext
		want string
	}{
		{
			step: Step{Name: "window.activate"},
			ctx:  ActionContext{Context: context.Background()},
			want: "window service unavailable",
		},
		{
			step: Step{Name: "window.copy_active_title_to_clipboard"},
			ctx:  ActionContext{Context: context.Background(), Services: Services{Clipboard: &fakeClipboard{}}},
			want: "window service unavailable",
		},
		{
			step: Step{Name: "window.copy_active_title_to_clipboard"},
			ctx: ActionContext{
				Context: context.Background(),
				Services: Services{
					ActiveWindowTitle: func(context.Context) (string, error) { return "x", nil },
				},
			},
			want: "clipboard service unavailable",
		},
	}
	for _, tc := range cases {
		h, _ := r.Lookup(tc.step.Name)
		err := h(tc.ctx, tc.step)
		if err == nil {
			t.Fatalf("expected error for %s", tc.step.Name)
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s err=%q missing %q", tc.step.Name, err.Error(), tc.want)
		}
		if !strings.Contains(err.Error(), tc.step.Name) {
			t.Fatalf("%s err=%q missing action identity", tc.step.Name, err.Error())
		}
	}
}
