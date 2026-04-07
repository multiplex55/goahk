package clipboard

import "context"

type Event struct {
	Text string
}

type Watcher interface {
	Events() <-chan Event
	Start(context.Context) error
	Stop() error
}

type NoopWatcher struct{}

func (NoopWatcher) Events() <-chan Event {
	ch := make(chan Event)
	close(ch)
	return ch
}

func (NoopWatcher) Start(context.Context) error { return nil }
func (NoopWatcher) Stop() error                 { return nil }
