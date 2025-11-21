package app

import "context"

type Config interface{}

func Run(ctx context.Context, cfg Config) error {
	// TODO: wire up logger, db, repositories, services, http server and run it here.
	<-ctx.Done()
	return nil
}
