package repositorypostgres

import (
	"context"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type teamRepo struct {
	db *DB
}

func NewTeamRepository(db *DB) repository.TeamRepository {
	return &teamRepo{db: db}
}

func (r *teamRepo) Create(ctx context.Context, team domain.Team) (domain.Team, error) {
	return domain.Team{}, errNotImplemented
}

func (r *teamRepo) Get(ctx context.Context, teamName string) (domain.Team, error) {
	return domain.Team{}, errNotImplemented
}
