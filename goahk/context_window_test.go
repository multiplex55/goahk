package goahk

import (
	"context"
	"testing"

	"goahk/internal/actions"
	"goahk/internal/window"
)

func TestContextWindow_WrappersDelegateToServices(t *testing.T) {
	t.Parallel()

	listed := []window.Info{{Title: "one"}, {Title: "two", Active: true}}
	activated := ""
	ctx := newContext(&actions.ActionContext{Context: context.Background(), Services: actions.Services{
		WindowList: func(context.Context) ([]window.Info, error) { return listed, nil },
		WindowActivate: func(_ context.Context, matcher string) error {
			activated = matcher
			return nil
		},
		ActiveWindowTitle: func(context.Context) (string, error) { return "active-title", nil },
	}}, newAppState())

	items, _ := ctx.Window.List()
	if len(items) != 2 {
		t.Fatalf("List len = %d, want 2", len(items))
	}
	active, _ := ctx.Window.Active()
	if active.Title != "two" {
		t.Fatalf("Active title = %q, want two", active.Title)
	}
	_ = ctx.Window.Activate("notepad")
	title, _ := ctx.Window.Title()
	if title != "active-title" || activated != "notepad" {
		t.Fatalf("Title/Activate = (%q, %q)", title, activated)
	}
}
