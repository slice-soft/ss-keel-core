package core

import (
	"context"
	"io"
	"time"
)

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
