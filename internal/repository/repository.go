package repository

import (
	"context"
	"time"

	"pr-reviewer/internal/domain"
)

type TeamRepository interface {
	UpsertTeam(ctx context.Context, team domain.Team) (domain.Team, error)
	GetTeamByName(ctx context.Context, teamName string) (domain.Team, error)
}

type UserRepository interface {
	GetUserByID(ctx context.Context, userID string) (domain.User, error)
	SetActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	ListActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error)
}

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error)
	GetPullRequestByID(ctx context.Context, prID string) (domain.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string, mergedAt time.Time) (domain.PullRequest, error)
	UpdateStatus(ctx context.Context, prID string, status domain.PullRequestStatus) (domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (domain.PullRequest, error)
	ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error)
}

type Tx interface {
	TeamRepository
	UserRepository
	PullRequestRepository
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type UnitOfWork interface {
	Begin(ctx context.Context) (Tx, error)
}
