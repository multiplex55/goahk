package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"goahk/internal/app"
	"goahk/internal/config"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()
	log.Printf("starting goahk version=%s commit=%s built=%s", version, commit, buildDate)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	runtime := app.NewRuntime(app.RuntimeDeps{
		InitLogging: func(_ context.Context, logging config.LoggingConfig) error {
			log.Printf("logging initialized (level=%s format=%s)", logging.Level, logging.Format)
			return nil
		},
		InitServices: func(_ context.Context, _ config.Config) (io.Closer, error) {
			return noopCloser{}, nil
		},
		RegisterHotkeys: func(_ context.Context, hotkeys []config.HotkeyBinding) (io.Closer, error) {
			log.Printf("registered %d hotkeys", len(hotkeys))
			return noopCloser{}, nil
		},
		RunMessageLoop: func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		},
	})

	if err := runtime.Run(ctx, *configPath); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

type noopCloser struct{}

func (noopCloser) Close() error { return nil }
