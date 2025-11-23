package http

import (
	"net/http"

	"pr-reviewer/internal/service"
)

func NewRouter(_ service.TeamService, _ service.UserService, _ service.PullRequestService) http.Handler {
	mux := http.NewServeMux()
	// TODO: register handlers.
	return mux
}
