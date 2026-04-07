package app

import (
	"context"
	"io"

	"goahk/internal/config"
)

type RuntimeDeps struct {
	Bootstrap       Bootstrap
	InitLogging     func(context.Context, config.LoggingConfig) error
	InitServices    func(context.Context, config.Config) (io.Closer, error)
	RegisterHotkeys func(context.Context, []config.HotkeyBinding) (io.Closer, error)
	RunMessageLoop  func(context.Context) error
}

type Runtime struct {
	deps RuntimeDeps
}

func NewRuntime(deps RuntimeDeps) *Runtime {
	if deps.Bootstrap.Load == nil {
		deps.Bootstrap = NewBootstrap()
	}
	return &Runtime{deps: deps}
}
