package core

import "context"

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
