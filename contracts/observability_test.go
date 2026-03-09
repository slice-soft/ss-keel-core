package contracts

import (
	"context"
	"testing"
	"time"
)

type metricsMock struct {
	last RequestMetrics
}

func (m *metricsMock) RecordRequest(r RequestMetrics) { m.last = r }

type spanMock struct{}

func (spanMock) SetAttribute(_ string, _ any) {}
func (spanMock) RecordError(_ error)          {}
func (spanMock) End()                         {}

type tracerMock struct{}

func (tracerMock) Start(ctx context.Context, _ string) (context.Context, Span) {
	return ctx, spanMock{}
}

type translatorMock struct{}

func (translatorMock) T(_, key string, _ ...any) string { return key + ".translated" }
func (translatorMock) Locales() []string                { return []string{"en", "es"} }

type loggerMock struct {
	entries []string
}

func (l *loggerMock) Info(format string, _ ...interface{}) {
	l.entries = append(l.entries, "info:"+format)
}
func (l *loggerMock) Warn(format string, _ ...interface{}) {
	l.entries = append(l.entries, "warn:"+format)
}
func (l *loggerMock) Error(format string, _ ...interface{}) {
	l.entries = append(l.entries, "error:"+format)
}
func (l *loggerMock) Debug(format string, _ ...interface{}) {
	l.entries = append(l.entries, "debug:"+format)
}

var (
	_ MetricsCollector = (*metricsMock)(nil)
	_ Tracer           = tracerMock{}
	_ Span             = spanMock{}
	_ Translator       = translatorMock{}
	_ Logger           = (*loggerMock)(nil)
)

func TestObservabilityDataStructures(t *testing.T) {
	rm := RequestMetrics{Method: "GET", Path: "/health", StatusCode: 200, Duration: time.Millisecond}
	if rm.Method != "GET" || rm.StatusCode != 200 {
		t.Fatalf("unexpected RequestMetrics value: %+v", rm)
	}
}

func TestObservabilityContractsAreCallable(t *testing.T) {
	ctx := context.Background()

	mm := &metricsMock{}
	mm.RecordRequest(RequestMetrics{Method: "POST"})
	if mm.last.Method != "POST" {
		t.Fatalf("metrics collector did not record request: %+v", mm.last)
	}

	_, span := (tracerMock{}).Start(ctx, "op")
	span.SetAttribute("k", "v")
	span.RecordError(nil)
	span.End()

	translated := (translatorMock{}).T("en", "key")
	if translated != "key.translated" {
		t.Fatalf("translator output = %q, want %q", translated, "key.translated")
	}

	log := &loggerMock{}
	log.Info("hello")
	log.Warn("warn")
	log.Debug("debug")
	if len(log.entries) != 3 {
		t.Fatalf("unexpected logger entries: %+v", log.entries)
	}
}
