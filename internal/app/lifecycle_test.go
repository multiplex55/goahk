package app

import (
	"context"
	"io"
	"reflect"
	"testing"

	"goahk/internal/config"
)

type closerFunc func() error

func (f closerFunc) Close() error { return f() }

func TestRuntimeRun_StartupAndCleanupOrder(t *testing.T) {
	var order []string
	cfg := config.Config{Hotkeys: []config.HotkeyBinding{{ID: "x", Hotkey: "Ctrl+1", Steps: []config.Step{{Action: "noop"}}}}}

	r := NewRuntime(RuntimeDeps{
		Bootstrap: Bootstrap{Load: func(string) (config.Config, error) {
			order = append(order, "load config")
			return cfg, nil
		}},
		InitLogging: func(context.Context, config.LoggingConfig) error {
			order = append(order, "init logging")
			return nil
		},
		InitServices: func(context.Context, config.Config) (io.Closer, error) {
			order = append(order, "init services")
			return closerFunc(func() error {
				order = append(order, "close services")
				return nil
			}), nil
		},
		RegisterHotkeys: func(context.Context, []config.HotkeyBinding) (io.Closer, error) {
			order = append(order, "register hotkeys")
			return closerFunc(func() error {
				order = append(order, "close hotkeys")
				return nil
			}), nil
		},
		RunMessageLoop: func(context.Context) error {
			order = append(order, "run loop")
			return nil
		},
	})

	if err := r.Run(context.Background(), "ignored"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := []string{
		"load config",
		"init logging",
		"init services",
		"register hotkeys",
		"run loop",
		"close hotkeys",
		"close services",
	}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %#v, want %#v", order, want)
	}
}

func TestRuntimeRun_CleanupOnRegistrationFailure(t *testing.T) {
	var order []string
	cfg := config.Config{}

	r := NewRuntime(RuntimeDeps{
		Bootstrap: Bootstrap{Load: func(string) (config.Config, error) { return cfg, nil }},
		InitServices: func(context.Context, config.Config) (io.Closer, error) {
			order = append(order, "init services")
			return closerFunc(func() error {
				order = append(order, "close services")
				return nil
			}), nil
		},
		RegisterHotkeys: func(context.Context, []config.HotkeyBinding) (io.Closer, error) {
			order = append(order, "register hotkeys")
			return nil, context.Canceled
		},
	})

	if err := r.Run(context.Background(), "ignored"); err == nil {
		t.Fatal("Run() error = nil, want failure")
	}

	want := []string{"init services", "register hotkeys", "close services"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %#v, want %#v", order, want)
	}
}
