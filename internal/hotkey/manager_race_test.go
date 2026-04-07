package hotkey

import (
	"context"
	"sync"
	"testing"
	"time"
)

type raceListener struct {
	events chan ListenerEvent
}

func (r *raceListener) Register(int, Chord) error    { return nil }
func (r *raceListener) Unregister(int) error         { return nil }
func (r *raceListener) Events() <-chan ListenerEvent { return r.events }
func (r *raceListener) Close() error                 { close(r.events); return nil }

func TestManager_RunCancellationRace(t *testing.T) {
	l := &raceListener{events: make(chan ListenerEvent, 512)}
	m := NewManager(l)
	if err := m.Register("hk", Chord{Modifiers: ModCtrl, Key: "K"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = m.Run(ctx)
	}()

	for i := 0; i < 100; i++ {
		l.events <- ListenerEvent{RegistrationID: 1, TriggeredAt: time.Now().UTC()}
	}
	cancel()
	wg.Wait()
}
