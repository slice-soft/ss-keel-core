package contracts

import (
	"context"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Guard is the contract for authentication/authorization middleware providers
// (e.g. ss-keel-jwt, ss-keel-oauth).
//
// Usage:
//
//	route.Use(jwtGuard.Middleware()).WithSecured("bearerAuth")
type Guard interface {
	Middleware() fiber.Handler
}

// Cache is the contract for key-value caching backends (e.g. Redis).
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// StorageObject holds metadata about an object in storage.
type StorageObject struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

// Storage is the contract for object storage backends (e.g. S3, GCS, local).
type Storage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Stat(ctx context.Context, key string) (*StorageObject, error)
}

// MailAttachment represents a file attached to an email.
type MailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Mail holds all data needed to send an email.
type Mail struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	HTMLBody    string
	TextBody    string
	Attachments []MailAttachment
}

// Mailer is the contract for email sending backends (e.g. ss-keel-mail).
type Mailer interface {
	Send(ctx context.Context, mail Mail) error
}

// Message is the envelope passed through messaging brokers.
type Message struct {
	Topic   string
	Key     []byte
	Payload []byte
	Headers map[string]string
}

// MessageHandler is the function signature for consuming messages.
type MessageHandler func(ctx context.Context, msg Message) error

// Publisher is the contract for publishing messages (e.g. ss-keel-amqp, ss-keel-kafka).
type Publisher interface {
	Publish(ctx context.Context, msg Message) error
	Close() error
}

// Subscriber is the contract for consuming messages from a topic.
type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Close() error
}

// Job represents a scheduled task.
type Job struct {
	Name     string
	Schedule string // cron expression, e.g. "*/5 * * * *"
	Handler  func(ctx context.Context) error
}

// Scheduler is the contract for cron-like task scheduling (e.g. ss-keel-cron).
type Scheduler interface {
	Add(job Job) error
	Start()
	Stop(ctx context.Context)
}

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

// Translator is the contract for i18n providers (e.g. ss-keel-i18n).
type Translator interface {
	T(locale, key string, args ...any) string
	Locales() []string
}
