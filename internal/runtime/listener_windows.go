//go:build windows

package runtime

import (
	"context"

	"goahk/internal/hotkey"
)

type windowsListener struct {
	inner *hotkey.Win32Listener
}

func NewWindowsListener(context.Context) (Listener, error) {
	inner, err := hotkey.NewWin32Listener()
	if err != nil {
		return nil, err
	}
	return &windowsListener{inner: inner}, nil
}

func (l *windowsListener) Register(registrationID int, chord hotkey.Chord) error {
	return l.inner.Register(registrationID, chord)
}

func (l *windowsListener) Unregister(registrationID int) error {
	return l.inner.Unregister(registrationID)
}

func (l *windowsListener) Events() <-chan hotkey.ListenerEvent { return l.inner.Events() }

func (l *windowsListener) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (l *windowsListener) Close() error {
	return l.inner.Close()
}
