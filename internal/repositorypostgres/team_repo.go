package repositorypostgres

import (
	"pr-reviewer/internal/repository"
)

type teamRepo struct {
	db *DB
}

func NewTeamRepository(db *DB) repository.TeamRepository {
	return &teamRepo{db: db}
}
