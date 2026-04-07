package actions

import (
	"context"
	"testing"
)

func TestClipboardActions_Semantics(t *testing.T) {
	r := NewRegistry()
	clip := &fakeClipboard{text: "base"}
	ctx := ActionContext{Context: context.Background(), Services: Services{Clipboard: clip}, Metadata: map[string]string{}}

	appendH, _ := r.Lookup("clipboard.append")
	if err := appendH(ctx, Step{Name: "clipboard.append", Params: map[string]string{"text": "🙂"}}); err != nil {
		t.Fatalf("append err: %v", err)
	}
	prependH, _ := r.Lookup("clipboard.prepend")
	if err := prependH(ctx, Step{Name: "clipboard.prepend", Params: map[string]string{"text": "<"}}); err != nil {
		t.Fatalf("prepend err: %v", err)
	}
	if clip.text != "<base🙂" {
		t.Fatalf("clipboard=%q", clip.text)
	}

	readH, _ := r.Lookup("clipboard.read")
	if err := readH(ctx, Step{Name: "clipboard.read", Params: map[string]string{"save_as": "clip"}}); err != nil {
		t.Fatalf("read err: %v", err)
	}
	if got := ctx.Metadata["clip"]; got != "<base🙂" {
		t.Fatalf("metadata clip=%q", got)
	}
}

func TestClipboardWrite_WithRestore(t *testing.T) {
	r := NewRegistry()
	clip := &fakeClipboard{text: "original"}
	ctx := ActionContext{Context: context.Background(), Services: Services{Clipboard: clip}}

	h, _ := r.Lookup("clipboard.write")
	err := h(ctx, Step{Name: "clipboard.write", Params: map[string]string{"text": "temp", "with_restore": "true"}})
	if err != nil {
		t.Fatalf("clipboard.write err: %v", err)
	}
	if clip.text != "original" {
		t.Fatalf("clipboard should be restored, got %q", clip.text)
	}
}
