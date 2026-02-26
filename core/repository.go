package core

import "context"

// Repository is the generic CRUD contract implemented by database modules
// such as ss-keel-gorm and ss-keel-mongo.
type Repository[T any, ID any] interface {
	FindByID(ctx context.Context, id ID) (*T, error)
	FindAll(ctx context.Context, q PageQuery) (Page[T], error)
	Create(ctx context.Context, entity *T) error
	Update(ctx context.Context, id ID, entity *T) error
	Delete(ctx context.Context, id ID) error
}
