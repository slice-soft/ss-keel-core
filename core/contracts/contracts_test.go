package contracts

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

type guardMock struct{}

func (guardMock) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error { return c.Next() }
}

type cacheMock struct{}

func (cacheMock) Get(_ context.Context, _ string) ([]byte, error)                  { return []byte("v"), nil }
func (cacheMock) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error { return nil }
func (cacheMock) Delete(_ context.Context, _ string) error                         { return nil }
func (cacheMock) Exists(_ context.Context, _ string) (bool, error)                 { return true, nil }

type storageMock struct{}

func (storageMock) Put(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error { return nil }
func (storageMock) Get(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("ok")), nil
}
func (storageMock) Delete(_ context.Context, _ string) error { return nil }
func (storageMock) URL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return "https://example.com/file", nil
}
func (storageMock) Stat(_ context.Context, _ string) (*StorageObject, error) {
	return &StorageObject{Key: "file", Size: 10}, nil
}

type mailerMock struct{}

func (mailerMock) Send(_ context.Context, _ Mail) error { return nil }

type publisherMock struct{}

func (publisherMock) Publish(_ context.Context, _ Message) error { return nil }
func (publisherMock) Close() error                               { return nil }

type subscriberMock struct{}

func (subscriberMock) Subscribe(_ context.Context, _ string, _ MessageHandler) error { return nil }
func (subscriberMock) Close() error                                                  { return nil }

type schedulerMock struct {
	started bool
	stopped bool
}

func (s *schedulerMock) Add(_ Job) error        { return nil }
func (s *schedulerMock) Start()                 { s.started = true }
func (s *schedulerMock) Stop(_ context.Context) { s.stopped = true }

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

var (
	_ Guard            = guardMock{}
	_ Cache            = cacheMock{}
	_ Storage          = storageMock{}
	_ Mailer           = mailerMock{}
	_ Publisher        = publisherMock{}
	_ Subscriber       = subscriberMock{}
	_ Scheduler        = (*schedulerMock)(nil)
	_ MetricsCollector = (*metricsMock)(nil)
	_ Tracer           = tracerMock{}
	_ Span             = spanMock{}
	_ Translator       = translatorMock{}
)

func TestContractDataStructures(t *testing.T) {
	att := MailAttachment{Filename: "f.txt", ContentType: "text/plain", Data: []byte("x")}
	mail := Mail{
		From:        "a@example.com",
		To:          []string{"b@example.com"},
		Subject:     "hello",
		TextBody:    "body",
		Attachments: []MailAttachment{att},
	}
	if mail.Subject != "hello" || len(mail.Attachments) != 1 {
		t.Fatalf("unexpected Mail value: %+v", mail)
	}

	msg := Message{Topic: "users.created", Key: []byte("1"), Payload: []byte(`{"id":"1"}`)}
	if msg.Topic == "" || len(msg.Payload) == 0 {
		t.Fatalf("unexpected Message value: %+v", msg)
	}

	obj := StorageObject{Key: "avatars/u1.png", Size: 128}
	if obj.Key == "" || obj.Size != 128 {
		t.Fatalf("unexpected StorageObject value: %+v", obj)
	}

	job := Job{Name: "cleanup", Schedule: "0 * * * *", Handler: func(context.Context) error { return nil }}
	if job.Name == "" || job.Handler == nil {
		t.Fatalf("unexpected Job value: %+v", job)
	}

	rm := RequestMetrics{Method: "GET", Path: "/health", StatusCode: 200, Duration: time.Millisecond}
	if rm.Method != "GET" || rm.StatusCode != 200 {
		t.Fatalf("unexpected RequestMetrics value: %+v", rm)
	}
}

func TestContractMocksAreCallable(t *testing.T) {
	ctx := context.Background()

	if _, err := (storageMock{}).Get(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	if _, err := (storageMock{}).Stat(ctx, "k"); err != nil {
		t.Fatal(err)
	}

	sm := &schedulerMock{}
	sm.Start()
	sm.Stop(ctx)
	if !sm.started || !sm.stopped {
		t.Fatalf("scheduler flags = started:%v stopped:%v", sm.started, sm.stopped)
	}

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
}
