//go:build windows
// +build windows

package input

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeWindowsBackend struct {
	sendTextErr  error
	sendChordErr error
}

func (f fakeWindowsBackend) sendText(context.Context, string) error       { return f.sendTextErr }
func (f fakeWindowsBackend) sendChord(context.Context, []string) error    { return f.sendChordErr }
func (f fakeWindowsBackend) moveAbsolute(context.Context, int, int) error { return nil }
func (f fakeWindowsBackend) moveRelative(context.Context, int, int) error { return nil }
func (f fakeWindowsBackend) position(context.Context) (MousePosition, error) {
	return MousePosition{}, nil
}
func (f fakeWindowsBackend) mouseButton(context.Context, string, string) error { return nil }
func (f fakeWindowsBackend) wheel(context.Context, int) error                  { return nil }

func TestWindowsServiceContract_ArgumentValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(Service) error
	}{
		{
			name: "negative delay rejected",
			call: func(s Service) error {
				return s.SendText(context.Background(), "ok", SendOptions{DelayBefore: -time.Millisecond})
			},
		},
		{
			name: "empty key in sequence rejected",
			call: func(s Service) error {
				return s.SendKeys(context.Background(), Sequence{Tokens: []Token{{Keys: []string{"ctrl", ""}}}}, SendOptions{})
			},
		},
		{
			name: "empty chord rejected",
			call: func(s Service) error {
				return s.SendChord(context.Background(), Chord{}, SendOptions{})
			},
		},
	}

	svc := windowsService{backend: fakeWindowsBackend{}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(svc)
			if !errors.Is(err, ErrInvalidInputArgument) {
				t.Fatalf("expected ErrInvalidInputArgument, got %v", err)
			}
		})
	}
}

func TestWindowsServiceContract_SendErrorMapping(t *testing.T) {
	t.Parallel()

	backendErr := errors.New("backend exploded")
	svc := windowsService{backend: fakeWindowsBackend{sendTextErr: backendErr}}
	err := svc.SendText(context.Background(), "hello", SendOptions{})
	if !errors.Is(err, ErrSendKeysFailed) {
		t.Fatalf("expected ErrSendKeysFailed, got %v", err)
	}
	if errors.Is(err, backendErr) {
		t.Fatalf("error should be mapped and not expose backend error identity: %v", err)
	}
}

func TestWindowsServiceContract_ContextErrorsPassThrough(t *testing.T) {
	t.Parallel()

	svc := windowsService{backend: fakeWindowsBackend{sendChordErr: context.Canceled}}
	err := svc.SendKeys(context.Background(), Sequence{Tokens: []Token{{Keys: []string{"ctrl", "c"}}}}, SendOptions{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
