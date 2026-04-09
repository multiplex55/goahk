//go:build windows
// +build windows

package input

import (
	"context"
	"errors"
	"testing"
	"time"
)

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

	svc := windowsService{}
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

	orig := platformSendKeys
	t.Cleanup(func() { platformSendKeys = orig })

	backendErr := errors.New("backend exploded")
	platformSendKeys = func(context.Context, string, bool) error {
		return backendErr
	}

	err := windowsService{}.SendText(context.Background(), "hello", SendOptions{})
	if !errors.Is(err, ErrSendKeysFailed) {
		t.Fatalf("expected ErrSendKeysFailed, got %v", err)
	}
	if errors.Is(err, backendErr) {
		t.Fatalf("error should be mapped and not expose backend error identity: %v", err)
	}
}

func TestWindowsServiceContract_ContextErrorsPassThrough(t *testing.T) {
	t.Parallel()

	orig := platformSendKeys
	t.Cleanup(func() { platformSendKeys = orig })

	platformSendKeys = func(context.Context, string, bool) error {
		return context.Canceled
	}

	err := windowsService{}.SendKeys(context.Background(), Sequence{Tokens: []Token{{Keys: []string{"ctrl", "c"}}}}, SendOptions{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
