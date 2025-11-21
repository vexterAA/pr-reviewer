package service

import (
	"context"

	"pr-reviewer/internal/domain"
)

type TeamRepository interface {
	// TODO: define repo methods.
}

type TeamService interface {
	_()
}

type teamService struct {
	repo TeamRepository
}

func NewTeamService(repo TeamRepository) TeamService {
	return &teamService{repo: repo}
}

func (s *teamService) _() {}
