package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pr-reviewer/internal/metrics"
)

type stubHTTPMetrics struct {
	called bool
	method string
	path   string
	status int
	dur    time.Duration
}

func (s *stubHTTPMetrics) ObserveRequest(method, path string, status int, dur time.Duration) {
	s.called = true
	s.method = method
	s.path = path
	s.status = status
	s.dur = dur
}

func TestMiddlewareObservesRequest(t *testing.T) {
	metricsStub := &stubHTTPMetrics{}
	handler := withHTTPMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}), metricsStub)

	req := httptest.NewRequest(http.MethodPost, "/team/get?foo=bar", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !metricsStub.called {
		t.Fatalf("expected metrics to be observed")
	}
	if metricsStub.method != http.MethodPost {
		t.Fatalf("expected method %s, got %s", http.MethodPost, metricsStub.method)
	}
	if metricsStub.path != "/team/get" {
		t.Fatalf("expected path label /team/get, got %s", metricsStub.path)
	}
	if metricsStub.status != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, metricsStub.status)
	}
}

var _ metrics.HTTPMetrics = (*stubHTTPMetrics)(nil)
