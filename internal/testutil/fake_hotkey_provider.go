package testutil

import (
	"context"
	"io"

	"goahk/internal/config"
	"goahk/internal/hotkey"
)

type FakeHotkeyProvider struct {
	RegisterCalls int
	LastBindings  []config.HotkeyBinding
	RegisterErr   error
	Closer        io.Closer
	EventsCh      chan hotkey.TriggerEvent
}

func (f *FakeHotkeyProvider) RegisterHotkeys(_ context.Context, bindings []config.HotkeyBinding) (io.Closer, error) {
	f.RegisterCalls++
	f.LastBindings = append([]config.HotkeyBinding(nil), bindings...)
	if f.RegisterErr != nil {
		return nil, f.RegisterErr
	}
	if f.Closer == nil {
		f.Closer = noopCloser{}
	}
	return f.Closer, nil
}

func (f *FakeHotkeyProvider) Events() <-chan hotkey.TriggerEvent {
	if f.EventsCh == nil {
		f.EventsCh = make(chan hotkey.TriggerEvent)
	}
	return f.EventsCh
}

type noopCloser struct{}

func (noopCloser) Close() error { return nil }
