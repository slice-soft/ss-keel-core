package core

import "context"

// noopTracer is the default tracer — performs no operations.
type noopTracer struct{}

func (noopTracer) Start(ctx context.Context, _ string) (context.Context, Span) {
	return ctx, noopSpan{}
}

// noopSpan is a span that does nothing.
type noopSpan struct{}

func (noopSpan) SetAttribute(_ string, _ any) {}
func (noopSpan) RecordError(_ error)          {}
func (noopSpan) End()                         {}
