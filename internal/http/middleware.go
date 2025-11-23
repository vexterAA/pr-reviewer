package http

import (
	"net/http"
	"time"

	"pr-reviewer/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func withHTTPMetrics(next http.Handler, m metrics.HTTPMetrics) http.Handler {
	if m == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rec, r)

		m.ObserveRequest(r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}
