package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"

	"pr-reviewer/internal/domain"
	serviceMocks "pr-reviewer/mocks/service"
)

func TestPRHandlers_Create_BadJSON(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPRHandlers_Create_MissingFields(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})
	body, _ := json.Marshal(map[string]any{"pull_request_id": "pr1"})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPRHandlers_Create_Success(t *testing.T) {
	prSvc := serviceMocks.NewMockPullRequestService(t)
	prSvc.On("Create", mock.Anything, mock.AnythingOfType("domain.PullRequest")).Return(&domain.PullRequest{
		ID:                "pr1",
		Name:              "Test",
		AuthorID:          "u1",
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
	}, nil)
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), prSvc, &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{
		"pull_request_id":   "pr1",
		"pull_request_name": "Test",
		"author_id":         "u1",
	})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	prSvc.AssertCalled(t, "Create", mock.Anything, mock.MatchedBy(func(pr domain.PullRequest) bool {
		return pr.ID == "pr1" && pr.AuthorID == "u1"
	}))
}

func TestPRHandlers_Create_PRExists(t *testing.T) {
	prSvc := serviceMocks.NewMockPullRequestService(t)
	prSvc.On("Create", mock.Anything, mock.AnythingOfType("domain.PullRequest")).Return(nil, domain.NewDomainError(domain.ErrorCodePRExists, "exists"))
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), prSvc, &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{
		"pull_request_id":   "pr1",
		"pull_request_name": "Test",
		"author_id":         "u1",
	})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestPRHandlers_Merge_BadRequest(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBufferString("{}"))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPRHandlers_Merge_Success(t *testing.T) {
	prSvc := serviceMocks.NewMockPullRequestService(t)
	prSvc.On("Merge", mock.Anything, "pr1").Return(&domain.PullRequest{ID: "pr1", Status: domain.PullRequestStatusMerged}, nil)
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), prSvc, &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{"pull_request_id": "pr1"})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	prSvc.AssertCalled(t, "Merge", mock.Anything, "pr1")
}

func TestPRHandlers_Merge_NotFound(t *testing.T) {
	prSvc := serviceMocks.NewMockPullRequestService(t)
	prSvc.On("Merge", mock.Anything, "pr1").Return(nil, domain.NewDomainError(domain.ErrorCodeNotFound, "not found"))
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), prSvc, &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{"pull_request_id": "pr1"})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPRHandlers_Reassign_BadRequest(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBufferString("{}"))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestPRHandlers_Reassign_Success(t *testing.T) {
	prSvc := serviceMocks.NewMockPullRequestService(t)
	prSvc.On("Reassign", mock.Anything, "pr1", "u2").Return(&domain.PullRequest{ID: "pr1"}, "u3", nil)
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), prSvc, &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{"pull_request_id": "pr1", "old_user_id": "u2"})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	prSvc.AssertCalled(t, "Reassign", mock.Anything, "pr1", "u2")
}

func TestPRHandlers_Reassign_DomainErrors(t *testing.T) {
	cases := []struct {
		name     string
		err      *domain.DomainError
		expected int
	}{
		{"notfound", domain.NewDomainError(domain.ErrorCodeNotFound, "missing"), http.StatusNotFound},
		{"merged", domain.NewDomainError(domain.ErrorCodePRMerged, "merged"), http.StatusConflict},
		{"notassigned", domain.NewDomainError(domain.ErrorCodeNotAssigned, "no"), http.StatusConflict},
		{"nocandidate", domain.NewDomainError(domain.ErrorCodeNoCandidate, "none"), http.StatusConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prSvc := serviceMocks.NewMockPullRequestService(t)
			prSvc.On("Reassign", mock.Anything, "pr1", "u2").Return(nil, "", tc.err)
			router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), prSvc, &stubHTTPMetrics{})
			body, _ := json.Marshal(map[string]any{"pull_request_id": "pr1", "old_user_id": "u2"})
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			if rr.Code != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, rr.Code)
			}
		})
	}
}
