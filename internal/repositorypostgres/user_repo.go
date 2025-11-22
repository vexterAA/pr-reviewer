package repositorypostgres

import (
	"context"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type userRepo struct {
	db *DB
}

func NewUserRepository(db *DB) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Get(ctx context.Context, userID string) (domain.User, error) {
	return domain.User{}, errNotImplemented
}

func (r *userRepo) SetActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	return domain.User{}, errNotImplemented
}
