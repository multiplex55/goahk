package goahk

import (
	"sync"

	"goahk/internal/program"
)

type App struct {
	bindings        []bindingSpec
	logger          Logger
	validateActions bool
	state           *appState
}

type bindingSpec struct {
	hotkey string
	steps  []stepSpecProvider
}

func NewApp(opts ...Option) *App {
	a := &App{validateActions: true, state: newAppState()}
	for _, opt := range opts {
		if opt != nil {
			opt(a)
		}
	}
	return a
}

func (a *App) toProgram() program.Program {
	out := program.Program{Bindings: make([]program.BindingSpec, 0, len(a.bindings))}
	for i, b := range a.bindings {
		steps := make([]program.StepSpec, 0, len(b.steps))
		for _, step := range b.steps {
			steps = append(steps, step.stepSpec())
		}
		out.Bindings = append(out.Bindings, program.BindingSpec{ID: bindingID(i), Hotkey: b.hotkey, Steps: steps})
	}
	return program.Normalize(out)
}

type StateStore interface {
	Get(key string) (string, bool)
	Set(key, value string)
	LoadOrStore(key, value string) (string, bool)
}

type appState struct {
	mu   sync.RWMutex
	data map[string]string
}

func newAppState() *appState {
	return &appState{data: map[string]string{}}
}

func (s *appState) Get(key string) (string, bool) {
	if s == nil {
		return "", false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.data[key]
	return value, ok
}

func (s *appState) Set(key, value string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *appState) LoadOrStore(key, value string) (string, bool) {
	if s == nil {
		return value, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.data[key]; ok {
		return existing, true
	}
	s.data[key] = value
	return value, false
}
