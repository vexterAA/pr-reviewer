package service

import (
	"context"

	"pr-reviewer/internal/domain"
)

type UserRepository interface {
	// TODO: define repo methods.
}

type UserService interface {
	_()
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) _() {}
