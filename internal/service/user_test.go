package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"pr-reviewer/internal/domain"
	repoMocks "pr-reviewer/mocks/repository"
)

func TestUserService_SetActive_Success(t *testing.T) {
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{ID: "u1", TeamName: "backend"}, nil)
	userRepo.On("SetActive", mock.Anything, "u1", true).Return(domain.User{ID: "u1", IsActive: true}, nil)

	prRepo := repoMocks.NewMockPullRequestRepository(t)
	svc := NewUserService(userRepo, prRepo)

	user, err := svc.SetActive(context.Background(), "u1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !user.IsActive {
		t.Fatalf("expected user active")
	}
}

func TestUserService_SetActive_NotFound(t *testing.T) {
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{}, domain.NewDomainError(domain.ErrorCodeNotFound, "missing"))
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	svc := NewUserService(userRepo, prRepo)

	_, err := svc.SetActive(context.Background(), "u1", true)
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeNotFound || derr.Message != "user not found" {
		t.Fatalf("expected not found domain error, got %v", err)
	}
}

func TestUserService_SetActive_GetError(t *testing.T) {
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{}, errors.New("db"))
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	svc := NewUserService(userRepo, prRepo)

	_, err := svc.SetActive(context.Background(), "u1", true)
	if err == nil || err.Error() != "db" {
		t.Fatalf("expected db error, got %v", err)
	}
}

func TestUserService_SetActive_SetError(t *testing.T) {
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{ID: "u1"}, nil)
	userRepo.On("SetActive", mock.Anything, "u1", true).Return(domain.User{}, errors.New("update fail"))
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	svc := NewUserService(userRepo, prRepo)

	_, err := svc.SetActive(context.Background(), "u1", true)
	if err == nil || err.Error() != "update fail" {
		t.Fatalf("expected update fail error, got %v", err)
	}
}

func TestUserService_GetReview_Success(t *testing.T) {
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{ID: "u1"}, nil)
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("ListByReviewer", mock.Anything, "u1").Return([]domain.PullRequestShort{{ID: "pr1"}, {ID: "pr2"}}, nil)

	svc := NewUserService(userRepo, prRepo)

	prs, err := svc.GetReviewPullRequests(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prs) != 2 {
		t.Fatalf("expected 2 prs, got %d", len(prs))
	}
}

func TestUserService_GetReview_NotFound(t *testing.T) {
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{}, domain.NewDomainError(domain.ErrorCodeNotFound, "nope"))
	prRepo := repoMocks.NewMockPullRequestRepository(t)

	svc := NewUserService(userRepo, prRepo)

	_, err := svc.GetReviewPullRequests(context.Background(), "u1")
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeNotFound || derr.Message != "user not found" {
		t.Fatalf("expected not found domain error, got %v", err)
	}
}

func TestUserService_GetReview_ListError(t *testing.T) {
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{ID: "u1"}, nil)
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("ListByReviewer", mock.Anything, "u1").Return(nil, errors.New("list fail"))

	svc := NewUserService(userRepo, prRepo)

	_, err := svc.GetReviewPullRequests(context.Background(), "u1")
	if err == nil || err.Error() != "list fail" {
		t.Fatalf("expected list fail error, got %v", err)
	}
}
