package app

import (
	"context"

	"pr-reviewer/internal/config"
)

func Run(ctx context.Context, cfg *config.Config) error {
	// TODO: wire up logger, db, repositories, services, http server and run it here.
	<-ctx.Done()
	return nil
}
