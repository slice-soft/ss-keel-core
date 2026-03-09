package contracts

import (
	"context"
	"time"
)

// RequestMetrics holds the data recorded for each HTTP request.
type RequestMetrics struct {
	Method     string
	Path       string
	StatusCode int
	Duration   time.Duration
}

// MetricsCollector is the contract for metrics backends
// (e.g. ss-keel-metrics / Prometheus).
type MetricsCollector interface {
	RecordRequest(m RequestMetrics)
}

// Span represents a single unit of work in a distributed trace.
type Span interface {
	SetAttribute(key string, value any)
	RecordError(err error)
	End()
}

// Tracer creates spans for distributed tracing
// (e.g. ss-keel-tracing / OpenTelemetry).
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}
