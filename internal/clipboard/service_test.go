package clipboard

import (
	"context"
	"errors"
	"testing"
)

type fakeBackend struct {
	text        string
	readErr     error
	writeErr    error
	writeInputs []string
}

func (f *fakeBackend) ReadText(context.Context) (string, error) {
	if f.readErr != nil {
		return "", f.readErr
	}
	return f.text, nil
}

func (f *fakeBackend) WriteText(_ context.Context, text string) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.text = text
	f.writeInputs = append(f.writeInputs, text)
	return nil
}

func TestService_ReadWriteContract(t *testing.T) {
	backend := &fakeBackend{text: "hello\x00\x00"}
	svc := NewService(backend)

	got, err := svc.ReadText(context.Background())
	if err != nil || got != "hello" {
		t.Fatalf("ReadText() = %q, %v", got, err)
	}

	if err := svc.WriteText(context.Background(), "world\x00"); err != nil {
		t.Fatalf("WriteText() err = %v", err)
	}
	if backend.text != "world" {
		t.Fatalf("backend.text = %q, want world", backend.text)
	}
}

func TestService_AppendPrepend(t *testing.T) {
	backend := &fakeBackend{text: "β"}
	svc := NewService(backend)

	if err := svc.AppendText(context.Background(), "🙂"); err != nil {
		t.Fatalf("AppendText() err = %v", err)
	}
	if backend.text != "β🙂" {
		t.Fatalf("after append %q", backend.text)
	}
	if err := svc.PrependText(context.Background(), "α"); err != nil {
		t.Fatalf("PrependText() err = %v", err)
	}
	if backend.text != "αβ🙂" {
		t.Fatalf("after prepend %q", backend.text)
	}
}

func TestService_PropagatesBackendErrors(t *testing.T) {
	boom := errors.New("boom")
	svc := NewService(&fakeBackend{readErr: boom})
	if _, err := svc.ReadText(context.Background()); !errors.Is(err, boom) {
		t.Fatalf("ReadText() err = %v", err)
	}
	svc = NewService(&fakeBackend{writeErr: boom})
	if err := svc.WriteText(context.Background(), "x"); !errors.Is(err, boom) {
		t.Fatalf("WriteText() err = %v", err)
	}
}
