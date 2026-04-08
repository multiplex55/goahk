package actions

import (
	"context"
	"testing"

	"goahk/internal/process"
	"goahk/internal/services/messagebox"
)

type fakeContractMessageBox struct{ called bool }

func (f *fakeContractMessageBox) Show(context.Context, messagebox.Request) error {
	f.called = true
	return nil
}

type fakeContractClipboard struct{}

func (fakeContractClipboard) ReadText(context.Context) (string, error) { return "", nil }
func (fakeContractClipboard) WriteText(context.Context, string) error  { return nil }
func (fakeContractClipboard) AppendText(context.Context, string) error { return nil }
func (fakeContractClipboard) PrependText(context.Context, string) error {
	return nil
}

type fakeContractProcess struct{ called bool }

func (f *fakeContractProcess) Launch(context.Context, process.Request) error {
	f.called = true
	return nil
}

func TestServiceContractsAreActionLocalInterfaces(t *testing.T) {
	mb := &fakeContractMessageBox{}
	procSvc := &fakeContractProcess{}
	ctx := ActionContext{Context: context.Background(), Services: Services{MessageBox: mb, Clipboard: fakeContractClipboard{}, Process: procSvc}}

	if err := runMessageBoxAction(ctx, Step{Name: "system.message_box", Params: map[string]string{"body": "hello"}}); err != nil {
		t.Fatalf("message box err = %v", err)
	}
	if err := runProcessLaunchAction(ctx, Step{Name: "process.launch", Params: map[string]string{"executable": "notepad.exe"}}); err != nil {
		t.Fatalf("process launch err = %v", err)
	}
	if !mb.called {
		t.Fatalf("expected messagebox service to be called")
	}
	if !procSvc.called {
		t.Fatalf("expected process service to be called")
	}
}
