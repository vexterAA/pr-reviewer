package service

import (
	"context"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type PullRequestService interface {
	Create(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error)
	Merge(ctx context.Context, prID string) (*domain.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error)
}

type pullRequestService struct {
	repo repository.PullRequestRepository
}

func NewPullRequestService(repo repository.PullRequestRepository) PullRequestService {
	return &pullRequestService{repo: repo}
}

func (s *pullRequestService) Create(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error) {
	created, err := s.repo.Create(ctx, pr)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *pullRequestService) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	merged, err := s.repo.Merge(ctx, prID)
	if err != nil {
		return nil, err
	}
	return &merged, nil
}

func (s *pullRequestService) Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	updated, replacedBy, err := s.repo.ReassignReviewer(ctx, prID, oldReviewerID)
	if err != nil {
		return nil, "", err
	}
	return &updated, replacedBy, nil
}
