package runtime

import (
	"context"
	"testing"

	"goahk/internal/hotkey"
)

type fakeContractListener struct {
	events chan hotkey.ListenerEvent
}

func (f *fakeContractListener) Register(int, hotkey.Chord) error { return nil }
func (f *fakeContractListener) Unregister(int) error             { return nil }
func (f *fakeContractListener) Events() <-chan hotkey.ListenerEvent {
	return f.events
}
func (f *fakeContractListener) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
func (f *fakeContractListener) Close() error { return nil }

func TestListenerContractAcceptsNarrowInterface(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	b := Bootstrap{}
	if err := b.runLoop(ctx, &fakeContractListener{events: make(chan hotkey.ListenerEvent)}, make(chan error, 1)); err != nil {
		t.Fatalf("runLoop err = %v", err)
	}
}
