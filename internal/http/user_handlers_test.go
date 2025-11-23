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

func TestUserHandlers_SetIsActive_BadJSON(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUserHandlers_SetIsActive_MissingUser(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})
	body, _ := json.Marshal(map[string]any{"is_active": true})
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUserHandlers_SetIsActive_Success(t *testing.T) {
	userSvc := serviceMocks.NewMockUserService(t)
	userSvc.On("SetActive", mock.Anything, "u1", true).Return(&domain.User{ID: "u1", Username: "Alice", TeamName: "backend", IsActive: true}, nil)
	router := NewRouter(serviceMocks.NewMockTeamService(t), userSvc, serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{"user_id": "u1", "is_active": true})
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	userSvc.AssertCalled(t, "SetActive", mock.Anything, "u1", true)
}

func TestUserHandlers_SetIsActive_NotFound(t *testing.T) {
	userSvc := serviceMocks.NewMockUserService(t)
	userSvc.On("SetActive", mock.Anything, "u1", true).Return(nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found"))
	router := NewRouter(serviceMocks.NewMockTeamService(t), userSvc, serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{"user_id": "u1", "is_active": true})
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestUserHandlers_GetReview_BadRequest(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})
	req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUserHandlers_GetReview_Success(t *testing.T) {
	userSvc := serviceMocks.NewMockUserService(t)
	userSvc.On("GetReviewPullRequests", mock.Anything, "u1").Return([]domain.PullRequestShort{{ID: "pr1"}, {ID: "pr2"}}, nil)
	router := NewRouter(serviceMocks.NewMockTeamService(t), userSvc, serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	userSvc.AssertCalled(t, "GetReviewPullRequests", mock.Anything, "u1")
}

func TestUserHandlers_GetReview_NotFound(t *testing.T) {
	userSvc := serviceMocks.NewMockUserService(t)
	userSvc.On("GetReviewPullRequests", mock.Anything, "u1").Return(nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found"))
	router := NewRouter(serviceMocks.NewMockTeamService(t), userSvc, serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
