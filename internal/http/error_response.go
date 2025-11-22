package http

import (
	"encoding/json"
	"net/http"

	"pr-reviewer/internal/domain"
)

type errorPayload struct {
	Code    domain.ErrorCode `json:"code"`
	Message string           `json:"message"`
}

type ErrorResponse struct {
	Error errorPayload `json:"error"`
}

func WriteError(w http.ResponseWriter, err error) {
	if domainErr, ok := domain.AsDomainError(err); ok {
		writeJSON(w, statusForDomainCode(domainErr.Code), ErrorResponse{
			Error: errorPayload{
				Code:    domainErr.Code,
				Message: domainErr.Message,
			},
		})
		return
	}

	writeJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error: errorPayload{
			Message: "internal error",
		},
	})
}

func statusForDomainCode(code domain.ErrorCode) int {
	switch code {
	case domain.ErrorCodeTeamExists:
		return http.StatusBadRequest
	case domain.ErrorCodePRExists, domain.ErrorCodePRMerged, domain.ErrorCodeNotAssigned, domain.ErrorCodeNoCandidate:
		return http.StatusConflict
	case domain.ErrorCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
