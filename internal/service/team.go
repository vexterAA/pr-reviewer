package service

import (
	"context"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type TeamService interface {
	AddTeam(ctx context.Context, team domain.Team) (*domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}

type teamService struct {
	repo repository.TeamRepository
}

func NewTeamService(repo repository.TeamRepository) TeamService {
	return &teamService{repo: repo}
}

func (s *teamService) AddTeam(ctx context.Context, team domain.Team) (*domain.Team, error) {
	created, err := s.repo.UpsertTeam(ctx, team)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *teamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	team, err := s.repo.GetTeamByName(ctx, teamName)
	if err != nil {
		return nil, err
	}
	return &team, nil
}
