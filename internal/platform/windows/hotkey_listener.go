//go:build windows
// +build windows

package windows

import (
	"context"

	"goahk/internal/hotkey"
)

// HotkeyListener adapts the hotkey Win32 listener into runtime-friendly shape.
type HotkeyListener struct {
	inner *hotkey.Win32Listener
}

func NewHotkeyListener(context.Context) (*HotkeyListener, error) {
	inner, err := hotkey.NewWin32Listener()
	if err != nil {
		return nil, err
	}
	return &HotkeyListener{inner: inner}, nil
}

func (l *HotkeyListener) Register(registrationID int, chord hotkey.Chord) error {
	return l.inner.Register(registrationID, chord)
}

func (l *HotkeyListener) Unregister(registrationID int) error {
	return l.inner.Unregister(registrationID)
}

func (l *HotkeyListener) Events() <-chan hotkey.ListenerEvent { return l.inner.Events() }

func (l *HotkeyListener) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (l *HotkeyListener) Close() error {
	return l.inner.Close()
}
