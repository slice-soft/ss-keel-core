package contracts

import "context"

// HealthChecker is the contract for external health check contributors.
// Implementations report the status of dependencies such as a DB or cache.
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) error
}
