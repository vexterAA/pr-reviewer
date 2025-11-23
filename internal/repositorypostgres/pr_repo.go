package repositorypostgres

import (
	"context"
	"database/sql"
	"time"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/repository"
)

type prRepo struct {
	exec executor
}

func NewPullRequestRepository(db *DB) repository.PullRequestRepository {
	return &prRepo{exec: db.SQL}
}

func (r *prRepo) CreatePullRequest(ctx context.Context, pr domain.PullRequest) (domain.PullRequest, error) {
	row := r.exec.QueryRowContext(ctx, `
		INSERT INTO pull_requests (id, name, author_id, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, COALESCE($5, NOW()), $6)
		RETURNING id, name, author_id, status, created_at, merged_at
	`, pr.ID, pr.Name, pr.AuthorID, pr.Status, timeOrNil(pr.CreatedAt), pr.MergedAt)

	created, err := scanPullRequest(row)
	if err != nil {
		return domain.PullRequest{}, err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		if _, err := r.exec.ExecContext(ctx, `
			INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, created.ID, reviewerID); err != nil {
			return domain.PullRequest{}, err
		}
	}

	created.AssignedReviewers = pr.AssignedReviewers
	return created, nil
}

func (r *prRepo) GetPullRequestByID(ctx context.Context, prID string) (domain.PullRequest, error) {
	row := r.exec.QueryRowContext(ctx, `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`, prID)

	pr, err := scanPullRequest(row)
	if err != nil {
		return domain.PullRequest{}, wrapNotFound(err)
	}

	reviewers, err := r.listReviewers(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	pr.AssignedReviewers = reviewers
	return pr, nil
}

func (r *prRepo) MergePullRequest(ctx context.Context, prID string, mergedAt time.Time) (domain.PullRequest, error) {
	row := r.exec.QueryRowContext(ctx, `
		UPDATE pull_requests
		SET status = $2,
		    merged_at = $3
		WHERE id = $1
		RETURNING id, name, author_id, status, created_at, merged_at
	`, prID, domain.PullRequestStatusMerged, mergedAt)

	pr, err := scanPullRequest(row)
	if err != nil {
		return domain.PullRequest{}, wrapNotFound(err)
	}

	reviewers, err := r.listReviewers(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	pr.AssignedReviewers = reviewers
	return pr, nil
}

func (r *prRepo) UpdateStatus(ctx context.Context, prID string, status domain.PullRequestStatus) (domain.PullRequest, error) {
	row := r.exec.QueryRowContext(ctx, `
		UPDATE pull_requests
		SET status = $2
		WHERE id = $1
		RETURNING id, name, author_id, status, created_at, merged_at
	`, prID, status)

	pr, err := scanPullRequest(row)
	if err != nil {
		return domain.PullRequest{}, wrapNotFound(err)
	}
	reviewers, err := r.listReviewers(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	pr.AssignedReviewers = reviewers
	return pr, nil
}

func (r *prRepo) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) (domain.PullRequest, error) {
	if _, err := r.exec.ExecContext(ctx, `
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1 AND reviewer_id = $2
	`, prID, oldReviewerID); err != nil {
		return domain.PullRequest{}, err
	}

	if _, err := r.exec.ExecContext(ctx, `
		INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, prID, newReviewerID); err != nil {
		return domain.PullRequest{}, err
	}

	return r.GetPullRequestByID(ctx, prID)
}

func (r *prRepo) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	rows, err := r.exec.QueryContext(ctx, `
		SELECT pr.id, pr.name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pull_request_reviewers r ON r.pull_request_id = pr.id
		WHERE r.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`, reviewerID)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	var result []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		result = append(result, pr)
	}
	return result, nil
}

func (r *prRepo) listReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := r.exec.QueryContext(ctx, `
		SELECT reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
		ORDER BY reviewer_id
	`, prID)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)

	var reviewers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, id)
	}
	return reviewers, nil
}

func scanPullRequest(row *sql.Row) (domain.PullRequest, error) {
	var pr domain.PullRequest
	if err := row.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
		return domain.PullRequest{}, err
	}
	return pr, nil
}

func timeOrNil(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
