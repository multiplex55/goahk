package goahk

import (
	"context"
	"errors"
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

	items, err := ctx.Window.List()
	if err != nil {
		t.Fatalf("List err = %v, want nil", err)
	}
	if len(items) != 2 {
		t.Fatalf("List len = %d, want 2", len(items))
	}
	active, err := ctx.Window.Active()
	if err != nil {
		t.Fatalf("Active err = %v, want nil", err)
	}
	if active.Title != "two" {
		t.Fatalf("Active title = %q, want two", active.Title)
	}
	if err := ctx.Window.Activate("notepad"); err != nil {
		t.Fatalf("Activate err = %v, want nil", err)
	}
	title, err := ctx.Window.Title()
	if err != nil {
		t.Fatalf("Title err = %v, want nil", err)
	}
	if title != "active-title" || activated != "notepad" {
		t.Fatalf("Title/Activate = (%q, %q)", title, activated)
	}
}

func TestContextWindow_MissingServiceReturnsSentinelErrors(t *testing.T) {
	t.Parallel()

	ctx := newContext(nil, newAppState())

	if _, err := ctx.Window.List(); !errors.Is(err, ErrWindowServiceUnavailable) {
		t.Fatalf("List err = %v, want ErrWindowServiceUnavailable", err)
	}
	if _, err := ctx.Window.Active(); !errors.Is(err, ErrWindowServiceUnavailable) {
		t.Fatalf("Active err = %v, want ErrWindowServiceUnavailable", err)
	}
	if err := ctx.Window.Activate("notepad"); !errors.Is(err, ErrWindowServiceUnavailable) {
		t.Fatalf("Activate err = %v, want ErrWindowServiceUnavailable", err)
	}
	if _, err := ctx.Window.Bounds(1); !errors.Is(err, ErrWindowServiceUnavailable) {
		t.Fatalf("Bounds err = %v, want ErrWindowServiceUnavailable", err)
	}
	if err := ctx.Window.Move(1, 1, 1); !errors.Is(err, ErrWindowServiceUnavailable) {
		t.Fatalf("Move err = %v, want ErrWindowServiceUnavailable", err)
	}
	if err := ctx.Window.Resize(1, 10, 10); !errors.Is(err, ErrWindowServiceUnavailable) {
		t.Fatalf("Resize err = %v, want ErrWindowServiceUnavailable", err)
	}
	if _, err := ctx.Window.Title(); !errors.Is(err, ErrWindowServiceUnavailable) {
		t.Fatalf("Title err = %v, want ErrWindowServiceUnavailable", err)
	}
}
