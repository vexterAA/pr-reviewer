package repositorypostgres

import (
	"pr-reviewer/internal/repository"
)

type userRepo struct {
	db *DB
}

func NewUserRepository(db *DB) repository.UserRepository {
	return &userRepo{db: db}
}
