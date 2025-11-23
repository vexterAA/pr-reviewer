package service

import (
	"context"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type UserService interface {
	SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetReviewPullRequests(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

type userService struct {
	users        repository.UserRepository
	pullRequests repository.PullRequestRepository
}

func NewUserService(users repository.UserRepository, pullRequests repository.PullRequestRepository) UserService {
	return &userService{
		users:        users,
		pullRequests: pullRequests,
	}
}

func (s *userService) SetActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	if _, err := s.users.GetUserByID(ctx, userID); err != nil {
		if derr, ok := domain.AsDomainError(err); ok && derr.Code == domain.ErrorCodeNotFound {
			return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found")
		}
		return nil, err
	}

	updated, err := s.users.SetActive(ctx, userID, isActive)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (s *userService) GetReviewPullRequests(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if _, err := s.users.GetUserByID(ctx, userID); err != nil {
		if derr, ok := domain.AsDomainError(err); ok && derr.Code == domain.ErrorCodeNotFound {
			return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found")
		}
		return nil, err
	}
	return s.pullRequests.ListByReviewer(ctx, userID)
}
