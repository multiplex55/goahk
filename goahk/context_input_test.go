package goahk

import (
	"context"
	"errors"
	"testing"

	"goahk/internal/actions"
	"goahk/internal/input"
)

type fakeInput struct {
	texts  []string
	keys   []input.Sequence
	chords []input.Chord
	pos    input.MousePosition
	calls  []string
}

func (f *fakeInput) SendText(_ context.Context, text string, _ input.SendOptions) error {
	f.texts = append(f.texts, text)
	return nil
}
func (f *fakeInput) SendKeys(_ context.Context, seq input.Sequence, _ input.SendOptions) error {
	f.keys = append(f.keys, seq)
	return nil
}
func (f *fakeInput) SendChord(_ context.Context, chord input.Chord, _ input.SendOptions) error {
	f.chords = append(f.chords, chord)
	return nil
}
func (f *fakeInput) MoveAbsolute(_ context.Context, _, _ int) error {
	f.calls = append(f.calls, "move_abs")
	return nil
}
func (f *fakeInput) MoveRelative(_ context.Context, _, _ int) error {
	f.calls = append(f.calls, "move_rel")
	return nil
}
func (f *fakeInput) Position(context.Context) (input.MousePosition, error) {
	f.calls = append(f.calls, "position")
	return f.pos, nil
}
func (f *fakeInput) ButtonDown(_ context.Context, _ string) error {
	f.calls = append(f.calls, "button_down")
	return nil
}
func (f *fakeInput) ButtonUp(_ context.Context, _ string) error {
	f.calls = append(f.calls, "button_up")
	return nil
}
func (f *fakeInput) Click(_ context.Context, _ string) error {
	f.calls = append(f.calls, "click")
	return nil
}
func (f *fakeInput) DoubleClick(_ context.Context, _ string) error {
	f.calls = append(f.calls, "double_click")
	return nil
}
func (f *fakeInput) Wheel(_ context.Context, _ int) error {
	f.calls = append(f.calls, "wheel")
	return nil
}
func (f *fakeInput) Drag(_ context.Context, _ string, _, _, _, _ int) error {
	f.calls = append(f.calls, "drag")
	return nil
}

func TestContextInput_WrappersDelegateToService(t *testing.T) {
	t.Parallel()

	in := &fakeInput{pos: input.MousePosition{X: 7, Y: 9}}
	clip := &fakeClipboard{}
	ctx := newContext(&actions.ActionContext{Context: context.Background(), Services: actions.Services{Input: in, Clipboard: clip}}, newAppState())

	if err := ctx.Input.SendText("hello"); err != nil {
		t.Fatalf("SendText err = %v, want nil", err)
	}
	if err := ctx.Input.SendKeys("a", "b"); err != nil {
		t.Fatalf("SendKeys err = %v, want nil", err)
	}
	if err := ctx.Input.SendChord("ctrl", "s"); err != nil {
		t.Fatalf("SendChord err = %v, want nil", err)
	}
	if err := ctx.Input.MouseMoveAbsolute(1, 2); err != nil {
		t.Fatalf("MouseMoveAbsolute err = %v, want nil", err)
	}
	if err := ctx.Input.MouseMoveRelative(3, 4); err != nil {
		t.Fatalf("MouseMoveRelative err = %v, want nil", err)
	}
	if _, err := ctx.Input.MousePosition(); err != nil {
		t.Fatalf("MousePosition err = %v, want nil", err)
	}
	if err := ctx.Input.MouseButtonDown("left"); err != nil {
		t.Fatalf("MouseButtonDown err = %v, want nil", err)
	}
	if err := ctx.Input.MouseButtonUp("left"); err != nil {
		t.Fatalf("MouseButtonUp err = %v, want nil", err)
	}
	if err := ctx.Input.MouseClick("left"); err != nil {
		t.Fatalf("MouseClick err = %v, want nil", err)
	}
	if err := ctx.Input.MouseDoubleClick("left"); err != nil {
		t.Fatalf("MouseDoubleClick err = %v, want nil", err)
	}
	if err := ctx.Input.MouseWheel(120); err != nil {
		t.Fatalf("MouseWheel err = %v, want nil", err)
	}
	if err := ctx.Input.MouseDrag("left", 1, 2, 3, 4); err != nil {
		t.Fatalf("MouseDrag err = %v, want nil", err)
	}
	if err := ctx.Input.Paste("paste-me"); err != nil {
		t.Fatalf("Paste err = %v, want nil", err)
	}

	if len(in.texts) != 1 || in.texts[0] != "hello" {
		t.Fatalf("SendText calls = %#v", in.texts)
	}
	if len(in.keys) != 1 || len(in.keys[0].Tokens) != 2 {
		t.Fatalf("SendKeys tokens = %#v", in.keys)
	}
	if len(in.chords) != 2 {
		t.Fatalf("SendChord calls = %d, want 2 (including Paste)", len(in.chords))
	}
	if len(in.calls) != 9 {
		t.Fatalf("mouse calls = %v", in.calls)
	}
	if clip.text != "paste-me" {
		t.Fatalf("Paste clipboard text = %q, want paste-me", clip.text)
	}
}

func TestContextInput_MissingServiceReturnsSentinelErrors(t *testing.T) {
	t.Parallel()

	ctx := newContext(nil, newAppState())

	assertInputErr := func(err error, method string) {
		t.Helper()
		if !errors.Is(err, ErrInputServiceUnavailable) {
			t.Fatalf("%s err = %v, want ErrInputServiceUnavailable", method, err)
		}
	}

	assertInputErr(ctx.Input.SendText("hello"), "SendText")
	assertInputErr(ctx.Input.SendKeys("a"), "SendKeys")
	assertInputErr(ctx.Input.SendChord("ctrl", "s"), "SendChord")
	assertInputErr(ctx.Input.MouseMoveAbsolute(1, 2), "MouseMoveAbsolute")
	assertInputErr(ctx.Input.MouseMoveRelative(1, 2), "MouseMoveRelative")
	if _, err := ctx.Input.MousePosition(); !errors.Is(err, ErrInputServiceUnavailable) {
		t.Fatalf("MousePosition err = %v, want ErrInputServiceUnavailable", err)
	}
	assertInputErr(ctx.Input.MouseButtonDown("left"), "MouseButtonDown")
	assertInputErr(ctx.Input.MouseButtonUp("left"), "MouseButtonUp")
	assertInputErr(ctx.Input.MouseClick("left"), "MouseClick")
	assertInputErr(ctx.Input.MouseDoubleClick("left"), "MouseDoubleClick")
	assertInputErr(ctx.Input.MouseWheel(120), "MouseWheel")
	assertInputErr(ctx.Input.MouseDrag("left", 1, 2, 3, 4), "MouseDrag")
	assertInputErr(ctx.Input.Paste("text"), "Paste")
}

func TestInputPaste_WhenClipboardUnavailableReturnsClipboardError(t *testing.T) {
	t.Parallel()

	ctx := newContext(&actions.ActionContext{Context: context.Background(), Services: actions.Services{Input: &fakeInput{}}}, newAppState())
	if err := ctx.Input.Paste("value"); !errors.Is(err, ErrClipboardServiceUnavailable) {
		t.Fatalf("Paste err = %v, want ErrClipboardServiceUnavailable", err)
	}
}
