package repositorypostgres

import (
	"pr-reviewer/internal/repository"
)

type prRepo struct {
	db *DB
}

func NewPullRequestRepository(db *DB) repository.PullRequestRepository {
	return &prRepo{db: db}
}
