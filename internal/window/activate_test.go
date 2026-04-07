package window

import (
	"context"
	"errors"
	"testing"
)

type fakeProvider struct {
	windows   []Info
	errEnum   error
	errAct    error
	activated HWND
}

func (f *fakeProvider) EnumerateWindows(context.Context) ([]Info, error) {
	if f.errEnum != nil {
		return nil, f.errEnum
	}
	return f.windows, nil
}

func (f *fakeProvider) ActivateWindow(_ context.Context, hwnd HWND) error {
	if f.errAct != nil {
		return f.errAct
	}
	f.activated = hwnd
	return nil
}

func TestActivateForeground_SelectsFirstMatch(t *testing.T) {
	fp := &fakeProvider{windows: []Info{{HWND: 10, Title: "A"}, {HWND: 20, Title: "A"}}}
	win, err := ActivateForeground(context.Background(), fp, Matcher{TitleExact: "A"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if win.HWND != 10 {
		t.Fatalf("selected=%v want 10", win.HWND)
	}
	if fp.activated != 10 {
		t.Fatalf("activated=%v want 10", fp.activated)
	}
}

func TestActivateForeground_NoMatches(t *testing.T) {
	fp := &fakeProvider{windows: []Info{{HWND: 10, Title: "A"}}}
	_, err := ActivateForeground(context.Background(), fp, Matcher{TitleExact: "B"})
	if !errors.Is(err, ErrNoMatchingWindow) {
		t.Fatalf("expected ErrNoMatchingWindow, got %v", err)
	}
}

func TestActivateForeground_EnumerateError(t *testing.T) {
	errBoom := errors.New("boom")
	fp := &fakeProvider{errEnum: errBoom}
	_, err := ActivateForeground(context.Background(), fp, Matcher{})
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected wrapped enumerate error, got %v", err)
	}
}

func TestResolveTargetWindow_StrictPolicyAmbiguous(t *testing.T) {
	fp := &fakeProvider{windows: []Info{{HWND: 10, Title: "A"}, {HWND: 20, Title: "A"}}}
	_, err := ResolveTargetWindow(context.Background(), fp, Matcher{TitleExact: "A"}, ActivationPolicy{RequireSingleMatch: true})
	if !errors.Is(err, ErrAmbiguousWindow) {
		t.Fatalf("expected ErrAmbiguousWindow, got %v", err)
	}
}

func TestActivateForegroundWithPolicy_StrictSingleMatch(t *testing.T) {
	fp := &fakeProvider{windows: []Info{{HWND: 10, Title: "A"}, {HWND: 20, Title: "B"}}}
	win, err := ActivateForegroundWithPolicy(context.Background(), fp, Matcher{TitleExact: "A"}, ActivationPolicy{RequireSingleMatch: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if win.HWND != 10 {
		t.Fatalf("selected=%v want 10", win.HWND)
	}
	if fp.activated != 10 {
		t.Fatalf("activated=%v want 10", fp.activated)
	}
}
