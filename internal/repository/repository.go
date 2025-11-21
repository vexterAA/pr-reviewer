package repository

import (
	"context"

	"pr-reviewer/internal/domain"
)

type TeamRepository interface {
	// TODO
}

type UserRepository interface {
	// TODO
}

type PullRequestRepository interface {
	// TODO
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
