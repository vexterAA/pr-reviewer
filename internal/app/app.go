package app

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"pr-reviewer/internal/config"
	httpapi "pr-reviewer/internal/http"
	"pr-reviewer/internal/logging"
	"pr-reviewer/internal/repositorypostgres"
	"pr-reviewer/internal/service"
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

	teamRepo := repositorypostgres.NewTeamRepository(db)
	userRepo := repositorypostgres.NewUserRepository(db)
	prRepo := repositorypostgres.NewPullRequestRepository(db)
	uow := repositorypostgres.NewUnitOfWork(db)

	teamService := service.NewTeamService(teamRepo)
	userService := service.NewUserService(userRepo, prRepo)
	prService := service.NewPullRequestService(prRepo, userRepo, uow)

	router := httpapi.NewRouter(teamService, userService, prService)

	addr := cfg.HTTPPort
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logger.Info("Shutting down HTTP server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown error: %v", err)
			return err
		}
		logger.Info("HTTP server stopped gracefully")
		return nil
	case err := <-errCh:
		logger.Error("HTTP server failed: %v", err)
		return err
	}
}
