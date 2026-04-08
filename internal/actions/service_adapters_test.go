package actions

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestServiceAdapters_MissingServiceErrors(t *testing.T) {
	r := NewRegistry()
	cases := []struct {
		step        Step
		ctx         ActionContext
		serviceText string
	}{
		{
			step:        Step{Name: "system.message_box", Params: map[string]string{"body": "hello"}},
			ctx:         ActionContext{Context: context.Background()},
			serviceText: "message box service unavailable",
		},
		{
			step:        Step{Name: "process.launch", Params: map[string]string{"executable": "notepad.exe"}},
			ctx:         ActionContext{Context: context.Background()},
			serviceText: "process service unavailable",
		},
		{
			step:        Step{Name: "system.open", Params: map[string]string{"target": "https://chatgpt.com"}},
			ctx:         ActionContext{Context: context.Background()},
			serviceText: "process service unavailable",
		},
	}
	for _, tc := range cases {
		h, _ := r.Lookup(tc.step.Name)
		err := h(tc.ctx, tc.step)
		if err == nil {
			t.Fatalf("expected error for %s", tc.step.Name)
		}
		if !strings.Contains(err.Error(), tc.serviceText) {
			t.Fatalf("%s err=%q missing %q", tc.step.Name, err.Error(), tc.serviceText)
		}
		if !strings.Contains(err.Error(), tc.step.Name) {
			t.Fatalf("%s err=%q missing action identity", tc.step.Name, err.Error())
		}
	}
}

func TestSystemOpenAction_URLNormalization(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("system.open")
	proc := &fakeProcessService{}
	ctx := ActionContext{Context: context.Background(), Services: Services{Process: proc}}

	err := h(ctx, Step{Name: "system.open", Params: map[string]string{"target": "www.chatgpt.com", "kind": "url"}})
	if err != nil {
		t.Fatalf("system.open err = %v", err)
	}
	if proc.lastReq.OpenKind != process.OpenKindURL || proc.lastReq.OpenTarget != "https://www.chatgpt.com" {
		t.Fatalf("request = %#v", proc.lastReq)
	}
}

func TestSystemOpenAction_InvalidURLRejected(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("system.open")
	ctx := ActionContext{Context: context.Background(), Services: Services{Process: &fakeProcessService{}}}

	err := h(ctx, Step{Name: "system.open", Params: map[string]string{"target": "http://bad url", "kind": "url"}})
	if err == nil || !strings.Contains(err.Error(), "invalid url") {
		t.Fatalf("expected invalid url error, got %v", err)
	}
}

func TestSystemOpenAction_FolderValidation(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("system.open")
	proc := &fakeProcessService{}
	ctx := ActionContext{Context: context.Background(), Services: Services{Process: proc}}

	root := t.TempDir()
	err := h(ctx, Step{Name: "system.open", Params: map[string]string{"target": root, "kind": "folder"}})
	if err != nil {
		t.Fatalf("expected existing folder to pass: %v", err)
	}
	if proc.lastReq.OpenKind != process.OpenKindFolder || proc.lastReq.OpenTarget == "" {
		t.Fatalf("request = %#v", proc.lastReq)
	}

	missing := filepath.Join(root, "does-not-exist")
	err = h(ctx, Step{Name: "system.open", Params: map[string]string{"target": missing, "kind": "folder"}})
	if err == nil {
		t.Fatal("expected missing folder error")
	}
}

func TestSystemOpenAction_ApplicationValidation(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("system.open")
	proc := &fakeProcessService{}
	ctx := ActionContext{Context: context.Background(), Services: Services{Process: proc}}

	exe := filepath.Join(t.TempDir(), "app.exe")
	if err := h(ctx, Step{Name: "system.open", Params: map[string]string{"target": exe, "kind": "application"}}); err != nil {
		t.Fatalf("expected absolute exe to pass: %v", err)
	}
	if proc.lastReq.Executable != exe {
		t.Fatalf("request = %#v", proc.lastReq)
	}

	err := h(ctx, Step{Name: "system.open", Params: map[string]string{"target": "notepad.exe", "kind": "application"}})
	if err == nil || !strings.Contains(err.Error(), "absolute .exe path") {
		t.Fatalf("expected relative exe rejection, got %v", err)
	}
}

func TestSystemOpenAction_FolderAliasDownloads(t *testing.T) {
	r := NewRegistry()
	h, _ := r.Lookup("system.open")
	proc := &fakeProcessService{}
	ctx := ActionContext{Context: context.Background(), Services: Services{Process: proc}}

	home := t.TempDir()
	downloads := filepath.Join(home, "Downloads")
	if err := os.MkdirAll(downloads, 0o755); err != nil {
		t.Fatalf("mkdir downloads: %v", err)
	}
	t.Setenv("USERPROFILE", home)

	err := h(ctx, Step{Name: "system.open", Params: map[string]string{"target": "downloads", "kind": "folder"}})
	if err != nil {
		t.Fatalf("downloads alias should resolve: %v", err)
	}
	if proc.lastReq.OpenTarget != downloads {
		t.Fatalf("open target = %q, want %q", proc.lastReq.OpenTarget, downloads)
	}
}
