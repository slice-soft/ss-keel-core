package contracts

import "context"

// Repository is the generic CRUD contract implemented by database modules
// such as ss-keel-gorm and ss-keel-mongo.
// Q is the query/pagination type and P is the paginated result type.
type Repository[T any, ID any, Q any, P any] interface {
	FindByID(ctx context.Context, id ID) (*T, error)
	FindAll(ctx context.Context, q Q) (P, error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, id ID, entity *T) error
	Patch(ctx context.Context, id ID, patch *T) error
	Delete(ctx context.Context, id ID) error
}
