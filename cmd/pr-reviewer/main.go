package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"pr-reviewer/internal/app"
	"pr-reviewer/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := app.Run(ctx, cfg); err != nil {
		log.Fatalf("app exited with error: %v", err)
	}
}
