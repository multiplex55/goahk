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

func TestContextInput_WrappersDelegateToService(t *testing.T) {
	t.Parallel()

	in := &fakeInput{}
	clip := &fakeClipboard{}
	ctx := newContext(&actions.ActionContext{Context: context.Background(), Services: actions.Services{Input: in, Clipboard: clip}}, newAppState())

	_ = ctx.Input.SendText("hello")
	_ = ctx.Input.SendKeys("a", "b")
	_ = ctx.Input.SendChord("ctrl", "s")
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
	if clip.text != "paste-me" {
		t.Fatalf("Paste clipboard text = %q, want paste-me", clip.text)
	}
}
