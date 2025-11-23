package repositorypostgres

import (
	"context"
	"database/sql"
	"time"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type unitOfWork struct {
	db *DB
}

func NewUnitOfWork(db *DB) repository.UnitOfWork {
	return &unitOfWork{db: db}
}

func (u *unitOfWork) Begin(ctx context.Context) (repository.Tx, error) {
	tx, err := u.db.SQL.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return newTx(tx), nil
}

type tx struct {
	tx    *sql.Tx
	teams *teamRepo
	users *userRepo
	prs   *prRepo
}

func newTx(t *sql.Tx) *tx {
	return &tx{
		tx:    t,
		teams: &teamRepo{exec: t},
		users: &userRepo{exec: t},
		prs:   &prRepo{exec: t},
	}
}

func (t *tx) Commit(ctx context.Context) error   { return t.tx.Commit() }
func (t *tx) Rollback(ctx context.Context) error { return t.tx.Rollback() }

// TeamRepository
func (t *tx) UpsertTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	return t.teams.UpsertTeam(ctx, team)
}

func (t *tx) GetTeamByName(ctx context.Context, teamName string) (domain.Team, error) {
	return t.teams.GetTeamByName(ctx, teamName)
}

// UserRepository
func (t *tx) GetUserByID(ctx context.Context, userID string) (domain.User, error) {
	return t.users.GetUserByID(ctx, userID)
}

func (t *tx) SetActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	return t.users.SetActive(ctx, userID, isActive)
}

func (t *tx) ListActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	return t.users.ListActiveByTeam(ctx, teamName)
}

// PullRequestRepository
func (t *tx) CreatePullRequest(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	return t.prs.CreatePullRequest(ctx, pr)
}

func (t *tx) GetPullRequestByID(ctx context.Context, prID string) (domain.PullRequest, error) {
	return t.prs.GetPullRequestByID(ctx, prID)
}

func (t *tx) MergePullRequest(ctx context.Context, prID string, mergedAt time.Time) (domain.PullRequest, error) {
	return t.prs.MergePullRequest(ctx, prID, mergedAt)
}

func (t *tx) UpdateStatus(ctx context.Context, prID string, status domain.PullRequestStatus) (domain.PullRequest, error) {
	return t.prs.UpdateStatus(ctx, prID, status)
}

func (t *tx) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (domain.PullRequest, error) {
	return t.prs.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID)
}

func (t *tx) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	return t.prs.ListByReviewer(ctx, reviewerID)
}
