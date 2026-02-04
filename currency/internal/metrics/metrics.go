package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// HTTPRequestsTotal counts total HTTP requests by method, path, and status code.
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration observes HTTP request durations by method and path.
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// WorkerFetchTotal counts worker fetch attempts by status.
	WorkerFetchTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_fetch_total",
			Help: "Total number of worker fetch operations.",
		},
		[]string{"status"},
	)

	// WorkerFetchDuration observes worker fetch durations.
	WorkerFetchDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "worker_fetch_duration_seconds",
			Help:    "Duration of worker fetch operations in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)

	// WorkerBackfillTotal counts backfill operations per date by status.
	WorkerBackfillTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_backfill_dates_total",
			Help: "Total number of backfill date operations.",
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		WorkerFetchTotal,
		WorkerFetchDuration,
		WorkerBackfillTotal,
	)
}
