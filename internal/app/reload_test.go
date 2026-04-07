package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"goahk/internal/config"
)

type notifierSpy struct{ msgs []string }

func (n *notifierSpy) Errorf(format string, args ...any) {
	n.msgs = append(n.msgs, fmt.Sprintf(format, args...))
}

type closeSpy struct{ closed int }

func (c *closeSpy) Close() error { c.closed++; return nil }

func TestReloadManager_TransactionalRollbackOnInvalidConfig(t *testing.T) {
	old := &closeSpy{}
	notify := &notifierSpy{}
	manager := &ReloadManager{
		LoadConfig: func(string) (config.Config, error) { return config.Config{}, errors.New("schema violation") },
		Activate: func(context.Context, config.Config) (io.Closer, error) {
			t.Fatal("activate should not run on invalid config")
			return nil, nil
		},
		Notify: notify,
		active: old,
		cfg:    config.Config{Logging: config.LoggingConfig{Level: "info"}},
	}

	err := manager.Reload(context.Background(), "ignored")
	if err == nil || !strings.Contains(err.Error(), "validate config") {
		t.Fatalf("Reload() error = %v", err)
	}
	if old.closed != 0 {
		t.Fatalf("old runtime closed during rollback: %d", old.closed)
	}
	if got := manager.CurrentConfig().Logging.Level; got != "info" {
		t.Fatalf("current config level = %q, want info", got)
	}
	if len(notify.msgs) == 0 || !strings.Contains(notify.msgs[0], "failed validation") {
		t.Fatalf("expected user-visible validation error, got %#v", notify.msgs)
	}
}

func TestReloadManager_TransactionalSwap(t *testing.T) {
	old := &closeSpy{}
	notify := &notifierSpy{}
	calls := 0
	manager := &ReloadManager{
		LoadConfig: func(string) (config.Config, error) {
			return config.Config{Logging: config.LoggingConfig{Level: "debug"}}, nil
		},
		Activate: func(context.Context, config.Config) (io.Closer, error) {
			calls++
			return &closeSpy{}, nil
		},
		Notify: notify,
		active: old,
		cfg:    config.Config{Logging: config.LoggingConfig{Level: "info"}},
	}

	if err := manager.Reload(context.Background(), "ignored"); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("activate calls = %d", calls)
	}
	if old.closed != 1 {
		t.Fatalf("old runtime close count = %d, want 1", old.closed)
	}
	if got := manager.CurrentConfig().Logging.Level; got != "debug" {
		t.Fatalf("current config level = %q, want debug", got)
	}
	if len(notify.msgs) != 0 {
		t.Fatalf("unexpected notifications: %#v", notify.msgs)
	}
}
