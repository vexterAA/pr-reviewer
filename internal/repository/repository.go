package repository

import (
	"context"

	"pr-reviewer/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, team domain.Team) (domain.Team, error)
	Get(ctx context.Context, teamName string) (domain.Team, error)
}

type UserRepository interface {
	Get(ctx context.Context, userID string) (domain.User, error)
	SetActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
}

type PullRequestRepository interface {
	Create(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error)
	Merge(ctx context.Context, prID string) (domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (domain.PullRequest, string, error)
	ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error)
}

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
