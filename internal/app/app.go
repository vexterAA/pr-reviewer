package app

import (
	"context"

	"pr-reviewer/internal/config"
	"pr-reviewer/internal/logging"
	"pr-reviewer/internal/repositorypostgres"
)

func Run(ctx context.Context, cfg *config.Config) error {
	logger := logging.StdLogger{}

	db, err := repositorypostgres.NewDB(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := repositorypostgres.ApplyMigrations(ctx, db, "migrations", logger); err != nil {
		return err
	}

	// TODO: wire up logger, repositories, services, http server and run it here.
	<-ctx.Done()
	return nil
}
