package core

import "time"

// RequestMetrics holds the data recorded for each HTTP request.
type RequestMetrics struct {
	Method     string
	Path       string
	StatusCode int
	Duration   time.Duration
}

// MetricsCollector is the contract for metrics backends (e.g. ss-keel-metrics / Prometheus).
type MetricsCollector interface {
	RecordRequest(m RequestMetrics)
}
