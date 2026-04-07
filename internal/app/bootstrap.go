package app

import (
	"context"
	"fmt"

	"goahk/internal/config"
)

type ConfigLoader func(string) (config.Config, error)

type Bootstrap struct {
	Load ConfigLoader
}

func NewBootstrap() Bootstrap {
	return Bootstrap{Load: config.LoadFile}
}

func (b Bootstrap) LoadConfig(_ context.Context, path string) (config.Config, error) {
	if b.Load == nil {
		return config.Config{}, fmt.Errorf("config loader is not configured")
	}
	return b.Load(path)
}
