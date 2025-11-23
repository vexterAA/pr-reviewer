package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"pr-reviewer/internal/domain"
	repoMocks "pr-reviewer/mocks/repository"
)

type metricsStub struct {
	created   int
	merged    int
	reassigns map[string]int
}

func (m *metricsStub) IncPRCreated() { m.created++ }
func (m *metricsStub) IncPRMerged()  { m.merged++ }
func (m *metricsStub) IncPRReassign(result string) {
	if m.reassigns == nil {
		m.reassigns = make(map[string]int)
	}
	m.reassigns[result]++
}

func TestPullRequestService_Create_PRExists(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1"}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}

	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)
	_, err := svc.Create(context.Background(), domain.PullRequest{ID: "pr1"})
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodePRExists {
		t.Fatalf("expected PR_EXISTS error, got %v", err)
	}
}

func TestPullRequestService_Create_AuthorNotFound(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, "missing"))
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{}, domain.NewDomainError(domain.ErrorCodeNotFound, "no author"))
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, err := svc.Create(context.Background(), domain.PullRequest{ID: "pr1", AuthorID: "u1"})
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeNotFound || derr.Message != "author not found" {
		t.Fatalf("expected author not found domain error, got %v", err)
	}
}

func TestPullRequestService_Create_ReviewerSelection(t *testing.T) {
	tests := []struct {
		name              string
		activeUsers       []domain.User
		expectedReviewers int
	}{
		{"none", []domain.User{}, 0},
		{"one", []domain.User{{ID: "u1", TeamName: "t"}, {ID: "u2", TeamName: "t"}}, 1},
		{"twoOrMore", []domain.User{{ID: "u2"}, {ID: "u1"}, {ID: "u3"}}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prRepo := repoMocks.NewMockPullRequestRepository(t)
			prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, "not found"))

			userRepo := repoMocks.NewMockUserRepository(t)
			userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{ID: "u1", TeamName: "t"}, nil)
			userRepo.On("ListActiveByTeam", mock.Anything, "t").Return(tt.activeUsers, nil)

			tx := repoMocks.NewMockTx(t)
			tx.On("CreatePullRequest", mock.Anything, mock.AnythingOfType("domain.PullRequest")).Return(func(_ context.Context, pr domain.PullRequest) domain.PullRequest {
				return pr
			}, nil)
			tx.On("Commit", mock.Anything).Return(nil)
			tx.On("Rollback", mock.Anything).Return(nil)

			uow := repoMocks.NewMockUnitOfWork(t)
			uow.On("Begin", mock.Anything).Return(tx, nil)

			metrics := &metricsStub{}
			svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

			pr, err := svc.Create(context.Background(), domain.PullRequest{ID: "pr1", AuthorID: "u1"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(pr.AssignedReviewers) != tt.expectedReviewers {
				t.Fatalf("expected %d reviewers, got %d", tt.expectedReviewers, len(pr.AssignedReviewers))
			}
			for _, rev := range pr.AssignedReviewers {
				if rev == "u1" {
					t.Fatalf("author should not be assigned as reviewer")
				}
			}
			if metrics.created != 1 {
				t.Fatalf("expected metrics created increment")
			}
		})
	}
}

func TestPullRequestService_Create_BeginError(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, "not found"))
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{ID: "u1", TeamName: "t"}, nil)
	userRepo.On("ListActiveByTeam", mock.Anything, "t").Return([]domain.User{}, nil)

	uow := repoMocks.NewMockUnitOfWork(t)
	uow.On("Begin", mock.Anything).Return(nil, errors.New("begin fail"))
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, err := svc.Create(context.Background(), domain.PullRequest{ID: "pr1", AuthorID: "u1"})
	if err == nil || err.Error() != "begin fail" {
		t.Fatalf("expected begin fail, got %v", err)
	}
}

func TestPullRequestService_Create_CreateError(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, "not found"))
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{ID: "u1", TeamName: "t"}, nil)
	userRepo.On("ListActiveByTeam", mock.Anything, "t").Return([]domain.User{}, nil)

	tx := repoMocks.NewMockTx(t)
	tx.On("CreatePullRequest", mock.Anything, mock.AnythingOfType("domain.PullRequest")).Return(domain.PullRequest{}, errors.New("create fail"))
	tx.On("Rollback", mock.Anything).Return(nil)
	uow := repoMocks.NewMockUnitOfWork(t)
	uow.On("Begin", mock.Anything).Return(tx, nil)

	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, err := svc.Create(context.Background(), domain.PullRequest{ID: "pr1", AuthorID: "u1"})
	if err == nil || err.Error() != "create fail" {
		t.Fatalf("expected create fail, got %v", err)
	}
}

func TestPullRequestService_Create_CommitError(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, "not found"))
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u1").Return(domain.User{ID: "u1", TeamName: "t"}, nil)
	userRepo.On("ListActiveByTeam", mock.Anything, "t").Return([]domain.User{}, nil)

	tx := repoMocks.NewMockTx(t)
	tx.On("CreatePullRequest", mock.Anything, mock.AnythingOfType("domain.PullRequest")).Return(func(_ context.Context, pr domain.PullRequest) domain.PullRequest {
		return pr
	}, nil)
	tx.On("Commit", mock.Anything).Return(errors.New("commit fail"))
	tx.On("Rollback", mock.Anything).Return(nil)
	uow := repoMocks.NewMockUnitOfWork(t)
	uow.On("Begin", mock.Anything).Return(tx, nil)

	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, err := svc.Create(context.Background(), domain.PullRequest{ID: "pr1", AuthorID: "u1"})
	if err == nil || err.Error() != "commit fail" {
		t.Fatalf("expected commit fail, got %v", err)
	}
}

func TestPullRequestService_Merge_NotFound(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, "missing"))
	userRepo := repoMocks.NewMockUserRepository(t)
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, err := svc.Merge(context.Background(), "pr1")
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeNotFound {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestPullRequestService_Merge_AlreadyMerged(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusMerged}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	pr, err := svc.Merge(context.Background(), "pr1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.Status != domain.PullRequestStatusMerged {
		t.Fatalf("expected merged status")
	}
	prRepo.AssertNotCalled(t, "MergePullRequest", mock.Anything, mock.Anything, mock.Anything)
}

func TestPullRequestService_Merge_Success(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusOpen}, nil)
	prRepo.On("MergePullRequest", mock.Anything, "pr1", mock.Anything).Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusMerged}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	pr, err := svc.Merge(context.Background(), "pr1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.Status != domain.PullRequestStatusMerged {
		t.Fatalf("expected merged status")
	}
	prRepo.AssertCalled(t, "MergePullRequest", mock.Anything, "pr1", mock.Anything)
	if metrics.merged != 1 {
		t.Fatalf("expected merged metric increment")
	}
}

func TestPullRequestService_Merge_Error(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusOpen}, nil)
	prRepo.On("MergePullRequest", mock.Anything, "pr1", mock.Anything).Return(domain.PullRequest{}, errors.New("merge fail"))
	userRepo := repoMocks.NewMockUserRepository(t)
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, err := svc.Merge(context.Background(), "pr1")
	if err == nil || err.Error() != "merge fail" {
		t.Fatalf("expected merge fail, got %v", err)
	}
}

func TestPullRequestService_Reassign_NotFound(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{}, domain.NewDomainError(domain.ErrorCodeNotFound, "missing"))
	userRepo := repoMocks.NewMockUserRepository(t)
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, _, err := svc.Reassign(context.Background(), "pr1", "u2")
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestPullRequestService_Reassign_Merged(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusMerged}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, _, err := svc.Reassign(context.Background(), "pr1", "u2")
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodePRMerged {
		t.Fatalf("expected pr merged, got %v", err)
	}
}

func TestPullRequestService_Reassign_NotAssigned(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusOpen, AssignedReviewers: []string{"u3"}}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, _, err := svc.Reassign(context.Background(), "pr1", "u2")
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeNotAssigned {
		t.Fatalf("expected not assigned, got %v", err)
	}
}

func TestPullRequestService_Reassign_ReviewerNotFound(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusOpen, AssignedReviewers: []string{"u2"}}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u2").Return(domain.User{}, domain.NewDomainError(domain.ErrorCodeNotFound, "no reviewer"))
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, _, err := svc.Reassign(context.Background(), "pr1", "u2")
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeNotFound || derr.Message != "reviewer not found" {
		t.Fatalf("expected reviewer not found, got %v", err)
	}
}

func TestPullRequestService_Reassign_NoCandidate(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusOpen, AuthorID: "u1", AssignedReviewers: []string{"u2"}}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u2").Return(domain.User{ID: "u2", TeamName: "t"}, nil)
	userRepo.On("ListActiveByTeam", mock.Anything, "t").Return([]domain.User{{ID: "u1", TeamName: "t"}, {ID: "u2", TeamName: "t"}}, nil)
	uow := repoMocks.NewMockUnitOfWork(t)
	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, _, err := svc.Reassign(context.Background(), "pr1", "u2")
	if derr, ok := domain.AsDomainError(err); !ok || derr.Code != domain.ErrorCodeNoCandidate {
		t.Fatalf("expected no candidate, got %v", err)
	}
}

func TestPullRequestService_Reassign_Success(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusOpen, AuthorID: "u1", AssignedReviewers: []string{"u2"}}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u2").Return(domain.User{ID: "u2", TeamName: "t"}, nil)
	userRepo.On("ListActiveByTeam", mock.Anything, "t").Return([]domain.User{{ID: "u1", TeamName: "t"}, {ID: "u2", TeamName: "t"}, {ID: "u3", TeamName: "t"}}, nil)

	tx := repoMocks.NewMockTx(t)
	tx.On("ReassignReviewer", mock.Anything, "pr1", "u2", mock.Anything).Return(domain.PullRequest{ID: "pr1", AssignedReviewers: []string{"u3"}}, nil)
	tx.On("Commit", mock.Anything).Return(nil)
	tx.On("Rollback", mock.Anything).Return(nil)

	uow := repoMocks.NewMockUnitOfWork(t)
	uow.On("Begin", mock.Anything).Return(tx, nil)

	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	pr, newReviewer, err := svc.Reassign(context.Background(), "pr1", "u2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newReviewer == "" || newReviewer == "u2" {
		t.Fatalf("expected new reviewer other than old")
	}
	if pr.ID != "pr1" {
		t.Fatalf("unexpected pr returned")
	}
}

func TestPullRequestService_Reassign_BeginError(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusOpen, AssignedReviewers: []string{"u2"}}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u2").Return(domain.User{ID: "u2", TeamName: "t"}, nil)
	userRepo.On("ListActiveByTeam", mock.Anything, "t").Return([]domain.User{{ID: "u3", TeamName: "t"}}, nil)

	uow := repoMocks.NewMockUnitOfWork(t)
	uow.On("Begin", mock.Anything).Return(nil, errors.New("begin fail"))

	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, _, err := svc.Reassign(context.Background(), "pr1", "u2")
	if err == nil || err.Error() != "begin fail" {
		t.Fatalf("expected begin fail, got %v", err)
	}
}

func TestPullRequestService_Reassign_ReassignError(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusOpen, AuthorID: "u1", AssignedReviewers: []string{"u2"}}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u2").Return(domain.User{ID: "u2", TeamName: "t"}, nil)
	userRepo.On("ListActiveByTeam", mock.Anything, "t").Return([]domain.User{{ID: "u3", TeamName: "t"}}, nil)

	tx := repoMocks.NewMockTx(t)
	tx.On("ReassignReviewer", mock.Anything, "pr1", "u2", mock.Anything).Return(domain.PullRequest{}, errors.New("reassign fail"))
	tx.On("Rollback", mock.Anything).Return(nil)

	uow := repoMocks.NewMockUnitOfWork(t)
	uow.On("Begin", mock.Anything).Return(tx, nil)

	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, _, err := svc.Reassign(context.Background(), "pr1", "u2")
	if err == nil || err.Error() != "reassign fail" {
		t.Fatalf("expected reassign fail, got %v", err)
	}
}

func TestPullRequestService_Reassign_CommitError(t *testing.T) {
	prRepo := repoMocks.NewMockPullRequestRepository(t)
	prRepo.On("GetPullRequestByID", mock.Anything, "pr1").Return(domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusOpen, AuthorID: "u1", AssignedReviewers: []string{"u2"}}, nil)
	userRepo := repoMocks.NewMockUserRepository(t)
	userRepo.On("GetUserByID", mock.Anything, "u2").Return(domain.User{ID: "u2", TeamName: "t"}, nil)
	userRepo.On("ListActiveByTeam", mock.Anything, "t").Return([]domain.User{{ID: "u3", TeamName: "t"}}, nil)

	tx := repoMocks.NewMockTx(t)
	tx.On("ReassignReviewer", mock.Anything, "pr1", "u2", mock.Anything).Return(domain.PullRequest{ID: "pr1"}, nil)
	tx.On("Commit", mock.Anything).Return(errors.New("commit fail"))
	tx.On("Rollback", mock.Anything).Return(nil)

	uow := repoMocks.NewMockUnitOfWork(t)
	uow.On("Begin", mock.Anything).Return(tx, nil)

	metrics := &metricsStub{}
	svc := NewPullRequestService(prRepo, userRepo, uow, metrics)

	_, _, err := svc.Reassign(context.Background(), "pr1", "u2")
	if err == nil || err.Error() != "commit fail" {
		t.Fatalf("expected commit fail, got %v", err)
	}
}

func TestPickReviewers(t *testing.T) {
	users := []domain.User{
		{ID: "author"},
		{ID: "u1"},
		{ID: "u2"},
		{ID: "u3"},
	}

	revs := pickReviewers(users, "author", 2)
	if len(revs) != 2 {
		t.Fatalf("expected 2 reviewers, got %d", len(revs))
	}
	for _, r := range revs {
		if r == "author" {
			t.Fatalf("author should not be reviewer")
		}
	}
}
