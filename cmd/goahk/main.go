package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"goahk/internal/runtime"
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

	bootstrap := runtime.NewBootstrap()
	if err := bootstrap.Run(ctx, *configPath); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
