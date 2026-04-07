package actions

import (
	"context"
	"errors"
	"testing"

	"goahk/internal/process"
	"goahk/internal/services/messagebox"
)

type fakeMessageBoxService struct {
	lastReq messagebox.Request
	err     error
}

func (f *fakeMessageBoxService) Show(_ context.Context, req messagebox.Request) error {
	f.lastReq = req
	return f.err
}

type fakeProcessService struct {
	lastReq process.Request
	err     error
}

func (f *fakeProcessService) Launch(_ context.Context, req process.Request) error {
	f.lastReq = req
	return f.err
}

func TestMessageBoxAction_NormalizesParams(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("system.message_box")
	mb := &fakeMessageBoxService{}
	ctx := ActionContext{Context: context.Background(), Services: Services{MessageBox: mb}}

	err := h(ctx, Step{Name: "system.message_box", Params: map[string]string{"title": "T", "message": "Body"}})
	if err != nil {
		t.Fatalf("system.message_box err = %v", err)
	}
	if mb.lastReq.Title != "T" || mb.lastReq.Body != "Body" || mb.lastReq.Icon != "info" || mb.lastReq.Options != "ok" {
		t.Fatalf("request = %#v", mb.lastReq)
	}
}

func TestActionValidation_MissingRequiredFields(t *testing.T) {
	r := NewRegistry()
	ctx := ActionContext{Context: context.Background(), Services: Services{MessageBox: &fakeMessageBoxService{}, Clipboard: &fakeClipboard{}, Process: &fakeProcessService{}}}

	msg, _ := r.Lookup("system.message_box")
	if err := msg(ctx, Step{Name: "system.message_box", Params: map[string]string{"title": "x"}}); err == nil {
		t.Fatal("expected message_box validation error")
	}

	clip, _ := r.Lookup("clipboard.write")
	if err := clip(ctx, Step{Name: "clipboard.write", Params: map[string]string{}}); err == nil {
		t.Fatal("expected clipboard.write validation error")
	}

	proc, _ := r.Lookup("process.launch")
	if err := proc(ctx, Step{Name: "process.launch", Params: map[string]string{"args": "a b"}}); err == nil {
		t.Fatal("expected process.launch validation error")
	}
}

func TestServiceErrorPropagation_PreservedInExecutionResult(t *testing.T) {
	boom := errors.New("service failed")
	r := NewRegistry()
	exec := NewExecutor(r)
	res := exec.Execute(ActionContext{Context: context.Background(), Services: Services{Process: &fakeProcessService{err: boom}}}, Plan{{Name: "process.launch", Params: map[string]string{"executable": "x"}}})
	if res.Success {
		t.Fatal("execution should fail")
	}
	if len(res.Steps) != 1 {
		t.Fatalf("steps = %d", len(res.Steps))
	}
	if got := res.Steps[0].Error; got != "service failed" {
		t.Fatalf("step error = %q", got)
	}
}

func TestProcessLaunchAction_NormalizesParams(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("process.launch")
	proc := &fakeProcessService{}
	ctx := ActionContext{Context: context.Background(), Services: Services{Process: proc}}
	err := h(ctx, Step{Name: "process.launch", Params: map[string]string{
		"path":        "notepad.exe",
		"args":        "one two",
		"working_dir": `C:\\Temp`,
		"env":         "A=1;B=2",
	}})
	if err != nil {
		t.Fatalf("process.launch err = %v", err)
	}
	if proc.lastReq.Executable != "notepad.exe" || len(proc.lastReq.Args) != 2 || proc.lastReq.WorkingDir != `C:\\Temp` || proc.lastReq.Env["A"] != "1" {
		t.Fatalf("request = %#v", proc.lastReq)
	}
}
