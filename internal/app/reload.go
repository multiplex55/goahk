package app

import (
	"context"
	"fmt"
	"io"
	"sync"

	"goahk/internal/config"
)

type ReloadNotifier interface {
	Errorf(format string, args ...any)
}

type ReloadManager struct {
	LoadConfig func(path string) (config.Config, error)
	Activate   func(context.Context, config.Config) (io.Closer, error)
	Notify     ReloadNotifier

	mu     sync.RWMutex
	active io.Closer
	cfg    config.Config
}

func (m *ReloadManager) CurrentConfig() config.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

func (m *ReloadManager) Reload(ctx context.Context, path string) error {
	if m.LoadConfig == nil || m.Activate == nil {
		return fmt.Errorf("reload manager is not configured")
	}
	cfg, err := m.LoadConfig(path)
	if err != nil {
		m.notify("config reload failed validation: %v", err)
		return fmt.Errorf("validate config: %w", err)
	}
	newActive, err := m.Activate(ctx, cfg)
	if err != nil {
		m.notify("config reload failed; previous config still active: %v", err)
		return fmt.Errorf("activate config: %w", err)
	}

	m.mu.Lock()
	oldActive := m.active
	m.active = newActive
	m.cfg = cfg
	m.mu.Unlock()

	if oldActive != nil {
		if err := oldActive.Close(); err != nil {
			m.notify("config reload applied but failed closing previous runtime: %v", err)
			return fmt.Errorf("close previous runtime: %w", err)
		}
	}
	return nil
}

func (m *ReloadManager) notify(format string, args ...any) {
	if m.Notify != nil {
		m.Notify.Errorf(format, args...)
	}
}
