package domain

import (
	"errors"
	"time"
)

type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)

type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

type Team struct {
	Name    string
	Members []User
}

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PullRequestStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

type PullRequestShort struct {
	ID       string
	Name     string
	AuthorID string
	Status   PullRequestStatus
}

type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged    ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
)

type DomainError struct {
	Code    ErrorCode
	Message string
}

func (e *DomainError) Error() string { return e.Message }

func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

func AsDomainError(err error) (*DomainError, bool) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}
