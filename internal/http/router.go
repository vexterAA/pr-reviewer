package http

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pr-reviewer/internal/metrics"
	"pr-reviewer/internal/service"
)

func NewRouter(teamSvc service.TeamService, userSvc service.UserService, prSvc service.PullRequestService, httpMetrics metrics.HTTPMetrics) http.Handler {
	mux := http.NewServeMux()

	teamHandlers := newTeamHandlers(teamSvc)
	userHandlers := newUserHandlers(userSvc)
	prHandlers := newPRHandlers(prSvc)

	mux.HandleFunc("/team/add", method("POST", teamHandlers.Add))
	mux.HandleFunc("/team/get", method("GET", teamHandlers.Get))

	mux.HandleFunc("/users/setIsActive", method("POST", userHandlers.SetIsActive))
	mux.HandleFunc("/users/getReview", method("GET", userHandlers.GetReview))

	mux.HandleFunc("/pullRequest/create", method("POST", prHandlers.Create))
	mux.HandleFunc("/pullRequest/merge", method("POST", prHandlers.Merge))
	mux.HandleFunc("/pullRequest/reassign", method("POST", prHandlers.Reassign))

	metricsHandler := promhttp.Handler()
	wrapped := withHTTPMetrics(mux, httpMetrics)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			metricsHandler.ServeHTTP(w, r)
			return
		}
		wrapped.ServeHTTP(w, r)
	})
}

func method(expected string, h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expected {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h(w, r)
	}
}
