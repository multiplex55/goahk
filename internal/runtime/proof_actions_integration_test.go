//go:build windows && integration

package runtime

import (
	"context"
	"testing"
	"time"

	"goahk/internal/actions"
	"goahk/internal/clipboard"
	"goahk/internal/hotkey"
	"goahk/internal/process"
	"goahk/internal/services/messagebox"
)

type recordingMessageBox struct {
	called bool
}

func (r *recordingMessageBox) Show(context.Context, messagebox.Request) error {
	r.called = true
	return nil
}

func TestIntegrationProof_HotkeyToMessageBoxPathReachable(t *testing.T) {
	reg := actions.NewRegistry()
	exec := actions.NewExecutor(reg)
	events := make(chan hotkey.TriggerEvent, 1)
	shutdown := make(chan struct{})
	mb := &recordingMessageBox{}

	results := DispatchHotkeyEvents(context.Background(), shutdown, events, map[string]actions.Plan{
		"hk.msg": {{Name: "system.message_box", Params: map[string]string{"body": "proof"}}},
	}, exec, actions.ActionContext{Services: actions.Services{MessageBox: mb}}, nil)

	events <- hotkey.TriggerEvent{BindingID: "hk.msg", Chord: hotkey.Chord{Modifiers: hotkey.ModCtrl, Key: "M"}}
	select {
	case res := <-results:
		if !res.Execution.Success {
			t.Fatalf("execution failed: %#v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for dispatch result")
	}
	close(shutdown)
	for range results {
	}
	if !mb.called {
		t.Fatal("message box service was not called")
	}
}

func TestIntegrationProof_DirectAdapterInvocationWritesClipboardText(t *testing.T) {
	reg := actions.NewRegistry()
	h, _ := reg.Lookup("clipboard.write")
	clipSvc := clipboard.NewService(nil)
	ctx := actions.ActionContext{Context: context.Background(), Services: actions.Services{Clipboard: clipSvc}}

	if err := h(ctx, actions.Step{Name: "clipboard.write", Params: map[string]string{"text": "integration-proof-🙂"}}); err != nil {
		t.Fatalf("clipboard.write err = %v", err)
	}
	got, err := clipSvc.ReadText(context.Background())
	if err != nil {
		t.Fatalf("ReadText err = %v", err)
	}
	if got != "integration-proof-🙂" {
		t.Fatalf("clipboard text = %q", got)
	}
}

func TestIntegrationProof_ProcessLaunchStartsBenignTarget(t *testing.T) {
	reg := actions.NewRegistry()
	h, _ := reg.Lookup("process.launch")
	ctx := actions.ActionContext{Context: context.Background(), Services: actions.Services{Process: process.NewService()}}

	err := h(ctx, actions.Step{Name: "process.launch", Params: map[string]string{
		"executable": "cmd.exe",
		"args":       "/C exit 0",
	}})
	if err != nil {
		t.Fatalf("process.launch err = %v", err)
	}
}
