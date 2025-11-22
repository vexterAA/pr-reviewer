package repositorypostgres

import (
	"context"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type prRepo struct {
	db *DB
}

func NewPullRequestRepository(db *DB) repository.PullRequestRepository {
	return &prRepo{db: db}
}

func (r *prRepo) Create(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	return domain.PullRequest{}, errNotImplemented
}

func (r *prRepo) Merge(ctx context.Context, prID string) (domain.PullRequest, error) {
	return domain.PullRequest{}, errNotImplemented
}

func (r *prRepo) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (domain.PullRequest, string, error) {
	return domain.PullRequest{}, "", errNotImplemented
}

func (r *prRepo) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	return nil, errNotImplemented
}
