package main

import (
	"context"
	"testing"
)

func TestStatusBarToggleAccPathCapture(t *testing.T) {
	c := NewController(context.Background(), &fakeInspectService{})
	upd := c.OnStatusInteractionUpdate()
	if !upd.CaptureEnabled || upd.Text != "Click on path to copy to Clipboard" {
		t.Fatalf("unexpected update after enable: %+v", upd)
	}
	upd = c.OnStatusInteractionUpdate()
	if upd.CaptureEnabled || upd.Text != "Click here to enable Acc path capturing (can't be used with UIA!)" {
		t.Fatalf("unexpected update after disable: %+v", upd)
	}
}

func TestStatusBarShowsPathWhenAvailable(t *testing.T) {
	c := NewController(context.Background(), &fakeInspectService{})
	c.ToggleAccPathCapture()
	c.lastACCPath = "Desktop/Pane/Button"
	upd := c.OnStatusInteractionUpdate()
	if upd.Text != "Path: Desktop/Pane/Button" || !upd.HasLastACCPath {
		t.Fatalf("unexpected update: %+v", upd)
	}
}

func TestStatusBarClickCopiesPathWhenEnabled(t *testing.T) {
	cb := &fakeClipboard{}
	c := NewController(context.Background(), &fakeInspectService{}).WithClipboard(cb)
	c.ToggleAccPathCapture()
	c.lastACCPath = "Desktop/Pane/Button"
	upd := c.OnStatusInteractionUpdate()
	if !upd.LastACCPathCopied || len(cb.copied) != 1 || cb.copied[0] != "Desktop/Pane/Button" {
		t.Fatalf("unexpected copy behavior: update=%+v copied=%v", upd, cb.copied)
	}
}
