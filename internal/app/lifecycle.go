package app

import (
	"context"
	"fmt"
	"io"
)

func (r *Runtime) Run(ctx context.Context, configPath string) (err error) {
	cfg, err := r.deps.Bootstrap.LoadConfig(ctx, configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if r.deps.InitLogging != nil {
		if err := r.deps.InitLogging(ctx, cfg.Logging); err != nil {
			return fmt.Errorf("initialize logging: %w", err)
		}
	}

	closers := make([]io.Closer, 0, 2)
	defer func() {
		closeErr := closeAll(closers)
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if r.deps.InitServices != nil {
		serviceCloser, serviceErr := r.deps.InitServices(ctx, cfg)
		if serviceErr != nil {
			return fmt.Errorf("initialize services: %w", serviceErr)
		}
		if serviceCloser != nil {
			closers = append(closers, serviceCloser)
		}
	}

	if r.deps.RegisterHotkeys != nil {
		hotkeyCloser, hotkeyErr := r.deps.RegisterHotkeys(ctx, cfg.Hotkeys)
		if hotkeyErr != nil {
			return fmt.Errorf("register hotkeys: %w", hotkeyErr)
		}
		if hotkeyCloser != nil {
			closers = append(closers, hotkeyCloser)
		}
	}

	if r.deps.RunMessageLoop != nil {
		if err := r.deps.RunMessageLoop(ctx); err != nil {
			return fmt.Errorf("run message loop: %w", err)
		}
	}

	return nil
}

func closeAll(closers []io.Closer) error {
	for i := len(closers) - 1; i >= 0; i-- {
		if err := closers[i].Close(); err != nil {
			return err
		}
	}
	return nil
}
