package input

import (
	"context"
	"time"
)

// Service provides synthetic input operations for actions.
//
// It is intentionally independent from the hotkey package internals so
// triggering and synthetic input can evolve separately.
type Service interface {
	SendText(ctx context.Context, text string, opts SendOptions) error
	SendKeys(ctx context.Context, seq Sequence, opts SendOptions) error
	SendChord(ctx context.Context, chord Chord, opts SendOptions) error
	MouseService
}

type SendOptions struct {
	DelayBefore        time.Duration
	SuppressReentrancy bool
}

type Chord struct {
	Keys []string
}
