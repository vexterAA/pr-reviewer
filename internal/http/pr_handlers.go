package http

import (
	"encoding/json"
	"net/http"
	"time"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/service"
)

type prHandlers struct {
	prs service.PullRequestService
}

func newPRHandlers(prs service.PullRequestService) *prHandlers {
	return &prHandlers{prs: prs}
}

type createPRRequest struct {
	ID     string `json:"pull_request_id"`
	Name   string `json:"pull_request_name"`
	Author string `json:"author_id"`
}

type mergePRRequest struct {
	ID string `json:"pull_request_id"`
}

type reassignPRRequest struct {
	ID          string `json:"pull_request_id"`
	OldReviewer string `json:"old_user_id"`
}

type pullRequestDTO struct {
	ID                string                   `json:"pull_request_id"`
	Name              string                   `json:"pull_request_name"`
	AuthorID          string                   `json:"author_id"`
	Status            domain.PullRequestStatus `json:"status"`
	AssignedReviewers []string                 `json:"assigned_reviewers"`
	CreatedAt         *time.Time               `json:"createdAt,omitempty"`
	MergedAt          *time.Time               `json:"mergedAt,omitempty"`
}

type createPRResponse struct {
	PR pullRequestDTO `json:"pr"`
}

type mergePRResponse struct {
	PR pullRequestDTO `json:"pr"`
}

type reassignPRResponse struct {
	PR         pullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

func (h *prHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req createPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.ID == "" || req.Name == "" || req.Author == "" {
		writeBadRequest(w, "pull_request_id, pull_request_name and author_id are required")
		return
	}

	pr, err := h.prs.Create(r.Context(), domain.PullRequest{
		ID:       req.ID,
		Name:     req.Name,
		AuthorID: req.Author,
	})
	if err != nil {
		WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, createPRResponse{
		PR: toPullRequestDTO(*pr),
	})
}

func (h *prHandlers) Merge(w http.ResponseWriter, r *http.Request) {
	var req mergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.ID == "" {
		writeBadRequest(w, "pull_request_id is required")
		return
	}

	pr, err := h.prs.Merge(r.Context(), req.ID)
	if err != nil {
		WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, mergePRResponse{
		PR: toPullRequestDTO(*pr),
	})
}

func (h *prHandlers) Reassign(w http.ResponseWriter, r *http.Request) {
	var req reassignPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.ID == "" || req.OldReviewer == "" {
		writeBadRequest(w, "pull_request_id and old_user_id are required")
		return
	}

	pr, replacedBy, err := h.prs.Reassign(r.Context(), req.ID, req.OldReviewer)
	if err != nil {
		WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, reassignPRResponse{
		PR:         toPullRequestDTO(*pr),
		ReplacedBy: replacedBy,
	})
}

func toPullRequestDTO(pr domain.PullRequest) pullRequestDTO {
	dto := pullRequestDTO{
		ID:                pr.ID,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            pr.Status,
		AssignedReviewers: pr.AssignedReviewers,
	}
	if !pr.CreatedAt.IsZero() {
		dto.CreatedAt = &pr.CreatedAt
	}
	if pr.MergedAt != nil {
		dto.MergedAt = pr.MergedAt
	}
	return dto
}
