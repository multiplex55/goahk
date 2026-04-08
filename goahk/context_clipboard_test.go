package goahk

import (
	"context"
	"testing"

	"goahk/internal/actions"
)

type fakeClipboard struct {
	text          string
	writeCalls    []string
	appendCalls   []string
	prependCalls  []string
	readCallCount int
}

func (f *fakeClipboard) ReadText(context.Context) (string, error) {
	f.readCallCount++
	return f.text, nil
}
func (f *fakeClipboard) WriteText(_ context.Context, text string) error {
	f.text = text
	f.writeCalls = append(f.writeCalls, text)
	return nil
}
func (f *fakeClipboard) AppendText(_ context.Context, text string) error {
	f.text += text
	f.appendCalls = append(f.appendCalls, text)
	return nil
}
func (f *fakeClipboard) PrependText(_ context.Context, text string) error {
	f.text = text + f.text
	f.prependCalls = append(f.prependCalls, text)
	return nil
}

func TestContextClipboard_WrappersDelegateToService(t *testing.T) {
	t.Parallel()

	clip := &fakeClipboard{text: "start"}
	ctx := newContext(&actions.ActionContext{Context: context.Background(), Services: actions.Services{Clipboard: clip}}, newAppState())

	got, err := ctx.Clipboard.ReadText()
	if err != nil || got != "start" {
		t.Fatalf("ReadText() = (%q, %v), want (start, nil)", got, err)
	}
	_ = ctx.Clipboard.WriteText("a")
	_ = ctx.Clipboard.AppendText("b")
	_ = ctx.Clipboard.PrependText("z")

	if clip.text != "zab" {
		t.Fatalf("clipboard final text = %q, want zab", clip.text)
	}
}
