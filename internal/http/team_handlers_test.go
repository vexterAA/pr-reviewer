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

func TestTeamHandlers_Add_BadJSON(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestTeamHandlers_Add_MissingName(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{
		"members": []map[string]any{},
	})
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestTeamHandlers_Add_Success(t *testing.T) {
	teamSvc := serviceMocks.NewMockTeamService(t)
	teamSvc.On("AddTeam", mock.Anything, mock.AnythingOfType("domain.Team")).Return(&domain.Team{
		Name:    "backend",
		Members: []domain.User{{ID: "u1", Username: "Alice"}},
	}, nil)

	router := NewRouter(teamSvc, serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	teamSvc.AssertCalled(t, "AddTeam", mock.Anything, mock.MatchedBy(func(team domain.Team) bool {
		return team.Name == "backend"
	}))
}

func TestTeamHandlers_Add_DomainError(t *testing.T) {
	teamSvc := serviceMocks.NewMockTeamService(t)
	teamSvc.On("AddTeam", mock.Anything, mock.Anything).Return(nil, domain.NewDomainError(domain.ErrorCodeTeamExists, "exists"))

	router := NewRouter(teamSvc, serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	body, _ := json.Marshal(map[string]any{
		"team_name": "backend",
	})
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code == http.StatusInternalServerError {
		t.Fatalf("expected mapped error, got 500")
	}
}

func TestTeamHandlers_Get_BadRequest(t *testing.T) {
	router := NewRouter(serviceMocks.NewMockTeamService(t), serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestTeamHandlers_Get_Success(t *testing.T) {
	teamSvc := serviceMocks.NewMockTeamService(t)
	teamSvc.On("GetTeam", mock.Anything, "backend").Return(&domain.Team{Name: "backend"}, nil)
	router := NewRouter(teamSvc, serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestTeamHandlers_Get_NotFound(t *testing.T) {
	teamSvc := serviceMocks.NewMockTeamService(t)
	teamSvc.On("GetTeam", mock.Anything, "backend").Return(nil, domain.NewDomainError(domain.ErrorCodeNotFound, "missing"))
	router := NewRouter(teamSvc, serviceMocks.NewMockUserService(t), serviceMocks.NewMockPullRequestService(t), &stubHTTPMetrics{})

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
