package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPMetrics interface {
	ObserveRequest(method, path string, status int, dur time.Duration)
}

type BusinessMetrics interface {
	IncPRCreated()
	IncPRMerged()
	IncPRReassign(result string)
}

type httpMetrics struct {
	requests  *prometheus.CounterVec
	latencies *prometheus.HistogramVec
}

type businessMetrics struct {
	prEvents   *prometheus.CounterVec
	prReassign *prometheus.CounterVec
}

var (
	initOnce     sync.Once
	httpRequests *prometheus.CounterVec
	httpDuration *prometheus.HistogramVec
	prEvents     *prometheus.CounterVec
	prReassign   *prometheus.CounterVec
)

func New() (HTTPMetrics, BusinessMetrics) {
	initOnce.Do(func() {
		httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests.",
		}, []string{"method", "path", "status"})
		httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path", "status"})
		prEvents = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "pr_events_total",
			Help: "Pull request events.",
		}, []string{"event"})
		prReassign = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "pr_reassign_total",
			Help: "Pull request reassign results.",
		}, []string{"result"})

		prometheus.MustRegister(httpRequests, httpDuration, prEvents, prReassign)
	})

	return &httpMetrics{
			requests:  httpRequests,
			latencies: httpDuration,
		}, &businessMetrics{
			prEvents:   prEvents,
			prReassign: prReassign,
		}
}

func (m *httpMetrics) ObserveRequest(method, path string, status int, dur time.Duration) {
	labels := prometheus.Labels{
		"method": method,
		"path":   path,
		"status": formatStatus(status),
	}
	m.requests.With(labels).Inc()
	m.latencies.With(labels).Observe(dur.Seconds())
}

func (m *businessMetrics) IncPRCreated() {
	m.prEvents.WithLabelValues("created").Inc()
}

func (m *businessMetrics) IncPRMerged() {
	m.prEvents.WithLabelValues("merged").Inc()
}

func (m *businessMetrics) IncPRReassign(result string) {
	m.prReassign.WithLabelValues(result).Inc()
}

func formatStatus(status int) string {
	return strconv.Itoa(status)
}
