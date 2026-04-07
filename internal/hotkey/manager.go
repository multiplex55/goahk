package hotkey

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Listener interface {
	Register(registrationID int, chord Chord) error
	Unregister(registrationID int) error
	Events() <-chan ListenerEvent
	Close() error
}

type Manager struct {
	listener Listener
	nextID   int

	mu           sync.RWMutex
	byBindingID  map[string]int
	byRegID      map[int]registration
	triggerEvent chan TriggerEvent

	runWG       sync.WaitGroup
	closeOnce   sync.Once
	closeDone   chan struct{}
	closeErr    error
	triggerOnce sync.Once
	shutdown    chan struct{}
}

type registration struct {
	bindingID string
	chord     Chord
}

func NewManager(listener Listener) *Manager {
	return &Manager{
		listener:     listener,
		nextID:       1,
		byBindingID:  map[string]int{},
		byRegID:      map[int]registration{},
		triggerEvent: make(chan TriggerEvent, 32),
		closeDone:    make(chan struct{}),
		shutdown:     make(chan struct{}),
	}
}

func (m *Manager) Register(bindingID string, chord Chord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.byBindingID[bindingID]; exists {
		return fmt.Errorf("binding %q already registered", bindingID)
	}
	regID := m.nextID
	m.nextID++
	if err := m.listener.Register(regID, chord); err != nil {
		return err
	}
	m.byBindingID[bindingID] = regID
	m.byRegID[regID] = registration{bindingID: bindingID, chord: chord}
	return nil
}

func (m *Manager) Unregister(bindingID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	regID, ok := m.byBindingID[bindingID]
	if !ok {
		return fmt.Errorf("binding %q is not registered", bindingID)
	}
	if err := m.listener.Unregister(regID); err != nil {
		return err
	}
	delete(m.byBindingID, bindingID)
	delete(m.byRegID, regID)
	return nil
}

func (m *Manager) Events() <-chan TriggerEvent {
	return m.triggerEvent
}

func (m *Manager) Run(ctx context.Context) error {
	m.runWG.Add(1)
	defer m.runWG.Done()
	defer m.triggerOnce.Do(func() {
		close(m.triggerEvent)
	})

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.shutdown:
			return nil
		case ev, ok := <-m.listener.Events():
			if !ok {
				return nil
			}
			m.mu.RLock()
			reg, exists := m.byRegID[ev.RegistrationID]
			m.mu.RUnlock()
			if !exists {
				continue
			}
			at := ev.TriggeredAt
			if at.IsZero() {
				at = time.Now().UTC()
			}
			select {
			case m.triggerEvent <- TriggerEvent{BindingID: reg.bindingID, Chord: reg.chord, TriggeredAt: at}:
			case <-ctx.Done():
				return ctx.Err()
			case <-m.shutdown:
				return nil
			}
		}
	}
}

func (m *Manager) Close() error {
	m.closeOnce.Do(func() {
		close(m.shutdown)
		m.closeErr = m.listener.Close()
		m.runWG.Wait()
		close(m.closeDone)
	})
	<-m.closeDone
	return m.closeErr
}
