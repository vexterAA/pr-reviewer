package service

import (
	"context"
	"time"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/metrics"
	"pr-reviewer/internal/repository"
)

type PullRequestService interface {
	Create(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error)
	Merge(ctx context.Context, prID string) (*domain.PullRequest, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error)
}

type pullRequestService struct {
	prs     repository.PullRequestRepository
	users   repository.UserRepository
	uow     repository.UnitOfWork
	metrics metrics.BusinessMetrics
}

func NewPullRequestService(prs repository.PullRequestRepository, users repository.UserRepository, uow repository.UnitOfWork, metrics metrics.BusinessMetrics) PullRequestService {
	return &pullRequestService{
		prs:     prs,
		users:   users,
		uow:     uow,
		metrics: metrics,
	}
}

func (s *pullRequestService) Create(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error) {
	if _, err := s.prs.GetPullRequestByID(ctx, pr.ID); err == nil {
		return nil, domain.NewDomainError(domain.ErrorCodePRExists, "pull request already exists")
	} else if derr, ok := domain.AsDomainError(err); ok && derr.Code == domain.ErrorCodeNotFound {
		// ok
	} else if err != nil {
		return nil, err
	}

	author, err := s.users.GetUserByID(ctx, pr.AuthorID)
	if err != nil {
		if derr, ok := domain.AsDomainError(err); ok && derr.Code == domain.ErrorCodeNotFound {
			return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "author not found")
		}
		return nil, err
	}

	active, err := s.users.ListActiveByTeam(ctx, author.TeamName)
	if err != nil {
		return nil, err
	}

	reviewers := pickReviewers(active, pr.AuthorID, 2)

	pr.Status = domain.PullRequestStatusOpen
	pr.AssignedReviewers = reviewers
	pr.CreatedAt = time.Now().UTC()

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	created, err := tx.CreatePullRequest(ctx, pr)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	if s.metrics != nil {
		s.metrics.IncPRCreated()
	}

	return &created, nil
}

func (s *pullRequestService) Merge(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prs.GetPullRequestByID(ctx, prID)
	if err != nil {
		if derr, ok := domain.AsDomainError(err); ok && derr.Code == domain.ErrorCodeNotFound {
			return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "pull request not found")
		}
		return nil, err
	}

	if pr.Status == domain.PullRequestStatusMerged {
		return &pr, nil
	}

	merged, err := s.prs.MergePullRequest(ctx, prID, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if s.metrics != nil {
		s.metrics.IncPRMerged()
	}
	return &merged, nil
}

func (s *pullRequestService) Reassign(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pr, err := s.prs.GetPullRequestByID(ctx, prID)
	if err != nil {
		if derr, ok := domain.AsDomainError(err); ok && derr.Code == domain.ErrorCodeNotFound {
			return nil, "", s.reassignMetricErr("not_found", domain.NewDomainError(domain.ErrorCodeNotFound, "pull request not found"))
		}
		return nil, "", s.reassignMetricErr("internal_error", err)
	}

	if pr.Status == domain.PullRequestStatusMerged {
		return nil, "", s.reassignMetricErr("pr_merged", domain.NewDomainError(domain.ErrorCodePRMerged, "cannot reassign on merged pull request"))
	}

	if !contains(pr.AssignedReviewers, oldReviewerID) {
		return nil, "", s.reassignMetricErr("not_assigned", domain.NewDomainError(domain.ErrorCodeNotAssigned, "reviewer is not assigned to this pull request"))
	}

	oldReviewer, err := s.users.GetUserByID(ctx, oldReviewerID)
	if err != nil {
		if derr, ok := domain.AsDomainError(err); ok && derr.Code == domain.ErrorCodeNotFound {
			return nil, "", s.reassignMetricErr("not_found", domain.NewDomainError(domain.ErrorCodeNotFound, "reviewer not found"))
		}
		return nil, "", s.reassignMetricErr("internal_error", err)
	}

	candidate, err := s.pickReplacementCandidate(ctx, oldReviewer.TeamName, pr.AuthorID, pr.AssignedReviewers)
	if err != nil {
		code := "internal_error"
		if derr, ok := domain.AsDomainError(err); ok && derr.Code == domain.ErrorCodeNoCandidate {
			code = "no_candidate"
		}
		return nil, "", s.reassignMetricErr(code, err)
	}

	tx, err := s.uow.Begin(ctx)
	if err != nil {
		return nil, "", s.reassignMetricErr("internal_error", err)
	}
	defer tx.Rollback(ctx)

	updated, err := tx.ReassignReviewer(ctx, prID, oldReviewerID, candidate)
	if err != nil {
		return nil, "", s.reassignMetricErr("internal_error", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", s.reassignMetricErr("internal_error", err)
	}

	if s.metrics != nil {
		s.metrics.IncPRReassign("success")
	}

	return &updated, candidate, nil
}

func (s *pullRequestService) reassignMetricErr(result string, err error) error {
	if s.metrics != nil {
		s.metrics.IncPRReassign(result)
	}
	return err
}

func pickReviewers(users []domain.User, authorID string, limit int) []string {
	var reviewers []string
	for _, u := range users {
		if u.ID == authorID {
			continue
		}
		reviewers = append(reviewers, u.ID)
		if len(reviewers) == limit {
			break
		}
	}
	return reviewers
}

func (s *pullRequestService) pickReplacementCandidate(ctx context.Context, teamName, authorID string, currentReviewers []string) (string, error) {
	users, err := s.users.ListActiveByTeam(ctx, teamName)
	if err != nil {
		return "", err
	}

	current := make(map[string]struct{}, len(currentReviewers))
	for _, id := range currentReviewers {
		current[id] = struct{}{}
	}

	for _, u := range users {
		if u.ID == authorID {
			continue
		}
		if _, exists := current[u.ID]; exists {
			continue
		}
		return u.ID, nil
	}

	return "", domain.NewDomainError(domain.ErrorCodeNoCandidate, "no active replacement candidate in team")
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
