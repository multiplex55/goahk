package goahk

import (
	"context"
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

	_ = ctx.Input.SendText("hello")
	_ = ctx.Input.SendKeys("a", "b")
	_ = ctx.Input.SendChord("ctrl", "s")
	_ = ctx.Input.MouseMoveAbsolute(1, 2)
	_ = ctx.Input.MouseMoveRelative(3, 4)
	_, _ = ctx.Input.MousePosition()
	_ = ctx.Input.MouseButtonDown("left")
	_ = ctx.Input.MouseButtonUp("left")
	_ = ctx.Input.MouseClick("left")
	_ = ctx.Input.MouseDoubleClick("left")
	_ = ctx.Input.MouseWheel(120)
	_ = ctx.Input.MouseDrag("left", 1, 2, 3, 4)
	_ = ctx.Input.Paste("paste-me")

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
