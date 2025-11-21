package service

import (
	"context"

	"pr-reviewer/internal/domain"
)

type PullRequestRepository interface {
	// TODO: define repo methods.
}

type PullRequestService interface {
	_()
}

type pullRequestService struct {
	repo PullRequestRepository
}

func NewPullRequestService(repo PullRequestRepository) PullRequestService {
	return &pullRequestService{repo: repo}
}

func (s *pullRequestService) _() {}
