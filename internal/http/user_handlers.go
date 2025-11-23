package http

import (
	"encoding/json"
	"net/http"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/service"
)

type userHandlers struct {
	users service.UserService
}

func newUserHandlers(users service.UserService) *userHandlers {
	return &userHandlers{users: users}
}

type setActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type userDTO struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type setActiveResponse struct {
	User userDTO `json:"user"`
}

func (h *userHandlers) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req setActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.UserID == "" {
		writeBadRequest(w, "user_id is required")
		return
	}

	user, err := h.users.SetActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, setActiveResponse{
		User: toUserDTO(*user),
	})
}

type getReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []pullRequestShortDTO `json:"pull_requests"`
}

type pullRequestShortDTO struct {
	ID       string                   `json:"pull_request_id"`
	Name     string                   `json:"pull_request_name"`
	AuthorID string                   `json:"author_id"`
	Status   domain.PullRequestStatus `json:"status"`
}

func (h *userHandlers) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeBadRequest(w, "user_id is required")
		return
	}

	prs, err := h.users.GetReviewPullRequests(r.Context(), userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	resp := getReviewResponse{
		UserID: userID,
	}
	for _, pr := range prs {
		resp.PullRequests = append(resp.PullRequests, pullRequestShortDTO{
			ID:       pr.ID,
			Name:     pr.Name,
			AuthorID: pr.AuthorID,
			Status:   pr.Status,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func toUserDTO(u domain.User) userDTO {
	return userDTO{
		ID:       u.ID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}
