package core

import "context"

// Span represents a single unit of work in a distributed trace.
type Span interface {
	SetAttribute(key string, value any)
	RecordError(err error)
	End()
}

// Tracer creates spans for distributed tracing (e.g. ss-keel-tracing / OpenTelemetry).
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

// noopTracer is the default tracer — performs no operations.
type noopTracer struct{}

func (noopTracer) Start(ctx context.Context, _ string) (context.Context, Span) {
	return ctx, noopSpan{}
}

// noopSpan is a span that does nothing.
type noopSpan struct{}

func (noopSpan) SetAttribute(_ string, _ any) {}
func (noopSpan) RecordError(_ error)           {}
func (noopSpan) End()                          {}
