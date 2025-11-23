package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"pr-reviewer/internal/domain"
	repoMocks "pr-reviewer/mocks/repository"
)

func TestTeamService_AddTeam_Success(t *testing.T) {
	repo := repoMocks.NewMockTeamRepository(t)
	repo.On("GetTeamByName", mock.Anything, "backend").Return(domain.Team{}, domain.NewDomainError(domain.ErrorCodeNotFound, "not found"))
	repo.On("UpsertTeam", mock.Anything, mock.MatchedBy(func(team domain.Team) bool { return team.Name == "backend" })).Return(domain.Team{Name: "backend"}, nil)

	svc := NewTeamService(repo)

	team, err := svc.AddTeam(context.Background(), domain.Team{Name: "backend"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.Name != "backend" {
		t.Fatalf("unexpected team name: %s", team.Name)
	}
}

func TestTeamService_AddTeam_Exists(t *testing.T) {
	repo := repoMocks.NewMockTeamRepository(t)
	repo.On("GetTeamByName", mock.Anything, "backend").Return(domain.Team{Name: "backend"}, nil)
	svc := NewTeamService(repo)

	_, err := svc.AddTeam(context.Background(), domain.Team{Name: "backend"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeTeamExists {
		t.Fatalf("expected team exists error, got %v", err)
	}
}

func TestTeamService_AddTeam_GetError(t *testing.T) {
	repo := repoMocks.NewMockTeamRepository(t)
	repo.On("GetTeamByName", mock.Anything, "backend").Return(domain.Team{}, errors.New("db down"))
	svc := NewTeamService(repo)

	_, err := svc.AddTeam(context.Background(), domain.Team{Name: "backend"})
	if err == nil || err.Error() != "db down" {
		t.Fatalf("expected db down error, got %v", err)
	}
}

func TestTeamService_AddTeam_UpsertError(t *testing.T) {
	repo := repoMocks.NewMockTeamRepository(t)
	repo.On("GetTeamByName", mock.Anything, "backend").Return(domain.Team{}, domain.NewDomainError(domain.ErrorCodeNotFound, "not found"))
	repo.On("UpsertTeam", mock.Anything, mock.Anything).Return(domain.Team{}, errors.New("fail"))
	svc := NewTeamService(repo)

	_, err := svc.AddTeam(context.Background(), domain.Team{Name: "backend"})
	if err == nil || err.Error() != "fail" {
		t.Fatalf("expected fail error, got %v", err)
	}
}

func TestTeamService_GetTeam_Success(t *testing.T) {
	repo := repoMocks.NewMockTeamRepository(t)
	repo.On("GetTeamByName", mock.Anything, "backend").Return(domain.Team{Name: "backend"}, nil)
	svc := NewTeamService(repo)

	team, err := svc.GetTeam(context.Background(), "backend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.Name != "backend" {
		t.Fatalf("unexpected team name: %s", team.Name)
	}
}

func TestTeamService_GetTeam_NotFound(t *testing.T) {
	repo := repoMocks.NewMockTeamRepository(t)
	repo.On("GetTeamByName", mock.Anything, "unknown").Return(domain.Team{}, domain.NewDomainError(domain.ErrorCodeNotFound, "missing"))
	svc := NewTeamService(repo)

	_, err := svc.GetTeam(context.Background(), "unknown")
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeNotFound {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestTeamService_GetTeam_OtherError(t *testing.T) {
	repo := repoMocks.NewMockTeamRepository(t)
	repo.On("GetTeamByName", mock.Anything, "backend").Return(domain.Team{}, errors.New("boom"))
	svc := NewTeamService(repo)

	_, err := svc.GetTeam(context.Background(), "backend")
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}
