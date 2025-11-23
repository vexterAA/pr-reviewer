package service

import (
	"context"
	"time"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type PullRequestService interface {
	Create(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error)
	Merge(ctx context.Context, prID string) (*domain.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) (*domain.PullRequest, error)
}

type pullRequestService struct {
	repo repository.PullRequestRepository
}

func NewPullRequestService(repo repository.PullRequestRepository) PullRequestService {
	return &pullRequestService{repo: repo}
}

func (s *pullRequestService) Create(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error) {
	created, err := s.repo.CreatePullRequest(ctx, pr)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *pullRequestService) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	merged, err := s.repo.MergePullRequest(ctx, prID, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return &merged, nil
}

func (s *pullRequestService) Reassign(ctx context.Context, prID, oldReviewerID, newReviewerID string) (*domain.PullRequest, error) {
	updated, err := s.repo.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID)
	if err != nil {
		return nil, err
	}
	return &updated, nil
}
